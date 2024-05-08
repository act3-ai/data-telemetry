package v1alpha1

// ServerConfigurationDefault defaults the configuration values.
func ServerConfigurationDefault(obj *ServerConfiguration) {
	// This is called after we decode the values (from file) so we need to be careful not to overwrite values that are already set.
	// We can use pointers if we need to know that a value has been set or not.

	// These might not be set in some cases
	obj.APIVersion = GroupVersion.String()
	obj.Kind = "ServerConfiguration"

	if obj.DB.DSN == "" {
		obj.DB.DSN = "file:test.db"
	}
}

// ClientConfigurationDefault defaults the configuration values.
func ClientConfigurationDefault(obj *ClientConfiguration) {
	// This is called after we decode the values (from file) so we need to be careful not to overwrite values that are already set.
	// We can use pointers if we need to know that a value has been set or not.

	// These might not be set in some cases
	obj.APIVersion = GroupVersion.String()
	obj.Kind = "ClientConfiguration"
}
