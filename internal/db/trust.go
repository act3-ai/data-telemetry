package db

// TrustAnchor is a source which we can validate an signature owner's identity against.
type TrustAnchor interface {
	VerifyTrust() bool
}

// DefaultTrustAnchor can be used to get a trusted value back when no other trust anchor is available.
type DefaultTrustAnchor struct{}

// VerifyTrust on the DefaultTrustAnchor always returns false.
func (t *DefaultTrustAnchor) VerifyTrust() bool {
	return false
}
