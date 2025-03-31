# API Reference

## Packages
- [config.telemetry.act3-ace.io/v1alpha1](#configtelemetryact3-aceiov1alpha1)


## config.telemetry.act3-ace.io/v1alpha1

Package v1alpha1 contains API schema definitions for managing Telemetry configuration.  Both client and server configuration are included in this group.

### Resource Types
- [ClientConfiguration](#clientconfiguration)
- [ServerConfiguration](#serverconfiguration)



#### ACEHubInstance



ACEHubInstance is an existing instance of ACE Hub that will be offered as a bottle viewer engine.



_Appears in:_
- [WebApp](#webapp)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the instance for display purposes |  |  |
| `url` _string_ | URL is the full URL for the ACE Hub instance |  |  |




#### ClientConfiguration



ClientConfiguration is configuration for making requests from the upload and download subcommands.  Not used by the serve command.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `config.telemetry.act3-ace.io/v1alpha1` | | |
| `kind` _string_ | `ClientConfiguration` | | |
| `locations` _[Location](#location) array_ | Locations is the list of Telemetry server locations.  Data will be pushed to all and pulled from all |  |  |


#### ClientConfigurationSpec



ClientConfigurationSpec is the actual configuration values.



_Appears in:_
- [ClientConfiguration](#clientconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `locations` _[Location](#location) array_ | Locations is the list of Telemetry server locations.  Data will be pushed to all and pulled from all |  |  |


#### Database



Database is configuration for the database connection.



_Appears in:_
- [ServerConfiguration](#serverconfiguration)
- [ServerConfigurationSpec](#serverconfigurationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dsn` _[SecretURL](#secreturl)_ | DSN is the database connection string |  |  |
| `password` _[Secret](#secret)_ | Password is the database account password |  |  |


#### Location



Location is a target location and the specific information needed for authentication for that location.



_Appears in:_
- [ClientConfiguration](#clientconfiguration)
- [ClientConfigurationSpec](#clientconfigurationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the display name of the telemetry server |  |  |
| `url` _[SecretURL](#secreturl)_ | URL is the base URL for the telemetry server (does not include the /api) |  |  |
| `oauth` _[OAuthProvider](#oauthprovider)_ | OAuth defines an OAuth2.0 provider used for authentication. |  |  |
| `cookies` _object (keys:string, values:[Secret](#secret))_ | Cookies to use for authentication |  |  |
| `token` _[Secret](#secret)_ | Bearer token to use for authentication |  |  |


#### OAuthProvider



OAuthProvider defines a host and client application ID used for OAuth2.0 Device Grant authentication
defined by RFC 8628; see https://www.rfc-editor.org/rfc/rfc8628.



_Appears in:_
- [Location](#location)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `issuer` _string_ | Issuer defines the authorization server. |  |  |
| `clientID` _string_ | ClientID is the client application identifier. Not a secret.<br />See https://www.rfc-editor.org/rfc/rfc6749#section-2.2 for more info. |  |  |


#### ServerConfiguration



ServerConfiguration is the Schema for the Telemetry Server Configurations API.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `config.telemetry.act3-ace.io/v1alpha1` | | |
| `kind` _string_ | `ServerConfiguration` | | |
| `db` _[Database](#database)_ | DB is the database configuration |  |  |
| `webapp` _[WebApp](#webapp)_ | WebApp specific configuration |  |  |


#### ServerConfigurationSpec



ServerConfigurationSpec is the actual configuration values.



_Appears in:_
- [ServerConfiguration](#serverconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `db` _[Database](#database)_ | DB is the database configuration |  |  |
| `webapp` _[WebApp](#webapp)_ | WebApp specific configuration |  |  |


#### ViewerSpec



ViewerSpec defines how to launch something in an environment.



_Appears in:_
- [WebApp](#webapp)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the view that will be presented to the user |  |  |
| `accept` _string_ | Accept is the HTTP accept formated string denoting what media types (with priorities) that will be accepted by this viewer. |  |  |
| `acehub` _[HubEnvTemplateSpec](#hubenvtemplatespec)_ | ACEHub is the launch template to be launch the viewer |  |  |


#### WebApp



WebApp is the configuration for the telemetry web application.



_Appears in:_
- [ServerConfiguration](#serverconfiguration)
- [ServerConfigurationSpec](#serverconfigurationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `acehubs` _[ACEHubInstance](#acehubinstance) array_ | ACEHubs is a list of ace hub instances that will be available to users of the web application for viewing bottles |  |  |
| `viewers` _[ViewerSpec](#viewerspec) array_ | Viewers is a list of viewer specifications for how to launch the viewer |  |  |
| `jupyter` _string_ | JupyterExecutable is the path to the "jupyter" executable |  |  |
| `defaultBottleSelectors` _string array_ | DefaultBottleSelectors is the list of selectors to use when searching the catalog by default |  |  |
| `assets` _string_ | AssetDir is the directory where the web assets reside, default value is "assets" |  |  |


