package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Public users of telemetry will not be able to utilize any configuration options defined in this file.
//
// Copied from git.act3-ace.com/ace/hub/api as of commit d56eb09953c34a555b7d15bdbfa05abee33ddfe6.
// BottleSpec already existed, however copying the remaining structs allow us to
// not rely on a private dependency.

// BottleSpec describes the bottle to attach.
type BottleSpec struct {
	// Name is the display name to use for the bottle within the container
	Name string `json:"name" query:"name" yaml:"name"`

	// Bottle is the bottle reference for the bottle (OCI reference or bottleID)
	BottleRef string `json:"bottleRef" query:"bottleRef" yaml:"bottleRef"`

	// Selector to use for selecting parts of a bottle.  Different selectors are separated by "|".
	// added omitempty for launchtemplates
	//+optional
	Selector []string `json:"selector,omitempty" query:"selector" validate:"option" yaml:"selector"`

	// IPS is the image pull secret to use for pulling bottles where authz is required
	// added omitempty for launchtemplates
	//+optional
	IPS string `json:"ips,omitempty" query:"ips" validate:"option" yaml:"ips"`
}

// GPU defines the GPUs desired for an environment.
type GPU struct {
	// Type is the name of the GPU from the configuration
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Count is the number of GPUs
	Count int `json:"count,omitempty" yaml:"count,omitempty"`
}

// HubEnvTemplateSpec defines the desired state of HubEnvTemplate.
type HubEnvTemplateSpec struct {

	// ServiceAccountName is the name of the service account to use for the pod.  This is also useful for injecting image pull secrets.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=128
	ServiceAccountName string `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`

	// EnvSecretPrefix is the prefix to use for names of secrets adding environment variables to the pods.
	// If the secret <envSecretprefix>-env is present then those environment variables are added to the pod.
	// If the secret <envSecretPrefix>-envfile is present then the keys are the variable names and the values are the path to a file in the pod with the content provided in the secret.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=64
	EnvSecretPrefix string `json:"envSecretPrefix,omitempty" yaml:"envSecretPrefix,omitempty"`

	// QueueName is the name fo the queue to submit this workload to.  This can be changed at runtime without restarting the pod.
	// +kubebuilder:validation:Optional
	QueueName string `json:"queueName,omitempty" yaml:"queueName,omitempty"`

	// GPU the GPU specification
	// +kubebuilder:validation:Optional
	GPU *GPU `json:"gpu,omitempty" yaml:"gpu,omitempty"`

	// Resources is the compute resources (e.g., cpu, memory)
	// +kubebuilder:validation:Optional
	Resources v1.ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`

	// Image is the image to use for the pod
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	Image string `json:"image" yaml:"image"`

	// Env is extra environment variables.  These must not hold credentials.  To add credentials to environment variables please use envSecretPrefix.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxProperties=32
	Env map[string]string `json:"env,omitempty" yaml:"env,omitempty"`

	// Script is a script (the string is the body of the, often multi-line, script).  This is often a long running script (does not terminate until work is completed or never in the case of a service).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=1024
	Script string `json:"script,omitempty" yaml:"script,omitempty"`

	// SharedMemory is the amount of shared memory for the pods.
	// +kubebuilder:validation:Optional
	SharedMemory *resource.Quantity `json:"shm,omitempty" yaml:"shm,omitempty"`

	// Bottles to mount into the pod
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxItems=16
	Bottles []BottleSpec `json:"bottles,omitempty" yaml:"bottles,omitempty"`

	// Ports is a list of exposed ports.  If the list is non-nil then this is considered an interactive hubenv.  If nil then it is a non-interactive/batch workload.  This can be changed at runtime without restarting the pod so long as the type of workload (interactive vs non-interactive) does not change.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxItems=16
	Ports []Port `json:"ports,omitempty" yaml:"ports,omitempty"`
}

// ProxyTypes
const (
	// ProxyTypeStraight is for services that do not support URL rewriting.  This includes Jupyter where they use absolute paths for data.
	// The service needs to know the relative path prefix used to access the pod, the environment variable $ACE_URL_PREFIX can be utilized.
	ProxyTypeStraight = "straight"

	// ProxyTypeNormal is used when a service has no absolute paths and thus supports being proxied through a URL re-writing HTTP reverse proxy.
	ProxyTypeNormal = "normal"

	// ProxyTypeSubdoman is used to create a subdomain for the hubenv
	ProxyTypeSubdomain = "subdomain"
)

// Port describes a listening HTTP server for access to services within the pod
type Port struct {
	// Name is used to describe the usage of the port (e.g., tensorboard, vscode, analytics)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])$`
	Name string `json:"name" yaml:"name"`

	// Number is the port number
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Number int `json:"number" yaml:"number"`

	// Protocol string // TCP vs HTTP vs HTTPS

	// ProxyType is the type of proxying needed to access this service.  "Straight" does no URL rewriting, while "normal" rewrite the URL so requests come in at the root level.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"straight", "normal", "subdomain"}
	ProxyType string `json:"proxyType" yaml:"proxyType"`
}
