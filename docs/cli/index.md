---
title: telemetry
description: ACE Data Bottle Telemetry Server
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# telemetry

ACE Data Bottle Telemetry Server

## Synopsis

Manages the discovery and advanced queries on metadata on bottles

## Options

```plaintext
Options:
      --config stringArray         server configuration file location (setable with env "ACE_TELEMETRY_CONFIG"). 
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-telemetry-config.yaml,/root/.config/ace/telemetry/config.yaml,/etc/ace/telemetry/config.yaml])
  -h, --help                       help for telemetry
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_TELEMETRY_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```

## Subcommands

- [`telemetry client`](client/index.md) - Client commands for interacting with a telemetry server at a low level.
- [`telemetry completion`](completion/index.md) - Generate the autocompletion script for the specified shell
- [`telemetry config`](config.md) - Show the current configuration
- [`telemetry filter`](filter/index.md) - Filters to use when pretty printing logs
- [`telemetry gendocs`](gendocs/index.md) - Generate documentation for the tool in various formats
- [`telemetry serve`](serve.md) - Start the server
- [`telemetry template`](template.md) - Template data in the given directory
- [`telemetry version`](version.md) - Print the version
