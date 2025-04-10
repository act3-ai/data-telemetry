package client

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/act3-ai/data-telemetry/v3/internal/actions"
	"github.com/act3-ai/data-telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// NewClientConfigCmd creates a new "config" subcommand.
func NewClientConfigCmd(clientAction *actions.Client) *cobra.Command {
	action := &actions.ClientConfig{
		Client: clientAction,
	}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show the current client configuration",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
		Example: fmt.Sprintf("Example configuration:\n\n%s", v1alpha2.SampleClientConfig),
	}

	cmd.Flags().BoolVarP(&action.Sample, "sample", "s", false,
		"Output a sample configuration that can be used in a configuration file.")

	return cmd
}
