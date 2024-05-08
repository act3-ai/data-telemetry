package api

import (
	"time"

	"github.com/opencontainers/go-digest"
)

// EventDto is the /event request body.
type EventDto struct {
	ManifestDigest digest.Digest             `json:"manifestDigest,omitempty"`
	Action         string                    `json:"action,omitempty"`
	Repository     string                    `json:"repository,omitempty"`
	Tag            string                    `json:"tag,omitempty"`
	AuthRequired   bool                      `json:"authRequired,omitempty"`
	Durations      map[digest.Digest]float32 `json:"durations,omitempty"`
	Timestamp      time.Time                 `json:"timestamp,omitempty"`
	Username       string                    `json:"username,omitempty"`
}
