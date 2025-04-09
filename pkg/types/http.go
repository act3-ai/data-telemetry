package types

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opencontainers/go-digest"

	"github.com/act3-ai/go-common/pkg/httputil"
)

const (
	// HeaderContentDigest is a header used to denote the body's digest.
	HeaderContentDigest = "X-Content-Digest"
)

// MissingDigestsError is used to denote that Blobs are missing.
type MissingDigestsError struct {
	httputil.HTTPError
	MissingDigests []digest.Digest `json:"missingDigests"`
}

func (e *MissingDigestsError) Error() string {
	return fmt.Sprintf("%s digests %v", e.Detail, e.MissingDigests)
}

// ErrorArgs returns extra KV args for logging the error.
func (e *MissingDigestsError) ErrorArgs() []any {
	return e.CauseArgs
}

// ResponseBody returns JSON response body.
func (e *MissingDigestsError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling response body: %w", err)
	}
	return body, nil
}

// NewMissingDigestsError created a new error with missing items referenced by digest.
func NewMissingDigestsError(item string, missing []digest.Digest, extraKV ...any) error {
	return &MissingDigestsError{
		HTTPError:      *httputil.NewHTTPError(nil, http.StatusPreconditionFailed, "Missing "+item, extraKV...),
		MissingDigests: missing,
	}
}

// ensure MissingDigestsError implements ClientError.
var (
	_ error                = &MissingDigestsError{}
	_ httputil.ClientError = &MissingDigestsError{}
)
