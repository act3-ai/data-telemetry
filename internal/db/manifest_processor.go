package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"

	"gitlab.com/act3-ai/asce/data/schema/pkg/validation"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
)

// ManifestProcessorVersion is the current version of the processor code.  This is incremented after each measurable change to the ManifestProcessor().
const ManifestProcessorVersion = 4

// ManifestProcessor handles bottle processing.
type ManifestProcessor struct{}

// Version returns the processor version.
func (p *ManifestProcessor) Version() uint {
	return ManifestProcessorVersion
}

// PrimaryTable returns primary table that this processor updates.
func (p *ManifestProcessor) PrimaryTable() string {
	return "manifests"
}

// Process converts Manifest data to the DB model.
func (p *ManifestProcessor) Process(con *gorm.DB, base Base) error {
	// parse the manifest
	manifest := ocispec.Manifest{}
	if err := json.Unmarshal(base.Data.RawData, &manifest); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Manifest is invalid", "request data", string(base.Data.RawData))
	}

	if manifest.MediaType == "" {
		manifest.MediaType = "application/vnd.oci.image.manifest.v1+json"
	}

	if err := validation.ValidateManifest(manifest); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid manifest definition: "+err.Error(), "manifest", manifest)
	}

	// lookup the bottle by digest to get its primary key (ID)
	bottleDigest := manifest.Config.Digest
	var bottle Bottle
	tx := con.Scopes(FilterByDigest(bottleDigest, "bottles")).Preload("Parts")
	if err := tx.First(&bottle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewMissingDigestsError("bottle", []digest.Digest{bottleDigest})
		}
		return err
	}

	// validate that this manifest is compatible with the bottle
	if err := validateManifestAgainstBottle(manifest, bottle); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Manifest is not compatible with the referenced bottle: "+err.Error(), "manifest", manifest, "bottle", bottleDigest)
	}

	dbManifest := Manifest{
		Base:         base,
		BottleID:     bottle.ID,
		BottleDigest: bottleDigest,
	}

	// Process the Layers
	layers, err := processByIndex(con, &dbManifest, "Layers", manifest.Layers, convertLayer)
	if err != nil {
		return err
	}
	dbManifest.Layers = layers

	return con.Session(&gorm.Session{FullSaveAssociations: true}).
		Save(&dbManifest).Error
}

func validateManifestAgainstBottle(m ocispec.Manifest, b Bottle) error {
	nl := len(m.Layers)
	np := len(b.Parts)
	if nl != np {
		return fmt.Errorf("there are %d layers but %d parts (they must be equal)", nl, np)
	}
	// TODO add validation of the media type of the descriptor matches the part type (file or directory)
	// We need to add a "type" field to the schema first. This would hep avoid a simple attack on the integrity of the bottle (replacing a directory part with a file part of the same digest, the .tar file itself)
	// TODO call schema's manifest validator
	return nil
}

func convertLayer(old Layer, i int, l ocispec.Descriptor) (*Layer, error) {
	layer := Layer{
		Digest: l.Digest,
	}
	layer.ID = old.ID
	layer.Location = uint(i)
	return &layer, nil
}
