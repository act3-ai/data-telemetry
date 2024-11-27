package db

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	telemsig "gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/signature"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/types"
)

// SignatureProcessorVersion is the current version of the processor code.  This is incremented after each measurable change to the SignatureProcessor().
const SignatureProcessorVersion = 1

// SignatureProcessor handles bottle processing.
type SignatureProcessor struct{}

// Version returns the processor version.
func (p *SignatureProcessor) Version() uint {
	return SignatureProcessorVersion
}

// PrimaryTable returns primary table that this processor updates.
func (p *SignatureProcessor) PrimaryTable() string {
	return "signatures"
}

// Process converts Signature data to the DB model.
func (p *SignatureProcessor) Process(con *gorm.DB, base Base) error {
	ctx := con.Statement.Context

	var signatureDto types.SignaturesSummary
	if err := json.Unmarshal(base.Data.RawData, &signatureDto); err != nil {
		return httputil.NewHTTPError(err, http.StatusConflict, "Failed to parse signature", "request data", string(base.Data.RawData))
	}

	// input validation
	if err := signatureDto.Validate(); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid signature definition: "+err.Error(), "signature", signatureDto)
	}

	// Find the manifest to which this signature corresponds.
	tx := con.
		Preload("Bottle").
		Preload("Layers").
		Scopes(FilterByDigest(signatureDto.SubjectManifest, "manifests"))
	dbManifest := Manifest{}
	if err := tx.First(&dbManifest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewMissingDigestsError("manifest", []digest.Digest{signatureDto.SubjectManifest})
		}
		return err
	}

	// Find the bottle to which this signature corresponds.
	tx = con.
		Table("bottles").
		Scopes(FilterByDigest(signatureDto.SubjectBottleID, "bottles"))
	dbBottle := Bottle{}
	if err := tx.First(&dbBottle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "Could not find bottle for signature", "signature", signatureDto)
		}
		return err
	}

	for _, s := range signatureDto.Signatures {
		annotations := []SignatureAnnotation{}
		for k, v := range s.Annotations {
			annotations = append(annotations, SignatureAnnotation{
				Key:   k,
				Value: v,
			})
		}

		publicKeyFP := keyFingerPrint(s)

		signatureBytes, err := base64.StdEncoding.DecodeString(s.Signature)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "could not decode signature. expected base64 format", "signature", signatureDto)
		}

		sigValid, err := validateSignature(ctx, s)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusInternalServerError, "read error during signature validation", "signature", signatureDto)
		}
		if !sigValid {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "invalid signature", "signature", signatureDto)
		}

		dbSignature := Signature{
			Base: base,

			ManifestID:     dbManifest.ID,
			ManifestDigest: signatureDto.SubjectManifest,

			BottleID:     dbBottle.ID,
			BottleDigest: signatureDto.SubjectBottleID,

			SignatureType:        s.SignatureType,
			Signature:            signatureBytes,
			PublicKey:            s.PublicKey,
			PublicKeyFingerPrint: publicKeyFP,
			Annotations:          annotations,
		}

		if err := con.Session(&gorm.Session{FullSaveAssociations: true}).
			Save(&dbSignature).Error; err != nil {
			return err
		}

	}

	return nil
}

// keyFingerPrint determines a public key fingerprint or signing certificate fingerprint from signature details
// This is based on the public key if present (in the case of cosign signatures), or the fingerprint annotation
// relevant to notary style signatures.
func keyFingerPrint(sigDetail types.SignatureDetail) digest.Digest {
	publicKeyFP := digest.FromString(sigDetail.PublicKey)
	// publicKey may be empty if certificate based signature, in which case try to get the fingerprint digest
	// from annotations.
	if a, ok := sigDetail.Annotations[types.AnnotationX509ChainThumbprint]; ok {
		// if we have a notary certificate chain thumbprint, use that as the fingerprint.  Note, the annotation key
		// mandates sh256 algorithm, so we can assume that here.
		publicKeyFP = digest.NewDigestFromEncoded(digest.SHA256, a)
	}
	return publicKeyFP
}

// validateSignature validates a signature, using a validation function appropriate for the type of signature provided.
func validateSignature(ctx context.Context, sigDetail types.SignatureDetail) (bool, error) {
	signatureBytes, err := base64.StdEncoding.DecodeString(sigDetail.Signature)
	if err != nil {
		return false, err
	}

	switch sigDetail.SignatureType {
	case types.NotarySignatureType:
		return validateSignatureNotary(ctx, sigDetail.Descriptor.MediaType, signatureBytes)
	case types.CosignSignatureType:
		return telemsig.ValidateSignatureCosign(ctx, []byte(sigDetail.PublicKey), signatureBytes, sigDetail.Descriptor.Digest)
	}
	// because of prior data validation, this case cannot be reached
	return false, fmt.Errorf("unexpected signature type: %s", sigDetail.SignatureType)
}

// trustStoreProxy is a simple structure to act as a trust store interface provider.
type trustStoreProxy struct {
	certs []*x509.Certificate
}

// GetCertificates satisfies the TrustStore.X509TrustStore interface for accessing a certificate during validation.
func (tsp *trustStoreProxy) GetCertificates(ctx context.Context, storeType truststore.Type, namedStore string) ([]*x509.Certificate, error) {
	return tsp.certs, nil
}

// getCertAndSubjectFromNotarySig returns a x509 trust store based on the certificate in the provided signature data,
// as well as a subject descriptor that is signed.  The sig is expected to be a notary style certificate signature.
func getCertAndSubjectFromNotarySig(sigMediaType string, sigData []byte) (truststore.X509TrustStore, v1.Descriptor, error) {
	var env signature.Envelope
	var err error
	switch sigMediaType {
	case jws.MediaTypeEnvelope:
		env, err = jws.ParseEnvelope(sigData)
		if err != nil {
			return nil, v1.Descriptor{}, fmt.Errorf("parsing jws signature envelope: %w", err)
		}
	case cose.MediaTypeEnvelope:
		env, err = cose.ParseEnvelope(sigData)
		if err != nil {
			return nil, v1.Descriptor{}, fmt.Errorf("parsing cose signature envelope: %w", err)
		}
	default:
		return nil, v1.Descriptor{}, &signature.UnsupportedSignatureFormatError{MediaType: sigMediaType}
	}
	if err != nil {
		return nil, v1.Descriptor{}, fmt.Errorf("parsing signature envelope: %w", err)
	}

	content, err := env.Content()
	if err != nil {
		return nil, v1.Descriptor{}, fmt.Errorf("extracting '%s' envelope content: %w", sigMediaType, err)
	}

	subject := &struct{ TargetArtifact v1.Descriptor }{}
	err = json.Unmarshal(content.Payload.Content, subject)
	if err != nil {
		return nil, v1.Descriptor{}, fmt.Errorf("extracting subject from envelope %w", err)
	}

	cert := content.SignerInfo.CertificateChain[len(content.SignerInfo.CertificateChain)-1]
	tsp := &trustStoreProxy{certs: []*x509.Certificate{cert}}

	return tsp, subject.TargetArtifact, nil
}

// validateSignatureNotary verifies the provided payload is signed with the provided signature using certificate based
// notary signing.
func validateSignatureNotary(ctx context.Context, sigMediaType string, sig []byte) (bool, error) {
	log := logger.FromContext(ctx)

	// retrieve the certificate and subject of the signature from signature data.
	trustStore, subjectDesc, err := getCertAndSubjectFromNotarySig(sigMediaType, sig)
	if err != nil {
		return false, fmt.Errorf("retrieving signing certificate: %w", err)
	}

	// verify the signature.  Failure is returned as an error, so we don't need to check outcome directly.
	_, err = telemsig.ValidateSignatureNotary(ctx, subjectDesc, sig, trustStore)
	// if failed, return a signature verification error.
	if err != nil {
		log.DebugContext(ctx, "Sig NOT verified")
		return false, nil
	}

	log.DebugContext(ctx, "Sig verified")
	return true, nil
}

// GetSignaturesWithAnnotations takes a list of signatures without annotations and returns a slice of those same signatures with annotations included.
func GetSignaturesWithAnnotations(ctx context.Context, con *gorm.DB, signatures *[]Signature) (*[]Signature, error) {
	fullSignatures := make([]Signature, 0)

	// pull out signature IDs to be easier to work with
	signatureIDMap := make(map[uint]*Signature, 0)
	signatureIDs := make([]uint, 0)
	for _, s := range *signatures {
		signatureIDMap[s.ID] = &s
		signatureIDs = append(signatureIDs, s.ID)
	}

	// get all annotations
	dbSignatureAnnotations := make([]SignatureAnnotation, 0)
	query := con.Session(&gorm.Session{NewDB: true}).
		Table("signature_annotations").
		Where("signature_annotations.signature_id IN (?)", signatureIDs)
	if err := query.Find(&dbSignatureAnnotations).Error; err != nil {
		return &fullSignatures, err
	}

	for _, sa := range dbSignatureAnnotations {
		sig := signatureIDMap[sa.SignatureID]
		sig.Annotations = append(sig.Annotations, sa)
	}

	for _, s := range signatureIDMap {
		fullSignatures = append(fullSignatures, *s)
	}

	return &fullSignatures, nil
}
