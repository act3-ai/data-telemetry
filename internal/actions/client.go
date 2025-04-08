package actions

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"oras.land/oras-go/v2/registry/remote/credentials"

	bottle "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io"
	"github.com/act3-ai/go-common/pkg/config"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/redact"

	"github.com/act3-ai/data-telemetry/v3/internal/api"
	"github.com/act3-ai/data-telemetry/v3/internal/db"
	"github.com/act3-ai/data-telemetry/v3/internal/middleware"
	"github.com/act3-ai/data-telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"github.com/act3-ai/data-telemetry/v3/pkg/oauth2/device"
)

// ClientConfigOverride is a function used to override the client configuration.
type ClientConfigOverride func(ctx context.Context, c *v1alpha2.ClientConfiguration) error

// Client is the action group for all client commands.
type Client struct {
	*Telemetry

	ConfigFiles []string

	// Handles overrides for configuration
	configOverrides []ClientConfigOverride
}

// NewHandler constructs a new http handler (DEPRECATED)
// TODO convert this to not return a http.Handler but instead start a local HTTP server on port 0 (any port).
func (action *Client) NewHandler(ctx context.Context) (http.Handler, error) {
	// retrieve the server configuration
	serverConfig, err := action.GetServerConfig(ctx)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	if err := bottle.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("adding bottle to scheme: %w", err)
	}

	// connect directly to the DB
	myDB, err := db.Open(ctx, serverConfig.DB, scheme)
	if err != nil {
		return nil, err
	}

	serveMux := http.NewServeMux()

	myAPI := api.API{}
	myAPI.Initialize(serveMux, scheme)

	return middleware.DatabaseMiddleware(myDB)(serveMux), nil
}

// AddClientConfigOverride adds an override function that will be called in GetConfig to edit config after loading.
func (action *Client) AddClientConfigOverride(override ...ClientConfigOverride) {
	action.configOverrides = append(action.configOverrides, override...)
}

// GetClientConfig returns the client's configuration object.
func (action *Client) GetClientConfig(ctx context.Context) (*v1alpha2.ClientConfiguration, error) {
	log := logger.FromContext(ctx)

	c := &v1alpha2.ClientConfiguration{}
	if err := config.Load(log, action.GetConfigScheme(), c, action.ConfigFiles); err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	// Loop through override functions, applying each to the configuration
	for _, overrideFunction := range action.configOverrides {
		if err := overrideFunction(ctx, c); err != nil {
			return c, err
		}
	}

	log.InfoContext(ctx, "Using config", "configuration", c)

	return c, nil
}

// matchURLConfig will find and return the config file of the url string given and if does not exist create a new config.
func matchURLConfig(urlString string, clientConfig *v1alpha2.ClientConfiguration) (*v1alpha2.Location, error) {
	for _, location := range clientConfig.Locations {
		if location.URL == redact.SecretURL(urlString) {
			return &location, nil
		}
	}
	return &v1alpha2.Location{
		Name: "",
		URL:  redact.SecretURL(urlString),
	}, nil
}

// authClientOrDefault creates an OAuth *http.Client if necessary, defaulting to
// the default http client if unnecessary or problems occur.
func authClientOrDefault(ctx context.Context, loc *v1alpha2.Location) *http.Client {
	log := logger.FromContext(ctx)
	httpClient := http.DefaultClient
	if loc.OAuth.Issuer != "" && loc.OAuth.ClientID != "" {
		// TODO: Errors here likely should be displayed
		issuerURL, err := url.Parse(loc.OAuth.Issuer)
		if err != nil {
			log.ErrorContext(ctx, "parsing host oauth issuer", "issuer", loc.OAuth.Issuer, "clientID", loc.OAuth.ClientID, "error", err) //nolint:sloglint
			goto Recover
		}

		// promptFn implements device.AuthPromtFn.
		promptFn := func(ctx context.Context, uri, userCode string) error {
			_, err := fmt.Fprintf(os.Stderr, "On the device you would like to authenticate, please visit %s?user_code=%s", uri, userCode)
			return err
		}

		var credStore credentials.Store
		credStore, err = credentials.NewStoreFromDocker(credentials.StoreOptions{})
		if err != nil {
			log.ErrorContext(ctx, "accessing docker credential store", "error", err)
			credStore = credentials.NewMemoryStore()
		}

		authClient, err := device.NewOAuthClient(ctx, issuerURL, string(loc.OAuth.ClientID), credStore, promptFn)
		if err != nil {
			log.ErrorContext(ctx, "initializing oauth client", "issuer", loc.OAuth.Issuer, "clientID", loc.OAuth.ClientID, "error", err) //nolint:sloglint
			goto Recover
		}
		httpClient = authClient
	}

Recover:
	return httpClient
}
