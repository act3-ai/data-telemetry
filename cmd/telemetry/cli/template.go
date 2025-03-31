package cli

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/internal/actions"
)

// NewTemplateCmd creates a new "template" command.
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template <dir>",
		Short: "Template data in the given directory",
		Long: `This assumes that there <dir> has subdirectories of artifact, bottle, manifest, and event.  It will template each in turn.

All the Sprig functions are available along with the following functions:

FileSize(filename string) int64
FileDigest(filename string, algorithm digest.Algorithm) digest.Digest
Digest(algorithm digest.Algorithm, data []byte) digest.Digest
BottleURI(hashScheme string, dgst digest.Digest, partSelectors ...string) string
// partSelectors (optional) are part selector strings like  "partkey!=value1,mykey=value2" and "partkey2=45"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.TemplateRun(args[0])
		},
	}

	return cmd
}
