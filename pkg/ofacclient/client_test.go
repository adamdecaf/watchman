// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apoface License
// license that can be found in the LICENSE file.

package ofacclient

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// TestOFAC__getOFACAddress will fail if ever ran inside a Kubernetes cluster.
func TestOFAC__getOFACAddress(t *testing.T) {
	// Local development
	if addr := getOFACAddress(); addr != "http://localhost:8084" {
		t.Error(addr)
	}

	// OFAC_ENDPOINT environment variable
	os.Setenv("OFAC_ENDPOINT", "https://api.moov.io/v1/ofac")
	if addr := getOFACAddress(); addr != "https://api.moov.io/v1/ofac" {
		t.Error(addr)
	}
}

func TestOFAC__pingRoute(t *testing.T) {
	ofacClient, _, server := MockClientServer("pingRoute", AddPingRoute)
	defer server.Close()

	// Make our ping request
	if err := ofacClient.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestOFAC__post(t *testing.T) {
	ofacClient, _, server := MockClientServer("post", func(r *mux.Router) { AddCreateRoute(nil, r) })
	defer server.Close()

	body := strings.NewReader(`{"id": "foo"}`) // partial ofac.File JSON

	resp, err := ofacClient.POST("/files/create", "unique", ioutil.NopCloser(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if v := resp.Header.Get("X-Idempotency-Key"); v != "unique" {
		t.Error(v)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if v := string(bs); !strings.HasPrefix(v, `{"id":`) {
		t.Error(v)
	}
}

func TestOFAC__buildAddress(t *testing.T) {
	ofacClient := &Client{
		endpoint: "http://localhost:8080",
	}
	if v := ofacClient.buildAddress("/ping"); v != "http://localhost:8080/ping" {
		t.Errorf("got %q", v)
	}

	ofacClient.endpoint = "http://localhost:8080/"
	if v := ofacClient.buildAddress("/ping"); v != "http://localhost:8080/ping" {
		t.Errorf("got %q", v)
	}

	ofacClient.endpoint = "https://api.moov.io/v1/ofac"
	if v := ofacClient.buildAddress("/ping"); v != "https://api.moov.io/v1/ofac/ping" {
		t.Errorf("got %q", v)
	}
}

func TestOFAC__addRequestHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	api := New("addRequestHeaders", log.NewNopLogger())
	api.addRequestHeaders("idempotencyKey", "requestId", req)

	if v := req.Header.Get("User-Agent"); !strings.HasPrefix(v, "ofac/") {
		t.Errorf("got %q", v)
	}
	if v := req.Header.Get("X-Request-Id"); v == "" {
		t.Error("empty header value")
	}
}
