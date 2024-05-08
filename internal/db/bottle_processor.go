package db

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/schema/pkg/util"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
)

// BottleProcessorVersion is the current version of the processor code.  This is incremented after each measurable change to the BottleProcessor().
const BottleProcessorVersion = 11

// BottleProcessor handles bottle processing.
type BottleProcessor struct {
	scheme *runtime.Scheme
	codecs *serializer.CodecFactory
}

// NewBottleProcessor creates a new BottleProcessor populating the Scheme and Codecs.
func NewBottleProcessor(scheme *runtime.Scheme) *BottleProcessor {
	codecs := serializer.NewCodecFactory(scheme, serializer.EnableStrict)

	return &BottleProcessor{scheme, &codecs}
}

// Version returns the processor version.
func (p *BottleProcessor) Version() uint {
	return BottleProcessorVersion
}

// PrimaryTable returns primary table that this processor updates.
func (p *BottleProcessor) PrimaryTable() string {
	return "bottles"
}

// Process converts Bottle data to the DB model for a Bottle.
func (p *BottleProcessor) Process(con *gorm.DB, base Base) error {
	// call the conversion function from ace/data/schema to convert to the latest bottle schema
	bottleDto := latest.Bottle{}
	if err := runtime.DecodeInto(p.codecs.UniversalDecoder(), base.Data.RawData, &bottleDto); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid JSON for bottle", "request data", string(base.Data.RawData))
	}

	// input validation
	// TODO we validate the upgraded bottle which might not validate once upgraded due to missing fields.
	// We might want to validate the original bottle instead.
	if err := bottleDto.Validate(); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle definition: "+err.Error(), "bottle", bottleDto)
	}

	dbBottle := Bottle{
		Base:        base,
		APIVersion:  bottleDto.APIVersion, // TODO this is the converted (upgraded) version (v1beta1) and not the old/original version.
		Description: bottleDto.Description,
	}

	// Process the Sources
	sources, err := processByIndex(con, &dbBottle, "Sources", bottleDto.Sources, convertSource)
	if err != nil {
		return err
	}
	dbBottle.Sources = sources

	// Process the Authors
	authors, err := processByIndex(con, &dbBottle, "Authors", bottleDto.Authors, convertAuthor)
	if err != nil {
		return err
	}
	dbBottle.Authors = authors

	// Process the Metrics
	metrics, err := processByIndex(con, &dbBottle, "Metrics", bottleDto.Metrics, convertMetric)
	if err != nil {
		return err
	}
	dbBottle.Metrics = metrics

	// Pre-Process the PublicArtifacts
	missingArtifactBlobs := make([]digest.Digest, 0, len(bottleDto.PublicArtifacts))
	artifactAliases := make(map[digest.Digest]uint, len(bottleDto.PublicArtifacts))
	for _, a := range bottleDto.PublicArtifacts {
		// TODO batch this with FindInBatches() or with primary key with Where()
		// search the "digests" table
		alias := Digest{}
		if err := con.Where(&Digest{Digest: a.Digest}).First(&alias).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Artifact not found
				missingArtifactBlobs = append(missingArtifactBlobs, a.Digest)
				continue
			}
			return err
		}
		artifactAliases[a.Digest] = alias.DataID
	}
	if len(missingArtifactBlobs) > 0 {
		return types.NewMissingDigestsError("blobs", missingArtifactBlobs, "publicArtifacts", bottleDto.PublicArtifacts)
	}

	// Process the PublicArtifacts
	publicArtifacts, err := processByIndex(con, &dbBottle, "PublicArtifacts", bottleDto.PublicArtifacts, convertPublicArtifact(artifactAliases))
	if err != nil {
		return err
	}
	dbBottle.PublicArtifacts = publicArtifacts

	// Process the Labels
	labels, err := processByKey(con, &dbBottle, "Labels", bottleDto.Labels, convertLabel)
	if err != nil {
		return err
	}
	dbBottle.Labels = labels

	// Process the Annotations
	annotations, err := processByKey(con, &dbBottle, "Annotations", bottleDto.Annotations, convertAnnotation)
	if err != nil {
		return err
	}
	dbBottle.Annotations = annotations

	// Process the Deprecates field
	deprecates := make([]Deprecates, len(bottleDto.Deprecates))
	for i, dgst := range bottleDto.Deprecates {
		deprecates[i] = Deprecates{DeprecatedBottleDigest: dgst}
	}
	dbBottle.Deprecates = deprecates

	// Process the parts
	parts, err := processByIndex(con, &dbBottle, "Parts", bottleDto.Parts, convertPart)
	if err != nil {
		return err
	}
	dbBottle.Parts = parts

	return con.Session(&gorm.Session{FullSaveAssociations: true}).Save(&dbBottle).Error
}

// convertSource process a single source entry.
func convertSource(old Source, i int, s latest.Source) (*Source, error) {
	bottleDigest, partSelectors, err := util.ParseSourceURI(s.URI)
	if err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid URI in sources")
	}

	source := Source{
		Name:          s.Name,
		URI:           s.URI,
		BottleDigest:  bottleDigest,
		PartSelectors: partSelectors,
	}
	source.ID = old.ID
	source.Location = uint(i)
	return &source, nil
}

func convertMetric(old Metric, i int, m latest.Metric) (*Metric, error) {
	value, err := strconv.ParseFloat(m.Value, 64)
	if err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid metric value")
	}
	metric := Metric{
		Name:        m.Name,
		Description: m.Description,
		Value:       value,
	}
	metric.ID = old.ID
	metric.Location = uint(i)
	return &metric, nil
}

func convertPublicArtifact(aliases map[digest.Digest]uint) func(old PublicArtifact, i int, a latest.PublicArtifact) (*PublicArtifact, error) {
	return func(old PublicArtifact, i int, a latest.PublicArtifact) (*PublicArtifact, error) {
		artifact := PublicArtifact{
			Name:      a.Name,
			MediaType: a.MediaType,
			Path:      a.Path,
			DataID:    aliases[a.Digest],
			Digest:    a.Digest,
		}
		artifact.ID = old.ID
		artifact.Location = uint(i)
		return &artifact, nil
	}
}

func convertAuthor(old Author, i int, a latest.Author) (*Author, error) {
	author := Author{
		Name:  a.Name,
		URL:   a.URL,
		Email: a.Email,
	}
	author.ID = old.ID
	author.Location = uint(i)
	return &author, nil
}

func convertLabel(old Label, key string, value string) (*Label, error) {
	label := Label{
		Key:   key,
		Value: value,
	}
	label.ID = old.ID
	return &label, nil
}

func convertAnnotation(old Annotation, key string, value string) (*Annotation, error) {
	annotation := Annotation{
		Key:   key,
		Value: value,
	}
	annotation.ID = old.ID
	return &annotation, nil
}

func convertPart(old Part, i int, p latest.Part) (*Part, error) {
	part := Part{
		Name:   p.Name,
		Size:   uint64(p.Size),
		Digest: p.Digest,
		Labels: p.Labels,
	}
	part.ID = old.ID
	part.Location = uint(i)
	return &part, nil
}
