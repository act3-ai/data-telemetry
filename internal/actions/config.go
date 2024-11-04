package actions

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// Config is the action for getting the server configuration.
type Config struct {
	*Telemetry

	Sample bool
}

// Run is the action method.
func (action *Config) Run(ctx context.Context, out io.Writer) error {
	if action.Sample {
		_, err := fmt.Fprint(out, v1alpha2.SampleServerConfig)
		return err
	}

	serverConfig, err := action.GetServerConfig(ctx)
	if err != nil {
		return err
	}

	confYAML, err := yaml.Marshal(serverConfig)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(confYAML))
	return err
}
