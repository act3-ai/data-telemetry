package types

import (
	"time"

	"github.com/opencontainers/go-digest"
)

// LocationResponse is a struct for the location.
type LocationResponse struct {
	Repository   string
	AuthRequired bool
	Digest       digest.Digest
}

// ListResultEntry is a single entry in list request.
type ListResultEntry struct {
	CreatedAt time.Time
	Digests   []digest.Digest
	Data      []byte
}

// SearchResult is a result from a BottleSearch Request.
type SearchResult struct {
	Digests []digest.Digest
	Data    []byte
}

// TopologicalOrderingOfTypes is the list of different input types in the order they need to be process/applied.
var TopologicalOrderingOfTypes = []string{"blob", "bottle", "manifest", "event", "signature"}
