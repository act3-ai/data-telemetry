package actions

import (
	"context"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/app"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
)

// Serve is the action for starting the server.
type Serve struct {
	*Telemetry

	Listen string
}

// Run is the action method.
func (action *Serve) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	serverConfig, err := action.GetServerConfig(ctx)
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	if err := bottle.AddToScheme(scheme); err != nil {
		return err
	}

	myDB, err := db.Open(ctx, serverConfig.DB, scheme)
	if err != nil {
		return err
	}

	myApp, err := app.NewApp(myDB, scheme, serverConfig.WebApp, log, action.GetVersionInfo().Version)
	if err != nil {
		return err
	}

	// graceful shutdown adapted from https://github.com/gorilla/mux#graceful-shutdown

	srv := &http.Server{
		Addr: action.Listen,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      myApp.Router,
	}

	return httputil.Serve(ctx, srv, 10*time.Second)
}
