package types

import (
	"fmt"

	val "gitlab.com/act3-ai/asce/data/schema/pkg/validation"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// CosignSignatureType is a mediatype associated with cosign-generated digital signatures.
	CosignSignatureType = "dev.cosignproject.cosign/signature"
	// NotarySignatureType is a mediatype associated with notary-generated digital signatures.
	NotarySignatureType = "application/vnd.cncf.notary.payload.v1+json"
	// SignaturePayloadMediaType is the media type for cosign signature payloads.
	SignaturePayloadMediaType = "application/vnd.dev.cosign.simplesigning.v1+json"
	// AnnotationX509ChainThumbprint stores a certificate chain as a list of thumbprints. A manifest annotation key.
	// Note: Notation keeps this internal at "github.com/notaryproject/notation-go/internal/envelope", which
	// is odd as it's a required property of a notation signature.
	AnnotationX509ChainThumbprint = "io.cncf.notary.x509chain.thumbprint#S256"
	maxSignaturePayloadSize       = 5000
)

// Incoming signature information

// SignatureDetail encapsulates important information about a signature, including media types, signature data itself,
// identity and metadata.
type SignatureDetail struct {
	SignatureType string             `json:"signatureType"` // currently dev.cosignproject.cosign/signature
	Signature     string             `json:"signature"`     // base64 encoded signature
	Descriptor    ocispec.Descriptor `json:"ociDescriptor"` // data about the oci payload
	PublicKey     string             `json:"publicKey"`     // public key associated with signature (Verify)
	Annotations   map[string]string  `json:"annotations"`   // extra data, such as verify api, userid, etc.
}

// Validate SignatureDetail.
func (s SignatureDetail) Validate() error {
	if err := validation.ValidateStruct(&s,
		validation.Field(&s.SignatureType, validation.Required, validation.In(CosignSignatureType, NotarySignatureType)),
		validation.Field(&s.Signature, validation.Required),
		validation.Field(&s.Descriptor, validation.Required),
		validation.Field(&s.PublicKey, validation.When(s.SignatureType == CosignSignatureType, validation.Required)),
	); err != nil {
		return fmt.Errorf("could not validate incoming signature detail struct: %w", err)
	}

	if err := validation.ValidateStruct(&s.Descriptor,
		validation.Field(&s.Descriptor.MediaType, validation.Required, val.IsMediaType,
			validation.In(SignaturePayloadMediaType, cose.MediaTypeEnvelope, jws.MediaTypeEnvelope)),
		validation.Field(&s.Descriptor.Size, validation.Required, validation.Max(maxSignaturePayloadSize)),
		validation.Field(&s.Descriptor.Digest, validation.Required, val.IsDigest),
	); err != nil {
		return fmt.Errorf("could not validate incoming signature detail struct: %w", err)
	}

	return nil
}

// SignaturesSummary represents a summary of all signature details in a concise format.  This summary structure is
// intended to be serialized into JSON for transmission to telemetry or another location, and is not compatible with
// any OCI structures.
type SignaturesSummary struct {
	SubjectManifest digest.Digest     `json:"subjectManifest"` // manifest digest, signed object
	SubjectBottleID digest.Digest     `json:"subjectBottleid"` // bottle digest, not currently part of sig data
	Signatures      []SignatureDetail `json:"signatures"`
}

// Validate SignatureSummary.
func (s SignaturesSummary) Validate() error {
	if err := validation.ValidateStruct(&s,
		validation.Field(&s.SubjectManifest, validation.Required, val.IsDigest),
		validation.Field(&s.SubjectBottleID, validation.Required, val.IsDigest),
	); err != nil {
		return fmt.Errorf("could not validate incoming signature summary struct: %w", err)
	}

	for _, sd := range s.Signatures {
		if err := sd.Validate(); err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}
	return nil
}

// Outgoing signature information

// SignatureValidationSummary provides brief information about signatures for a bottle/manifest.  Only validation
// and Trust information is included, along with metadata annotations to describe the signature identity and intent.
type SignatureValidationSummary struct {
	SubjectManifest digest.Digest `json:"subjectManifest"` // manifest digest, signed object
	SubjectBottleid digest.Digest `json:"subjectBottleID"` // bottle digest, not currently part of sig data
	Validated       bool          `json:"sigValid"`        // true if signature was validated (self-consistent)
	Trusted         bool          `json:"sigTrusted"`      // true if signature identity was validated
	Fingerprint     string        `json:"sigFingerprint"`  // signature fingerprint data
	// TODO: add attestation key values?
	Annotations map[string]string `json:"sigAnnotations"` // signature annotations, including ident and attestations
}

// SignatureValidation provides validation details about a specific signature, along with relevant information.
type SignatureValidation struct {
	Signature   string            `json:"signature"`   // base64 encoded signature data
	PublicKey   string            `json:"publicKey"`   // public key associated with signature
	Validated   bool              `json:"validated"`   // true if the signature was validated (public key -> signature match)
	Trusted     bool              `json:"trusted"`     // true if the signature identity has been validated (fingerprint+id known)
	Annotations map[string]string `json:"annotations"` // extra data including identity and id verification details
}

// SignatureValid provides a simple view of a single signature's validation status.
type SignatureValid struct {
	BottleID  digest.Digest `json:"subjectBottleID"` // bottle digest
	KeyFp     string        `json:"keyFingerprint"`  // fingerprint of key
	Validated bool          `json:"validated"`       // true if the validation process succeeds (validated or trusted)
}

// SignatureIdentity provides a simple view of a single signature's Identity information.
type SignatureIdentity struct {
	BottleID digest.Digest     `json:"subjectBottleID"` // bottle digest
	KeyFp    string            `json:"keyFingerprint"`  // fingerprint of key
	Identity map[string]string `json:"identity"`        // a set of key-value pairs (annotations) for known key identity
}
