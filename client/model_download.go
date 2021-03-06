/*
 * Watchman API
 *
 * Moov Watchman offers download, parse, and search functions over numerous U.S. trade sanction lists for complying with regional laws. Also included is a web UI and async webhook notification service to initiate processes on remote systems.
 *
 * API version: v1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client

import (
	"time"
)

// Download Metadata and stats about downloaded OFAC data
type Download struct {
	SDNs              int32     `json:"SDNs,omitempty"`
	AltNames          int32     `json:"altNames,omitempty"`
	Addresses         int32     `json:"addresses,omitempty"`
	SectoralSanctions int32     `json:"sectoralSanctions,omitempty"`
	DeniedPersons     int32     `json:"deniedPersons,omitempty"`
	BisEntities       int32     `json:"bisEntities,omitempty"`
	Timestamp         time.Time `json:"timestamp,omitempty"`
}
