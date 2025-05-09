package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "config.telemetry.act3-ace.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&ServerConfiguration{},
		&ClientConfiguration{},
	)
	scheme.AddTypeDefaultingFunc(&ServerConfiguration{}, func(in any) { ServerConfigurationDefault(in.(*ServerConfiguration)) })
	scheme.AddTypeDefaultingFunc(&ClientConfiguration{}, func(in any) { ClientConfigurationDefault(in.(*ClientConfiguration)) })
	return nil
}
