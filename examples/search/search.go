// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/moov-io/base/http/bind"
	"github.com/moov-io/base/k8s"
	ofac "github.com/moov-io/ofac/client"

	// "github.com/antihax/optional"
	"github.com/go-kit/kit/log"
)

func ofacClient(logger log.Logger) *ofac.APIClient {
	conf := ofac.NewConfiguration()
	conf.BasePath = "http://localhost" + bind.HTTP("ofac")
	if k8s.Inside() {
		conf.BasePath = "http://ofac.apps.svc.cluster.local:8080"
	}

	// OFAC_ENDPOINT is a DNS record responsible for routing us to an ACH instance.
	// Example: http://ofac.apps.svc.cluster.local:8080
	if v := os.Getenv("OFAC_ENDPOINT"); v != "" {
		conf.BasePath = v
	}

	logger.Log("ofac", fmt.Sprintf("using %s for OFAC address", conf.BasePath))

	return ofac.NewAPIClient(conf)
}
