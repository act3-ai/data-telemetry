package v1alpha1

import (
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/act3-ai/go-common/pkg/redact"
)

// +kubebuilder:object:root=true

// ClientConfiguration is configuration for making requests from the upload and download subcommands.  Not used by the serve command.
type ClientConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// ClientConfigurationSpec is inlined so that all the fields in ClientConfigurationSpec are included in ClientConfiguration (not nested)
	ClientConfigurationSpec `json:",inline"`
}

// ClientConfigurationSpec is the actual configuration values.
type ClientConfigurationSpec struct {
	// Locations is the list of Telemetry server locations.  Data will be pushed to all and pulled from all
	Locations []Location `json:"locations,omitempty"`
}

// Location is a target location and the specific information needed for authentication for that location.
type Location struct {
	// Name is the display name of the telemetry server
	Name string `json:"name,omitempty"`

	// URL is the base URL for the telemetry server (does not include the /api)
	URL redact.SecretURL `json:"url,omitempty" datapolicy:"url"` // this URL should not contain any password but just incase

	// OAuth defines an OAuth2.0 provider used for authentication.
	OAuth OAuthProvider `json:"oauth,omitempty"`

	// Cookies to use for authentication
	Cookies map[string]redact.Secret `json:"cookies,omitempty" datapolicy:"values,token"`

	// Bearer token to use for authentication
	Token redact.Secret `json:"token,omitempty" datapolicy:"token"`
}

// OAuthProvider defines a host and client application ID used for OAuth2.0 Device Grant authentication
// defined by RFC 8628; see https://www.rfc-editor.org/rfc/rfc8628.
type OAuthProvider struct {
	// Issuer defines the authorization server.
	Issuer string `json:"issuer,omitempty"`
	// ClientID is the client application identifier. Not a secret.
	// See https://www.rfc-editor.org/rfc/rfc6749#section-2.2 for more info.
	ClientID string `json:"clientID,omitempty"`
}

// LogValue implements slog.LogValuer.
func (c ClientConfiguration) LogValue() slog.Value {
	return slog.AnyValue(c.ClientConfigurationSpec)
}

// LogValue implements slog.LogValuer.
func (c ClientConfigurationSpec) LogValue() slog.Value {
	values := make([]slog.Attr, 0)
	for _, l := range c.Locations {
		values = append(values, slog.Any(l.Name, l))
	}
	return slog.GroupValue(values...)
}

// LogValue implements slog.LogValuer.
func (l Location) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", l.Name),
		slog.Any("url", l.URL),
		slog.Any("cookies", l.Cookies),
		slog.Any("token", l.Token),
	)
}

// SampleClientConfig is a representative ClientConfiguration snippet.
const SampleClientConfig = `# ACE Data Telemetry Client Configuration
apiVersion: config.telemetry.act3-ace.io/v1alpha1
kind: ClientConfiguration

# Only used in the upload and download commands
request:
  locations:
  - name: Telemetry Server
    url: https://telemetry.example.com
    cookies:
      jwt: my-token
  - name: localhost
    url: http://localhost:8100
    cookies:
      foo: something else
`
