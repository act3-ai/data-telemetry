package db

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"

	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"
	"gitlab.com/act3-ai/asce/data/schema/pkg/util"
)

// Model is the base of all records.
type Model struct {
	gorm.Model
}

// GetPrimaryKey gets the primary key.
func (m Model) GetPrimaryKey() uint {
	return m.ID
}

// SetPrimaryKey sets the primary key.
func (m *Model) SetPrimaryKey(pk uint) {
	m.ID = pk
}

// Base is the common base for all four core types.
type Base struct {
	Model
	ProcessorVersion uint `gorm:"index"`
	Data             Data // Base belongs to Data
	DataID           uint
}

// Data stores the actual data for all types (blob, bottles, manifests, events).
type Data struct {
	Model
	RawData         []byte
	CanonicalDigest digest.Digest `gorm:"uniqueIndex"` // canonical digest is an internal only digest (like blake2 or sha3 used for de-duplication)
}

// CanonicalDigestAlgorithm is the digest used internally for de-duplication and discovering aliases.
const CanonicalDigestAlgorithm = digest.Canonical

// Canonical digest should be a very fast algorithm (security and universality are less important since it is only used internally)
// BLAKE3-256 is probably the best option

// Digest is a table of aliases.
type Digest struct {
	Model
	Data   Data // Digest belongs to Data
	DataID uint
	Digest digest.Digest `gorm:"uniqueIndex"` // A user provided name/digest for this data
}

// Blob is raw data (typically a public artifact of a bottle).
type Blob struct {
	Base
}

// BottleMemberLocated is the base for all "has many" members of a Bottle.
type BottleMemberLocated struct {
	Model
	BottleID uint // foreign key

	// Location in the array from JSON
	Location uint
}

// GetLocation gets the location in the original array.
func (b BottleMemberLocated) GetLocation() uint {
	return b.Location
}

// PublicArtifact is metadata about a PublicArtifact as specified in the Bottle.
type PublicArtifact struct {
	BottleMemberLocated

	Name      string
	MediaType string
	Path      string `gorm:"index"` // unique per bottle
	Data      Data   // PublicArtifact belongs to Data
	DataID    uint

	// Digest is the digest of the artifact.
	// We need to preserve the digest used by the bottle to reference this artifact.
	Digest digest.Digest // `gorm:"index"` we might want an index here
}

// Source is a data source for the bottle.
type Source struct {
	BottleMemberLocated

	Name string `gorm:"index"`
	URI  string `gorm:"index"`

	// optional if the URL is a bottle
	BottleDigest digest.Digest `gorm:"index"` // we do not guarantee this exists in the telemetry server so this is just a string, not a DataID, DigestID, or BottleID

	// PartSelectors are used in the code but not saved in the database. They are derived from the URI query parameters
	PartSelectors selectors.LabelSelectorSet `gorm:"-"`
}

// AfterFind is called after a find() to extract the part selectors to the LabelSelectorSet.
func (s *Source) AfterFind(tx *gorm.DB) error {
	if tx.Error != nil {
		return tx.Error
	}
	_, partSelectors, err := util.ParseSourceURI(s.URI)
	if err != nil {
		return fmt.Errorf("could not parse selectors from URI (%s): %w", s.URI, err)
	}

	s.PartSelectors = partSelectors
	return nil
}

// BeforeSave is called before the struct is saved to the DB to convert the part selectors back to URI.
func (s *Source) BeforeSave(tx *gorm.DB) error {
	u, err := url.Parse(s.URI)
	if err != nil {
		return fmt.Errorf("could not parse source URI: %w", err)
	}

	// rebuild the URI fragment before saving it
	q := u.Query()
	q.Del("selector")
	for _, p := range s.PartSelectors {
		q.Add("selector", p.String())
	}
	u.RawQuery = q.Encode()

	s.URI = u.String()
	return nil
}

// Part is a data part of a Bottle.
type Part struct {
	BottleMemberLocated

	Name   string // unique per bottle
	Size   uint64
	Digest digest.Digest `gorm:"index"` // we do not guarantee this exists in the telemetry server so this is just a string, not a Digest and DigestID

	// We do not index PartLabels because those are specific to the bottle
	// Labels []PartLabel

	// LabelStr is stored in the DB but code is expected to use Labels directly.
	LabelsStr string            `json:"-"`
	Labels    map[string]string `gorm:"-"`
}

// AfterFind is called after a find() to convert the LabelStr to the Labels map.
func (p *Part) AfterFind(tx *gorm.DB) error {
	if tx.Error == nil {
		set, err := labels.ConvertSelectorToLabelsMap(p.LabelsStr)
		if err != nil {
			return fmt.Errorf("parsing labels from database record: %w", err)
		}
		p.Labels = set
	}
	return nil
}

// BeforeSave is called before the struct is saved to the DB to convert the Label map to the LabelStr.
func (p *Part) BeforeSave(tx *gorm.DB) error {
	p.LabelsStr = labels.Set(p.Labels).String()
	return nil
}

// Author is author information.
type Author struct {
	BottleMemberLocated

	Name  string `gorm:"index"`
	URL   string // TODO should this be URI as well?
	Email string `gorm:"index"`
}

// Label is bottle labels.
type Label struct {
	Model
	BottleID uint

	Key          string  `gorm:"index"` // unique per bottle
	Value        string  `gorm:"index"`
	NumericValue float64 `gorm:"index" json:"-"` // We drop this is JSON marshalling to avoid issues with logging
}

// GetLocation gets the index (key of the label).
func (l Label) GetLocation() string {
	return l.Key
}

// BeforeSave is called before the struct is saved to the DB to convert the value to a float if possible.
func (l *Label) BeforeSave(tx *gorm.DB) error {
	val, err := strconv.ParseFloat(l.Value, 64)
	if err != nil {
		val = math.NaN()
	}
	l.NumericValue = val
	return nil
}

// Annotation is bottle annotations.
type Annotation struct {
	Model
	BottleID uint

	Key   string // unique per bottle
	Value string
}

// GetLocation gets the index (key of the annotation).
func (a Annotation) GetLocation() string {
	return a.Key
}

// Metric is metrics on Bottles.
type Metric struct {
	BottleMemberLocated

	Name        string // unique per bottle
	Description string
	Value       float64
}

// Bottle is the metadata for a Bottle.
type Bottle struct {
	Base
	APIVersion      string
	Description     string
	Authors         []Author         // Bottle has many authors
	Sources         []Source         // Bottle has many sources
	Metrics         []Metric         // Bottle has many metrics
	PublicArtifacts []PublicArtifact // Bottle has many public_artifacts
	Labels          []Label          // Bottle has many labels
	Annotations     []Annotation     // Bottle has many annotations
	Parts           []Part           // Bottle has many parts
	Deprecates      []Deprecates     // Bottle has many deprecates
	Signatures      []Signature      // Bottle has many signatures
}

// Layer is a manifest layer.
type Layer struct {
	Model
	ManifestID uint
	Location   uint

	Digest digest.Digest `gorm:"index"` // This is not tracked by the telemetry server so it is just a string (not a reference to a Digest record)
	// TODO: Add other OCI descriptor fields
}

// GetLocation gets the index (key of the annotation).
func (l Layer) GetLocation() uint {
	return l.Location
}

// Manifest is the OCI manifest v2 and points to a Bottle.
type Manifest struct {
	Base
	BottleID     uint
	Bottle       Bottle        // Manifest belongs to Bottle
	BottleDigest digest.Digest `gorm:"index"`
	Layers       []Layer       // Bottle has many layers
}

// Event is used to record an actual download/upload event.
type Event struct {
	Base

	ManifestID uint
	Manifest   Manifest // Event belongs to Manifest
	// While a manifest may have different digests (different algorithms) this event is specific to this manifest.
	// The repository in this event may only have one (name) digest for this Manifest.
	ManifestDigest digest.Digest `gorm:"index"`

	BottleID     uint // Event belongs to Bottle (prejoining since the association is not allowed to change)
	Bottle       Bottle
	BottleDigest digest.Digest `gorm:"index"` // prejoin since it does not change

	Action       string // pull or push
	Repository   string
	Tag          string
	AuthRequired bool
	Bandwidth    uint64
	Timestamp    time.Time `gorm:"index"`
	Username     string    `gorm:"index"`
}

// Deprecates is a deprecated Bottle.
type Deprecates struct {
	BottleMemberLocated

	DeprecatedBottleDigest digest.Digest
}

// Signature is used to record "signed-off" attributes for a bottle.
type Signature struct {
	Base

	ManifestID uint
	Manifest   Manifest // Event belongs to Manifest
	// While a manifest may have different digests (different algorithms) this signature is specific to this manifest.
	// The repository in this signature may only have one (name) digest for this Manifest.
	ManifestDigest digest.Digest `gorm:"index"`

	BottleID     uint // Signature belongs to Bottle (prejoining since the association is not allowed to change)
	Bottle       Bottle
	BottleDigest digest.Digest `gorm:"index"` // prejoin since it does not change

	SignatureType        string                // currently dev.cosignproject.cosign/signature
	Signature            []byte                // raw signature payload data
	PublicKey            string                // PEM encoded public key associated with signature (Verify)
	PublicKeyFingerPrint digest.Digest         `gorm:"index"` // the digest of the public key
	Annotations          []SignatureAnnotation // extra data, such as verify api, userid, etc.

	// Trusted is used in the code but not saved in the database.
	Trusted func(TrustAnchor) bool `gorm:"-"` // true if the signature identity can be validated against a given trust anchor (fingerprint+id known)
}

// AfterFind is called after a find() to add the Trusted func to the signature.
func (s *Signature) AfterFind(tx *gorm.DB) error {
	s.Trusted = func(ta TrustAnchor) bool { return ta.VerifyTrust() }
	return nil
}

// SignatureAnnotation is used to record extra data on a signature, such as verify api, userid, etc.
type SignatureAnnotation struct {
	Model
	SignatureID uint

	Key   string // unique per bottle
	Value string
}
