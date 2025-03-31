// Package app represents the composed REST API and Web application
package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/internal/api"
	mware "gitlab.com/act3-ai/asce/data/telemetry/v3/internal/middleware"
	"gitlab.com/act3-ai/asce/data/telemetry/v3/internal/webapp"
	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil/promhttputil"
)

// DatabaseType is the type of the database being used.
type DatabaseType string

// Database types.
const (
	DatabaseSQLite   DatabaseType = "sqlite"
	DatabasePostgres DatabaseType = "postgres"
)

// App implements the top level application (REST + HTML).
type App struct {
	HTTPHandler http.Handler
	DB          *gorm.DB
}

// NewApp create a new Telemetry application.
func NewApp(db *gorm.DB, scheme *runtime.Scheme, webConf v1alpha2.WebApp, log *slog.Logger, version string) (*App, error) {
	if db == nil {
		return nil, errors.New("DB is required")
	}

	mainMux := http.NewServeMux()

	// add some middleware
	// NOTE from a security perspective sharing your server version is considered a security issue by some, but not by me.
	// The troubleshooting value out weights the security concerns.
	wrappedMainMuxHandler := httputil.ServerHeaderMiddleware(fmt.Sprintf("telemetry/%s", version))(
		httputil.TracingMiddleware(
			httputil.LoggingMiddleware(log)(
				mware.DatabaseMiddleware(db)(
					promhttputil.PrometheusMiddleware(
						// Set a timeout value on the request context (ctx), that will signal
						// through ctx.Done() that the request has timed out and further
						// processing should be stopped.
						httputil.TimeoutMiddleware(mainMux, 60*time.Second))))))
	// mware.RecovererMiddleware,

	a := &App{wrappedMainMuxHandler, db}

	prometheus.DefaultRegisterer.MustRegister(promhttputil.HTTPDuration)

	// TODO this should be exposed on its own port
	mainMux.Handle("GET /metrics", promhttp.Handler())

	mainMux.Handle("GET /health", httputil.RootHandler(healthHandler))
	mainMux.Handle("GET /readiness", httputil.RootHandler(readinessHandler))
	mainMux.Handle("GET /version", versionHandler(version))

	// Setup the REST API
	myAPI := api.API{}
	apiMux := http.NewServeMux()
	mainMux.Handle("/api/", http.StripPrefix("/api", apiMux))
	myAPI.Initialize(apiMux, scheme)

	// Setup the Web App (leaderboard, catalog, ...)
	webApp, err := webapp.NewWebApp(webConf, log, version)
	if err != nil {
		return nil, err
	}
	webMux := http.NewServeMux()
	mainMux.Handle("/www/", http.StripPrefix("/www", webMux))
	webApp.Initialize(webMux)

	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// We do not use this one because it converts it to an absolute path (preventing relocation behind a reverse proxy)
		// http.Redirect(w, r, "www/", http.StatusFound)
		if r.Method == http.MethodGet {
			w.Header().Set("Location", "/www/")
			w.WriteHeader(http.StatusFound)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return a, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) error {
	if err := httputil.WriteJSON(w, map[string]string{"status": "healthy"}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

func readinessHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	con := mware.DatabaseFromContext(ctx)

	db, err := con.DB()
	if err != nil {
		return fmt.Errorf("readiness handler getting DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("readiness handler pinging: %w", err)
	}
	if err := httputil.WriteJSON(w, map[string]string{"status": "ready"}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

func versionHandler(version string) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.WriteJSON(w, map[string]string{"version": version})
	})
}
