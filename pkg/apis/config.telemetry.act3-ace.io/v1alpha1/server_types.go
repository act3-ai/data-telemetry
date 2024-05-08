package v1alpha1

import (
	// "log/slog".

	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
)

// +kubebuilder:object:root=true

// ServerConfiguration is the Schema for the Telemetry Server Configurations API.
type ServerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// ServerConfigurationSpec is inlined so that all the fields in ServerConfigurationSpec are included in ServerConfiguration (not nested)
	ServerConfigurationSpec `json:",inline"`
}

// ServerConfigurationSpec is the actual configuration values.
type ServerConfigurationSpec struct {
	// DB is the database configuration
	DB Database `json:"db,omitempty"`

	// WebApp specific configuration
	WebApp WebApp `json:"webapp,omitempty"`
}

// Database is configuration for the database connection.
type Database struct {
	// DSN is the database connection string
	DSN redact.SecretURL `json:"dsn" datapolicy:"url"`
}

// WebApp is the configuration for the telemetry web application.
type WebApp struct {
	// ACEHubs is a list of ace hub instances that will be available to users of the web application for viewing bottles
	ACEHubs []ACEHubInstance `json:"acehubs,omitempty"`

	// Viewers is a list of viewer specifications for how to launch the viewer
	Viewers []ViewerSpec `json:"viewers,omitempty"`

	// JupyterExecutable is the path to the "jupyter" executable
	JupyterExecutable string `json:"jupyter,omitempty"`

	// DefaultBottleSelectors is the list of selectors to use when searching the catalog by default
	DefaultBottleSelectors []string `json:"defaultBottleSelectors,omitempty"`

	// AssetDir is the directory where the web assets reside, default value is "assets"
	AssetDir string `json:"assets,omitempty"`
}

// ACEHubInstance is an existing instance of ACE Hub that will be offered as a bottle viewer engine.
type ACEHubInstance struct {
	// Name is the name of the instance for display purposes
	Name string `json:"name"`

	// URL is the full URL for the ACE Hub instance
	URL string `json:"url"`
}

// BottleSpec describes a bottle as used in ACE Hub.
type BottleSpec struct {
	// Name is the display name to use for the bottle
	Name string `json:"name" query:"name"`

	// Bottle is the bottle reference for the bottle (OCI reference or bottleID)
	Bottle string `json:"bottle" query:"bottle"`

	// Selector to use for selecting parts of a bottle.  Different selectors are separated by "|".
	Selector string `json:"selector" query:"selector"`

	// IPS is the image pull secret to use for pulling bottles where authz is required
	IPS string `json:"ips" query:"ips"`
}

// ACEHubLaunchTemplate is the launch specification for ACE Hub.  It is the same as the PodGroupOptions from ace/hub/api.
// This struct also supports marshalling to URL query parameters with "github.com/hetiansu5/urlquery".
type ACEHubLaunchTemplate struct {
	// HubName is the name of the pod group
	HubName string `json:"hubName,omitempty" query:"hubName"`

	// Image is the OCI image
	Image string `json:"image,omitempty" query:"image"`

	// Replicas is the number of pods to launch
	Replicas uint `json:"replicas,omitempty" query:"replicas"`

	// Resources is the number of resources to use (e.g., {"cpu":"4", "memory":"8Gi"})
	Resources map[string]string `json:"resources,omitempty" query:"resources"`

	// Shm is the amount of shared memory to use (e.g., "1Gi")
	Shm string `json:"shm,omitempty" query:"shm"`

	// Env
	Env map[string]string `json:"env,omitempty" query:"env"`

	// ProxyType straight or normal
	ProxyType string `json:"proxyType,omitempty" query:"proxyType"`

	// Jupyter is the flag that enables the jupyter hack to support vanilla jupyter images from jupyter-stack
	Jupyter bool `json:"jupyter,omitempty" query:"jupyter"`

	// JetBrains enables the JetBrains hack
	JetBrains bool `json:"jetBrains,omitempty" query:"jetBrains"`

	// StartScript is a script (the string is the body of the script) that will run on startup in the main container
	StartScript string `json:"startScript,omitempty" query:"startScript"`

	// Bottles to mount into the pod
	Bottles []BottleSpec `json:"bottles,omitempty" query:"bottles"`
}

// ViewerSpec defines how to launch something in an environment.
type ViewerSpec struct {
	// Name is the name of the view that will be presented to the user
	Name string `json:"name"`

	// Accept is the HTTP accept formated string denoting what media types (with priorities) that will be accepted by this viewer.
	Accept string `json:"accept"`

	// ACEHub is the launch template to be launch the viewer
	ACEHub ACEHubLaunchTemplate `json:"acehub"`
}

// LogValue implements slog.LogValuer.
func (d Database) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("dsn", d.DSN),
	)
}

// LogValue implements slog.LogValuer.
func (c ServerConfigurationSpec) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("db", c.DB),
		slog.Any("webapp", c.WebApp),
	)
}

// LogValue implements slog.LogValuer.
func (c ServerConfiguration) LogValue() slog.Value {
	return c.ServerConfigurationSpec.LogValue()
}

// SampleServerConfig is a sample ServerConfiguration snippet.
const SampleServerConfig = `# ACE Data Telemetry Server Configuration
apiVersion: config.telemetry.act3-ace.io/v1alpha1
kind: ServerConfiguration

db:
  # To use SQLite
  # dsn: file:test.db

  # To use PostgreSQL
  # dsn: "postgres://tester:myPassword@localhost/test"

webapp:
  # path to the jupyter executable
  jupyter: /home/user/env/bin/jupyter
  
  # ACE Hub instances that we can use to display data via the viewers (below)
  acehubs:
  - name: Lion
    url:  https://hub.lion.act3-ace.ai
  - name: GCP
    url:  https://hub.ace.afresearchlab.com

  # Viewer specifications tell ACE Hub how to display an artifact that matches by "Accept"
  viewers:
  - name: "VS Code"
    accept: "image/*,application/json,text/plain;q=0.5, application/vnd.act3-ace.bottle;q=0.9"
    acehub:
      image: reg.git.act3-ace.com/ace/hub/vscode-server
      resources:
        cpu: "2"
        memory: "2Gi"
      proxyType: normal
      jupyter: false

  defaultBottleSelectors:
   - type != testing
   - foo!=bar
`
