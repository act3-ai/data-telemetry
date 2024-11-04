package actions

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// ClientConfig is an action for managing the client configuration.
type ClientConfig struct {
	*Client

	Sample bool
}

// Run is the action method.
func (action *ClientConfig) Run(ctx context.Context, out io.Writer) error {
	if action.Sample {
		_, err := fmt.Fprint(out, v1alpha2.SampleClientConfig)
		return err
	}

	clientConfig, err := action.GetClientConfig(ctx)
	if err != nil {
		return err
	}

	confYAML, err := yaml.Marshal(clientConfig)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(confYAML))
	return err
}
