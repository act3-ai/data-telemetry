package client

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-telemetry/v3/internal/actions"
)

// NewUploadCmd creates a new "upload" command.
func NewUploadCmd(clientAction *actions.Client) *cobra.Command {
	action := &actions.Upload{
		Client: clientAction,
	}

	cmd := &cobra.Command{
		Use:   "upload <path> <url>",
		Short: "Upload test data at <path> into the server at <url>",
		Long:  `This command is used for testing as well as restoring from backups and replication.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&action.All, "all", true, "all object types (specify the <path> as the test set directory)")
	cmd.Flags().BoolVar(&action.SkipInvalid, "continue", false, "continue uploading after encountering an invalid manifest error")

	return cmd
}
