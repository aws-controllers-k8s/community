#!/usr/bin/env bash

# A script that creates a Helm chart package for a specific ACK service
# controller

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
BUILD_DIR="$ROOT_DIR/build"

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed helm "You can install Helm with the helper scripts/install-helm.sh"

USAGE="
Usage:
  $(basename "$0") <service>

<service> should be an AWS service API aliases that you wish to build -- e.g.
's3' 'sns' or 'sqs'

Environment variables:
  CHART_INPUT_PATH:         Specify a path for the Helm chart to use as input.
                            Default: services/{SERVICE}/helm
  PACKAGE_OUTPUT_PATH:      Specify a path for the Helm chart package to output to.
                            Default: $BUILD_DIR/release/{SERVICE}
"

if [ $# -ne 1 ]; then
    echo "ERROR: $(basename "$0") accepts one parameter, the SERVICE" 1>&2
    echo "$USAGE"
    exit 1
fi

SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')

PACKAGE_OUTPUT_PATH=${PACKAGE_OUTPUT_PATH:-"$BUILD_DIR/release/$SERVICE"}
CHART_INPUT_PATH=${CHART_INPUT_PATH:-"$ROOT_DIR/services/$SERVICE/helm"}

if [[ ! -d "$CHART_INPUT_PATH" ]]; then
    echo "Chart input path: $CHART_INPUT_PATH does not exist."
    exit 1
fi

mkdir -p $PACKAGE_OUTPUT_PATH

helm package $CHART_INPUT_PATH --destination $PACKAGE_OUTPUT_PATH
