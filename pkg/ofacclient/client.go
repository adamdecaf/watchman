// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apoface License
// license that can be found in the LICENSE file.

package ofacclient

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/moov-io/base/http/bind"
	"github.com/moov-io/base/k8s"
	"github.com/moov-io/ofac"

	"github.com/go-kit/kit/log"
)

var (
	// httpClient is an HTTP client that implements retries.
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     1 * time.Minute,
		},
	}
)

// New creates and returns an Client instance which can be used to make HTTP requests
// to an OFAC service.
//
// There is a shared *http.Client used across all instances.
//
// If ran inside a Kubernetes cluster then Moov's kube-dns record will be the default endpoint.
func New(userId string, logger log.Logger) *Client {
	return &Client{
		client:   httpClient,
		endpoint: getOFACAddress(),
		logger:   logger,
		userId:   userId,
	}
}

// getOFACAddress returns a URL pointing to where an OFAC service lives.
// This method handles Kubernetes and local deployments.
func getOFACAddress() string {
	if k8s.Inside() {
		return "http://ofac.apps.svc.cluster.local:8080/"
	}

	// OFAC_ENDPOINT is a DNS record responsible for routing us to an OFAC instance.
	// Example: http://ofac.apps.svc.cluster.local:8080/
	addr := os.Getenv("OFAC_ENDPOINT")
	if addr != "" {
		return addr
	}

	return "http://localhost" + bind.HTTP("ofac") // local dev
}

// Client is an object for interacting with the Moov OFAC service.
//
// This is not intended to be a complete implementation of the API endpoints. Moov offers an OpenAPI specification
// and Go client library that does cover the entire set of API endpoints.
type Client struct {
	client   *http.Client
	endpoint string

	logger log.Logger

	userId string
}

// Ping makes an HTTP GET /ping request to the OFAC service and returns any errors encountered.
func (c *Client) Ping() error {
	resp, err := c.GET("/ping")
	if err != nil {
		return fmt.Errorf("error getting /ping from OFAC service: %v", err)
	}
	defer resp.Body.Close()

	// parse content-length header
	n, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return fmt.Errorf("error parsing OFAC service /ping response: %v", err)
	}
	if n > 0 {
		return nil
	}
	return fmt.Errorf("no /ping response from OFAC")
}

func createRequestId() string {
	bs := make([]byte, 20)
	n, err := rand.Read(bs)
	if err != nil || n == 0 {
		return ""
	}
	return strings.ToLower(hex.EncodeToString(bs))
}

func (c *Client) addRequestHeaders(idempotencyKey, requestId string, r *http.Request) {
	r.Header.Set("User-Agent", fmt.Sprintf("ofac/%s", ofac.Version))
	if idempotencyKey != "" {
		r.Header.Set("X-Idempotency-Key", idempotencyKey)
	}
	if requestId != "" {
		r.Header.Set("X-Request-Id", requestId)
	}
	if c.userId != "" {
		r.Header.Set("X-User-Id", c.userId)
	}
}

// GET performs a HTTP GET request against the c.endpoint and relPath.
// Retries are supported and handled within this method, so if you can't block
// run this method in a goroutine.
func (c *Client) GET(relPath string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.buildAddress(relPath), nil)
	if err != nil {
		return nil, err
	}
	requestId := createRequestId()
	c.addRequestHeaders("", requestId, req)
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("OFAC GET requestId=%s : %v", requestId, err)
	}
	return resp, nil
}

// POST performs a HTTP POST request against c.endpoint and relPath.
// Retries are supported only if idempotencyKey is non-empty, otherwise only one attempt is made.
//
// This method assumes a non-nil body is JSON.
func (c *Client) POST(relPath string, idempotencyKey string, body io.ReadCloser) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.buildAddress(relPath), body)
	if err != nil {
		return nil, err
	}

	requestId := createRequestId()
	c.addRequestHeaders(idempotencyKey, requestId, req)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// We only will make one HTTP attempt if X-Idempotency-Key is empty.
	// This is done because without a key there's no way to prevent retries, so
	// we've added this to prevent bugs.
	if idempotencyKey == "" {
		resp, err := c.client.Do(req)
		if err != nil {
			return resp, fmt.Errorf("OFAC POST requestId=%q : %v", requestId, err)
		}
		return resp, nil
	}

	// Use our retrying client
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("OFAC POST requestId=%q : %v", requestId, err)
	}
	return resp, nil
}

// buildAddress takes c.endpoint's path and joins it with path to use
// as the full URL for an http.Client request.
//
// This is to handle differences in k8s and local dev. (i.e. /v1/ofac/)
func (c *Client) buildAddress(p string) string {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return ""
	}
	if u.Scheme == "" && c.logger != nil {
		c.logger.Log("ofac", fmt.Sprintf("invalid endpoint=%s", u.String()))
		return ""
	}
	u.Path = path.Join(u.Path, p)
	return u.String()
}
