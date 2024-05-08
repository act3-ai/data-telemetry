#!/usr/bin/env bash

set -e

ver=$1
# export so yq can access it by name (internally)
export ver

echo "$ver" > VERSION

yq e '(.version = env(ver)) | (.appVersion = env(ver))' -i charts/telemetry/Chart.yaml
yq e '.image.tag = "v" + env(ver)' -i charts/telemetry/values.yaml
