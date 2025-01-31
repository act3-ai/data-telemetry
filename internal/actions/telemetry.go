package actions

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/version"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// ServerConfigOverride is a function used to override the server configuration.
type ServerConfigOverride func(ctx context.Context, c *v1alpha2.ServerConfiguration) error

// Telemetry action for all commands.
type Telemetry struct {
	versionInfo  version.Info
	configScheme *runtime.Scheme

	ConfigFiles []string

	// Handles overrides for configuration
	configOverrides []ServerConfigOverride
}

// NewTelemetry creates a new telemetry action
// versionOverride is the optional build version provided by the build system (not necessarily GIT).
func NewTelemetry(info version.Info) *Telemetry {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha2.AddToScheme(scheme))

	return &Telemetry{
		versionInfo:  info,
		configScheme: scheme,
	}
}

// GetVersionInfo returns the version information (overwritten by main.version if needed).
func (action Telemetry) GetVersionInfo() version.Info {
	return action.versionInfo
}

// GetConfigScheme returns the runtime scheme used for configuration file loading.
func (action *Telemetry) GetConfigScheme() *runtime.Scheme {
	return action.configScheme
}

// AddServerConfigOverride adds an override function that will be called in GetConfig to edit config after loading.
func (action *Telemetry) AddServerConfigOverride(override ...ServerConfigOverride) {
	action.configOverrides = append(action.configOverrides, override...)
}

// GetServerConfig returns the server configuration.
func (action *Telemetry) GetServerConfig(ctx context.Context) (*v1alpha2.ServerConfiguration, error) {
	log := logger.FromContext(ctx)

	c := &v1alpha2.ServerConfiguration{}
	if err := config.Load(log, action.GetConfigScheme(), c, action.ConfigFiles); err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	// Loop through override functions, applying each to the configuration
	for _, override := range action.configOverrides {
		if err := override(ctx, c); err != nil {
			return c, err
		}
	}

	log.InfoContext(ctx, "Using config", "configuration", c)

	return c, nil
}
