package client

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-telemetry/v3/internal/actions"
)

// NewDownloadCmd creates a new "upload" command.
func NewDownloadCmd(clientAction *actions.Client) *cobra.Command {
	action := &actions.Download{
		Client: clientAction,
	}

	cmd := &cobra.Command{
		Use:   "download <path> <url>",
		Short: "Download data to <path> from the server at [<url>]",
		Long:  `Typically this command is used for replication or backup processes.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&action.Since, "since", "",
		"Date of which to start pulling data")

	cmd.Flags().IntVarP(&action.BatchSize, "batch_size", "b", 100,
		"Maximum size of a batch of records to download")

	cmd.Flags().BoolVar(&action.All, "all", true, "all object types (specify the <path> as the test set directory)")

	cmd.Flags().BoolVar(&action.FromLatest, "from-latest", true, `read the index.latest file for each type and only request the data since then.
		This is very useful when combined with --all to provide incremental backups/mirrors.`)

	return cmd
}
