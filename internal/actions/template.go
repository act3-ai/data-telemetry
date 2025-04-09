package actions

import (
	"path/filepath"

	"github.com/act3-ai/data-telemetry/v3/internal/testing"
	"github.com/act3-ai/data-telemetry/v3/pkg/types"
)

// TemplateRun generates all templates for the given directory.
func TemplateRun(datadir string) error {
	// generate all files from templates
	for _, dir := range types.TopologicalOrderingOfTypes {
		if err := testing.ProcessTemplates(filepath.Join(datadir, dir)); err != nil {
			return err
		}
	}

	return nil
}
