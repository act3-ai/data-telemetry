// Package signature provides shared signature validation code functions
package signature

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// simpleTrustPolicy creates a basic trust policy for certificate validation.
func simpleTrustPolicy() *trustpolicy.Document {
	return &trustpolicy.Document{
		Version: "1.0",
		TrustPolicies: []trustpolicy.TrustPolicy{
			{
				Name:           "bottle-verify-policy",
				RegistryScopes: []string{"local/bottle"},
				SignatureVerification: trustpolicy.SignatureVerification{
					VerificationLevel: trustpolicy.LevelStrict.Name,
					Override: map[trustpolicy.ValidationType]trustpolicy.ValidationAction{
						trustpolicy.TypeRevocation: trustpolicy.ActionSkip,
					},
				},
				TrustStores:       []string{"ca:valid-trust-store"},
				TrustedIdentities: []string{"*"},
			},
		},
	}
}

// ValidateSignatureNotary performs certificate signature validation using Notary.  This function constructs
// the necessary options and trust policy from a subject descriptor expected to have been signed, and a trust store
// that provides a x509 certificate.
func ValidateSignatureNotary(ctx context.Context, subjectDesc v1.Descriptor, sig []byte, trustStore truststore.X509TrustStore) (*notation.VerificationOutcome, error) {
	// simple trust policy document.  TODO: replace this with something that assures trust
	policyDoc := simpleTrustPolicy()

	// set verification options, using a local/bottle generic artifact reference.
	verifyOptions := notation.VerifierVerifyOptions{
		ArtifactReference:  "local/bottle@" + subjectDesc.Digest.String(),
		SignatureMediaType: jws.MediaTypeEnvelope,
	}

	// create a verifier, using the certificates included in this object, and the basic policy document.
	notationVerifier, err := verifier.New(policyDoc, trustStore, nil)
	if err != nil {
		return nil, fmt.Errorf("simple cert create verifier: %w", err)
	}

	// verify the signature.
	outcome, err := notationVerifier.Verify(ctx, subjectDesc, sig, verifyOptions)
	if err != nil {
		return nil, fmt.Errorf("validating signature: %w", err)
	}
	return outcome, nil
}

// pubPEMToECDSA converts a PEM encoded public key into a raw ECDSA public key.
func pubPEMToECDSA(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	// decode PEM to DER format
	pubKeyDER, rest := pem.Decode(pubKeyBytes) // discard leftover bytes
	if pubKeyDER == nil || pubKeyDER.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("decoding PEM key to DER format; type = %s, rest = %s", pubKeyDER.Type, rest)
	}

	// parse the ecdsa key
	pubECDSAKey, err := x509.ParsePKIXPublicKey(pubKeyDER.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing DER bytes: %w", err)
	}

	return pubECDSAKey.(*ecdsa.PublicKey), nil
}

// ValidateSignatureCosign verifies the provided payload is signed with the provided signature, using cosign PKI sig
// Currently only supports sha256 hashing algorithm.
func ValidateSignatureCosign(ctx context.Context, pubkey, sig []byte, payloadHash digest.Digest) (bool, error) {
	log := logger.FromContext(ctx)

	publicKey, err := pubPEMToECDSA(pubkey)
	if err != nil {
		return false, fmt.Errorf("decoding public key from PEM: %w", err)
	}
	decodedHash, err := hex.DecodeString(payloadHash.Encoded())
	if err != nil {
		return false, fmt.Errorf("decoding payload hash: %w", err)
	}
	if ecdsa.VerifyASN1(publicKey, decodedHash, sig) {
		log.DebugContext(ctx, "Sig verified")
		return true, nil
	}
	log.DebugContext(ctx, "Sig NOT verified")
	return false, nil
}
