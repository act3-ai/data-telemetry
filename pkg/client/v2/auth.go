package client

import (
	"net/http"
)

// AuthRequestOptsFunc can be passed to all API requests to customize the API request with auth options.
type AuthRequestOptsFunc func(*http.Request) error

// WithBearerTokenAuth takes a token which is then used when making this one request.
func WithBearerTokenAuth(token string) AuthRequestOptsFunc {
	return func(req *http.Request) error {
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return nil
	}
}
