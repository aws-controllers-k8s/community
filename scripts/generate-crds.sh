#!/usr/bin/env bash

set -Eeuxo pipefail
trap 'on_error $LINENO' ERR

echo "Generating CRDs"

__service_types_path=$1
__output_artifacts=$2

which controller-gen
controller-gen "crd:trivialVersions=true" paths="$__service_types_path"/... output:crd:artifacts:config="$__output_artifacts"
