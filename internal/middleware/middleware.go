// Package middleware contains specific middleware for Telemetry's server
package middleware

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"
)

type middlewareFunc = func(http.Handler) http.Handler

// dbInstanceKey is how we find the DB in a context.Context.
type dbInstanceKey struct{}

// DatabaseFromContext returns the database instance customized for this request.
func DatabaseFromContext(ctx context.Context) *gorm.DB {
	if v := ctx.Value(dbInstanceKey{}); v != nil {
		return v.(*gorm.DB)
	}
	// panic("database missing from context")
	return nil
}

// DatabaseMiddleware returns a middleware that injects the db into the context.  This depends on the LoggingMiddleware so this must be applied after the LoggingMiddleware.
func DatabaseMiddleware(con *gorm.DB) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx).WithGroup("gorm")
			tx := con.Session(&gorm.Session{
				Context: logger.NewContext(ctx, log),
			})
			ctx = context.WithValue(ctx, dbInstanceKey{}, tx)
			// Call the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SignatureVerifyMiddleware records timing metrics.
func SignatureVerifyMiddleware(next http.Handler) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusInternalServerError, "Unable to read body")
		}
		// must close
		if err := r.Body.Close(); err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Grab the signature
		signatureBase64 := r.Header.Get("x-signature")
		signature, err := base64.StdEncoding.DecodeString(signatureBase64)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusInternalServerError, "Incorrectly formatted x-signature header")
		}

		// Grab the public key
		pubkeyBase64 := r.Header.Get("x-publickey")
		pubkey, err := base64.StdEncoding.DecodeString(pubkeyBase64)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusInternalServerError, "Incorrectly formatted x-publickey header")
		}

		// Do the actual verification
		if !ed25519.Verify(pubkey, bodyBytes, signature) {
			return httputil.NewHTTPError(nil, http.StatusBadRequest, "Unable to verify signature")
		}

		// call the next handler because everything checked out
		next.ServeHTTP(w, r)
		return nil
	})
}

var _ middlewareFunc = SignatureVerifyMiddleware
