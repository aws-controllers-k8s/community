#!/usr/bin/env bash

# A script that builds Helm packages for all service controllers that have a
# Helm chart, builds the Helm repository index.yaml and git commits and pushes
# the updated Helm packages and repository index to the gh-pages branch of the
# upstream source repository.

set -eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
SERVICES_DIR="$ROOT_DIR/services"
DEFAULT_HELM_REGISTRY="public.ecr.aws/aws-controllers-k8s"
DEFAULT_HELM_REPO="chart"
DEFAULT_RELEASE_VERSION=$(git describe --tags --always --dirty || echo "unknown")


RELEASE_VERSION=${RELEASE_VERSION:-$DEFAULT_RELEASE_VERSION}
HELM_REGISTRY=${HELM_REGISTRY:-$DEFAULT_HELM_REGISTRY}
HELM_REPO=${HELM_REPO:-$DEFAULT_HELM_REPO}

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed helm "You can install Helm with the helper scripts/install-helm.sh"

USAGE="
Usage:
  $(basename "$0")

Environment variables:
  RELEASE_VERSION:          The semver release version to use.
                            Default: $DEFAULT_RELEASE_VERSION
  HELM_REGISTRY:            The name of the Helm registry.
                            Default: $DEFAULT_HELM_REGISTRY
  HELM_REPO:                The name of the Helm repository.
                            Default: $DEFAULT_HELM_REPO
"

export HELM_EXPERIMENTAL_OCI=1

for SERVICE_DIR in $SERVICES_DIR/*; do
    SERVICE=$( basename $SERVICE_DIR)
    if [[ -d "$SERVICES_DIR/$SERVICE/helm" ]]; then
        echo -n "Generating Helm chart package for $SERVICE@$RELEASE_VERSION ... "
        helm chart save $SERVICES_DIR/$SERVICE/helm/ $HELM_REGISTRY/$HELM_REPO:$SERVICE-$RELEASE_VERSION
        echo "ok."
        helm chart push $HELM_REGISTRY/$HELM_REPO:$SERVICE-$RELEASE_VERSION
    fi
done
