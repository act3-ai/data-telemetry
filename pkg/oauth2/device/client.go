// Package device provides utility functions for device authentication flows as defined by RFC 8628, https://datatracker.ietf.org/doc/html/rfc8628.
package device

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// AuthPromptFn prompts the authorization of a device by providing
// the verification URI and user code to the user. The "User Interaction" step of RFC 8628
// https://www.rfc-editor.org/rfc/rfc8628#section-3.3.
type AuthPromptFn func(ctx context.Context, uri, userCode string) error

// DefaultAuthPromptFn is the default DeviceAuthPromptFn that prints to stderr.
func DefaultAuthPromptFn(ctx context.Context, uri, userCode string) error {
	_, err := fmt.Fprintf(os.Stderr, "On the device you would like to authenticate, please visit %s?user_code=%s", uri, userCode)
	return err
}

// NewOAuthClient creates an *http.Client capable of refreshing access tokens as needed. It
// prioritizes existing access and refresh tokens provided by the credentials.Store, falling
// back to device authorization as defined by RFC 8628 if necessary.
func NewOAuthClient(ctx context.Context, issuer *url.URL, clientID string, credStore credentials.Store, promptFn AuthPromptFn) (*http.Client, error) {
	// Scopes for Access and Refresh tokens, see
	// Example: https://github.com/zitadel/oidc/blob/v3.30.1/example/client/device/device.go
	// Scopes: https://github.com/zitadel/oidc/blob/main/pkg/oidc/authorization.go#L7
	scopes := []string{oidc.ScopeEmail, oidc.ScopeProfile, oidc.ScopeOfflineAccess, oidc.ScopeOpenID}
	relyParty, err := rp.NewRelyingPartyOIDC(ctx, issuer.String(), clientID, "", "", scopes, rp.WithSigningAlgsFromDiscovery())
	if err != nil {
		return nil, fmt.Errorf("error creating provider: %w", err)
	}

	oauthToken, err := refreshTokenOrAuthorize(ctx, relyParty, credStore, issuer.Host, promptFn)
	if err != nil {
		return nil, fmt.Errorf("refreshing token or authorizing device: %w", err)
	}

	return oauth2.NewClient(ctx, relyParty.OAuthConfig().TokenSource(ctx, oauthToken)), nil
}

// refreshTokenOrAuthorize returns valid access and refresh tokens from the credentials store.
// It prioritizes using an existing refresh token, falling back to device authentication.
// New tokens are saved to the credentials store.
func refreshTokenOrAuthorize(ctx context.Context, relyParty rp.RelyingParty, credStore credentials.Store,
	hostName string, promptFn AuthPromptFn) (*oauth2.Token, error) {
	// we could get the hostName with relyParty.Issuer but this includes
	// the scheme and re-parsing feels hacky, so we pass it directly
	log := logger.FromContext(ctx)

	var oauthToken *oauth2.Token
	cred, err := credStore.Get(ctx, hostName) // no error if cred DNE
	switch {
	case err == nil && cred.RefreshToken != "":
		// Always attempt a token refresh in case the refresh token itself is invalid.
		log.InfoContext(ctx, "resolving access token using existing refresh token")
		newTks, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, relyParty, cred.RefreshToken, "", "")
		if err != nil {
			log.ErrorContext(ctx, "refreshing access token", "error", err)
		} else {
			oauthToken = newTks.Token
			break // successful refresh
		}
		fallthrough
	default:
		log.InfoContext(ctx, "resolving access token with device authorization method")
		oauthToken, err = authorizeDevice(ctx, relyParty, promptFn)
		if err != nil {
			return nil, fmt.Errorf("resolving device access token: %w", err)
		}
	}

	// store updated tokens
	newCred := auth.Credential{
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
	}
	if newCred.RefreshToken == "" && cred.RefreshToken != "" {
		log.InfoContext(ctx, "new refresh token not provided, keeping old token")
		newCred.RefreshToken = cred.RefreshToken
	}
	if err := credStore.Put(ctx, hostName, newCred); err != nil {
		log.ErrorContext(ctx, "storing access and refresh tokens", "error", err)
	}

	return oauthToken, nil
}

// authorizeDevice authorizes a client using device authorization defined by RFC 8628.
// https://www.rfc-editor.org/rfc/rfc8628
func authorizeDevice(ctx context.Context, relyParty rp.RelyingParty, promptFn AuthPromptFn) (*oauth2.Token, error) {
	authResp, err := rp.DeviceAuthorization(ctx, relyParty.OAuthConfig().Scopes, relyParty, nil)
	if err != nil {
		return nil, fmt.Errorf("starting device authorization flow: %w", err)
	}

	if promptFn == nil {
		promptFn = DefaultAuthPromptFn
	}

	if err := promptFn(ctx, authResp.VerificationURI, authResp.UserCode); err != nil {
		return nil, fmt.Errorf("prompting user interaction: %w", err)
	}

	tokenResp, err := rp.DeviceAccessToken(ctx, authResp.DeviceCode, time.Duration(authResp.Interval)*time.Second, relyParty)
	if err != nil {
		return nil, fmt.Errorf("starting access token polling: %w", err)
	}

	oauthToken := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    int64(tokenResp.ExpiresIn),
	}

	return oauthToken, nil
}
