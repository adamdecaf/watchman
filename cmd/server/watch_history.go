// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"time"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/watchman/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addWatchRoutes(logger log.Logger, r *mux.Router, repo watchRepository) {
	r.Methods("GET").Path("/ofac/sdn/watches/{watchID}").HandlerFunc(getOFACWatchHistory(logger, repo))
}

func getOFACWatchHistory(logger log.Logger, repo watchRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		watchID := getWatchID(w, r)
		if watchID == "" {
			return
		}

		results, err := repo.getOFACWatchHistory(watchID, extractSearchLimit(r))
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	}
}

func (r *sqliteWatchRepository) getOFACWatchHistory(watchID string, limit int) ([]client.OfacWatchHistory, error) {
	return []client.OfacWatchHistory{
		{
			WatchID: watchID,
			Type:    "customer_name",
			Results: []client.OfacSearch{
				{
					EntityID:  "123",
					SdnName:   "nicolas maduro",
					SdnType:   "individual",
					Match:     0.88234,
					CreatedAt: time.Now(),
				},
			},
		},
	}, nil
}
