// Package app represents the composed REST API and Web application
package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/api"
	mware "gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/webapp"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
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
	Router chi.Router
	DB     *gorm.DB
}

// NewApp create a new Telemetry application.
func NewApp(db *gorm.DB, scheme *runtime.Scheme, webConf v1alpha1.WebApp, log *slog.Logger, version string) (*App, error) {
	if db == nil {
		return nil, errors.New("DB is required")
	}

	r := chi.NewRouter()

	a := &App{r, db}

	// add some middleware
	r.Use(
		// NOTE from a security perspective sharing your server version is considered a security issue by some, but not by me.
		// The troubleshooting value out weights the security concerns.
		httputil.ServerHeaderMiddleware(fmt.Sprintf("telemetry/%s", version)),

		httputil.TracingMiddleware,
		httputil.LoggingMiddleware(log),
		mware.DatabaseMiddleware(db),
		httputil.PrometheusMiddleware,
		// mware.RecovererMiddleware,
	)

	prometheus.DefaultRegisterer.MustRegister(httputil.HTTPDuration)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// TODO this should be exposed on its own port
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Get("/health", httputil.RootHandler(healthHandler).ServeHTTP)
	r.Get("/readiness", httputil.RootHandler(readinessHandler).ServeHTTP)
	r.Get("/version", versionHandler(version).ServeHTTP)

	// Setup the REST API
	r.Route("/api", func(router chi.Router) {
		myAPI := api.API{}
		myAPI.Initialize(router, scheme)
	})

	// Setup the Web App (leaderboard, catalog, ...)
	webApp, err := webapp.NewWebApp(webConf, log, version)
	if err != nil {
		return nil, err
	}
	r.Route("/www", func(router chi.Router) {
		webApp.Initialize(router)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// We do not use this one because it converts it to an absolute path (preventing relocation behind a reverse proxy)
		// http.Redirect(w, r, "www/", http.StatusFound)
		w.Header().Set("Location", "www/")
		w.WriteHeader(http.StatusFound)
	})

	return a, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) error {
	return httputil.WriteJSON(w, map[string]string{"status": "healthy"})
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
	return httputil.WriteJSON(w, map[string]string{"status": "ready"})
}

func versionHandler(version string) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.WriteJSON(w, map[string]string{"version": version})
	})
}
