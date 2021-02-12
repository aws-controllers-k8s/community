#!/usr/bin/env bash

# A script that builds Helm packages for specified service controller that have a
# Helm chart, builds the Helm repository index.yaml and git commits and pushes
# the updated Helm packages and repository index to the gh-pages branch of the
# upstream source repository.

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
DEFAULT_HELM_REGISTRY="public.ecr.aws/aws-controllers-k8s"
DEFAULT_HELM_REPO="chart"
DEFAULT_RELEASE_VERSION="unknown"

HELM_REGISTRY=${HELM_REGISTRY:-$DEFAULT_HELM_REGISTRY}
HELM_REPO=${HELM_REPO:-$DEFAULT_HELM_REPO}

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed helm "You can install Helm with the helper scripts/install-helm.sh"

USAGE="
Usage:
  $(basename "$0") <AWS_SERVICE>

<AWS_SERVICE> AWS Service name (ecr, sns, sqs)

Environment variables:
  SERVICE_CONTROLLER_SOURCE_PATH:       Path to the service controller source code
                                        repository.
                                        Default: ../{AWS_SERVICE}-controller
  RELEASE_VERSION:                      The semver release version to use.
                                        Default: $DEFAULT_RELEASE_VERSION
  HELM_REGISTRY:                        The name of the Helm registry.
                                        Default: $DEFAULT_HELM_REGISTRY
  HELM_REPO:                            The name of the Helm repository.
                                        Default: $DEFAULT_HELM_REPO
"

if [ $# -ne 1 ]; then
    echo "AWS_SERVICE is not defined. Script accepts one parameter, <AWS_SERVICE> to build Helm packages for that service" 1>&2
    echo "${USAGE}"
    exit 1
fi

AWS_SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')

# Source code for the controller will be in a separate repo, typically in
# $GOPATH/src/github.com/aws-controllers-k8s/$AWS_SERVICE-controller/
DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH="$ROOT_DIR/../$AWS_SERVICE-controller"
SERVICE_CONTROLLER_SOURCE_PATH=${SERVICE_CONTROLLER_SOURCE_PATH:-$DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH}

if [[ ! -d $SERVICE_CONTROLLER_SOURCE_PATH ]]; then
    echo "Error evaluating SERVICE_CONTROLLER_SOURCE_PATH environment variable:" 1>&2
    echo "$SERVICE_CONTROLLER_SOURCE_PATH is not a directory." 1>&2
    echo "${USAGE}"
    exit 1
fi

pushd "$SERVICE_CONTROLLER_SOURCE_PATH" 1>/dev/null
DEFAULT_RELEASE_VERSION=$(git describe --tags --always --dirty || echo "unknown")
RELEASE_VERSION=${RELEASE_VERSION:-$DEFAULT_RELEASE_VERSION}
popd 1>/dev/null

export HELM_EXPERIMENTAL_OCI=1

if [[ -d "$SERVICE_CONTROLLER_SOURCE_PATH/helm" ]]; then
    echo -n "Generating Helm chart package for $AWS_SERVICE@$RELEASE_VERSION ... "
    helm chart save $SERVICE_CONTROLLER_SOURCE_PATH/helm/ $HELM_REGISTRY/$HELM_REPO:$AWS_SERVICE-$RELEASE_VERSION
    echo "ok."
    helm chart push $HELM_REGISTRY/$HELM_REPO:$AWS_SERVICE-$RELEASE_VERSION
else
    echo "Error building Helm packages:" 1>&2
    echo "$SERVICE_CONTROLLER_SOURCE_PATH/helm is not a directory." 1>&2
    echo "${USAGE}"
    exit 1
fi
