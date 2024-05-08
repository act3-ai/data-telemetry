## telemetry

ACE Data Bottle Telemetry Server

### Synopsis

Manages the discovery and advanced queries on metadata on bottles

### Options

```
      --config stringArray   server configuration file location (setable with env "ACE_TELEMETRY_CONFIG"). 
                             The first configuration file present is used.  Others are ignored.
                              (default [ace-telemetry-config.yaml,HOMEDIR/.config/ace/telemetry/config.yaml,/etc/ace/telemetry/config.yaml])
  -h, --help                 help for telemetry
  -v, --verbosity int[=0]    Logging verbosity level (also setable with environment variable ACE_TELEMETRY_VERBOSITY)
```

### SEE ALSO

* [telemetry client](telemetry_client.md)	 - Client commands for interacting with a telemetry server at a low level.
* [telemetry completion](telemetry_completion.md)	 - Generate the autocompletion script for the specified shell
* [telemetry config](telemetry_config.md)	 - Show the current configuration
* [telemetry filter](telemetry_filter.md)	 - Filters to use when pretty printing logs
* [telemetry serve](telemetry_serve.md)	 - Start the server
* [telemetry template](telemetry_template.md)	 - Template data in the given directory
* [telemetry version](telemetry_version.md)	 - Print the version

