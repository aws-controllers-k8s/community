#!/usr/bin/env bash

# A script that creates the Operator Lifecycle Manager bundle for a
# specific ACK service controller

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
BUILD_DIR="$ROOT_DIR/build"
DEFAULT_BUNDLE_CHANNEL="alpha"
DEFAULT_SERVICE_CONTROLLER_CONTAINER_REPOSITORY="public.ecr.aws/aws-controllers-k8s/controller"
PRESERVE=${PRESERVE:-"false"}

export DOCKER_BUILDKIT=${DOCKER_BUILDKIT:-1}

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed uuidgen
check_is_installed kustomize "You can install kustomize with the helper scripts/install-kustomize.sh"
check_is_installed operator-sdk "You can install Operator SDK with the helmer scripts/install-operator-sdk.sh"

function clean_up {
    if [[ "$PRESERVE" == false ]]; then
        rm -r "$TMP_DIR" || :
        return
    fi
    echo "--"
    echo "To regenerate bundle with the same kustomize configs, re-run with:
    \"TMP_DIR=$TMP_DIR\""
}

USAGE="
Usage:
  $(basename "$0") <service> <version>

<service> should be an AWS service API aliases that you wish to build -- e.g.
's3' 'sns' or 'sqs'

Environment variables:
  BUNDLE_DEFAULT_CHANNEL                    The default channel to publish the OLM bundle to.
                                            Default: $DEFAULT_BUNDLE_CHANNEL
  BUNDLE_VERSION                            The semantic version of the bundle should it need
                                            to differ from the version of the controller which is
                                            passed into the script. Optional. If unset, this
                                            value will match that passed in by the user as the
                                            <version> parameter (i.e. the controller version)
  BUNDLE_CHANNELS                           A comma-separated list of channels the bundle belongs
                                            to (e.g. \"alpha,beta\").
                                            Default: $DEFAULT_BUNDLE_CHANNEL
  BUNDLE_OUTPUT_PATH:                       Specify a path for the OLM bundle to output to.
                                            Default: {SERVICE_CONTROLLER_SOURCE_PATH}/olm
  AWS_SERVICE_DOCKER_IMG:                   Specify the docker image for the AWS service.
                                            This should include the registry, namespace,
                                            image name, and tag. Takes precedence over
                                            SERVICE_CONTROLLER_CONTAINER_REPOSITORY
  SERVICE_CONTROLLER_SOURCE_PATH:           Path to the service controller source code
                                            repository.
                                            Default: ../{SERVICE}-controller
  SERVICE_CONTROLLER_CONTAINER_REPOSITORY   The container repository where the controller exists.
                                            This should include the registry, namespace, and
                                            image name.
                                            Default: $DEFAULT_SERVICE_CONTROLLER_CONTAINER_REPOSITORY
  TMP_DIR                                   Directory where kustomize assets will be temporarily
                                            copied before they are modified and passed to bundle
                                            generation logic.
                                            Default: $BUILD_DIR/tmp-olm-{RANDOMSTRING}
  PRESERVE:                                 Preserves modified kustomize configs for
                                            inspection. (<true|false>)
                                            Default: false
  BUNDLE_GENERATE_EXTRA_ARGS                Extra arguments to pass into the command
                                            'operator-sdk generate bundle'.
"

if [ $# -ne 2 ]; then
    echo "ERROR: $(basename "$0") accepts two parameters, the SERVICE and VERSION" 1>&2
    echo "$USAGE"
    exit 1
fi

SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')
VERSION=$2
BUNDLE_VERSION=${BUNDLE_VERSION:-$VERSION}

DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH="$ROOT_DIR/../$SERVICE-controller"
SERVICE_CONTROLLER_SOURCE_PATH=${SERVICE_CONTROLLER_SOURCE_PATH:-$DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH}

BUNDLE_OUTPUT_PATH=${BUNDLE_OUTPUT_PATH:-$SERVICE_CONTROLLER_SOURCE_PATH/olm}

if [ -z "$TMP_DIR" ]; then
    TMP_ID=$(uuidgen | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
    TMP_DIR=$BUILD_DIR/tmp-olm-$TMP_ID
fi
# if TMP_DIR is provided but doesn't exist, we still use
# it and create it later.
tmp_kustomize_config_dir="$TMP_DIR/config"

if [[ ! -d $SERVICE_CONTROLLER_SOURCE_PATH ]]; then
    echo "Error evaluating SERVICE_CONTROLLER_SOURCE_PATH environment variable:" 1>&2
    echo "$SERVICE_CONTROLLER_SOURCE_PATH is not a directory." 1>&2
    echo "${USAGE}"
    exit 1
fi

# Set controller image.
if [ -n "$AWS_SERVICE_DOCKER_IMG" ] && [ -n "$SERVICE_CONTROLLER_CONTAINER_REPOSITORY" ] ; then
  # It's possible to set the repository (i.e. everything except the tag) as well as the
  # entire path including the tag using AWS_SERVIC_DOCKER_IMG. If the latter is set, it
  # will take precedence, so inform the user that this will happen in case the usage of
  # each configurable is unclear before runtime. 
  echo "both AWS_SERVICE_DOCKER_IMG and SERVICE_CONTROLLER_CONTAINER REPOSITORY are set."
  echo "AWS_SERVICE_DOCKER_IMG is expected to be more complete and will take precedence."
  echo "Ignoring SERVICE_CONTROLLER_CONTAINER_REPOSITORY."
fi

SERVICE_CONTROLLER_CONTAINER_REPOSITORY=${SERVICE_CONTROLLER_CONTAINER_REPOSITORY:-$DEFAULT_SERVICE_CONTROLLER_CONTAINER_REPOSITORY}
DEFAULT_AWS_SERVICE_DOCKER_IMAGE="$SERVICE_CONTROLLER_CONTAINER_REPOSITORY:$SERVICE-$VERSION"
AWS_SERVICE_DOCKER_IMG=${AWS_SERVICE_DOCKER_IMG:-$DEFAULT_AWS_SERVICE_DOCKER_IMAGE}

trap "clean_up" EXIT

# prepare the temporary config dir
mkdir -p $TMP_DIR
cp -a $SERVICE_CONTROLLER_SOURCE_PATH/config $TMP_DIR
pushd $tmp_kustomize_config_dir/controller 1>/dev/null
kustomize edit set image controller="$AWS_SERVICE_DOCKER_IMG"
popd 1>/dev/null

# prepare bundle generate arguments
opsdk_gen_bundle_args="--version $BUNDLE_VERSION --package ack-$SERVICE-controller --kustomize-dir $SERVICE_CONTROLLER_SOURCE_PATH/config/manifests --overwrite "

# specify default channel and all channels if not specified by user
BUNDLE_DEFAULT_CHANNEL=${BUNDLE_DEFAULT_CHANNEL:-$DEFAULT_BUNDLE_CHANNEL}
BUNDLE_CHANNELS=${BUNDLE_CHANNELS:-$DEFAULT_BUNDLE_CHANNEL}

opsdk_gen_bundle_args="$opsdk_gen_bundle_args --default-channel $DEFAULT_BUNDLE_CHANNEL --channels $BUNDLE_CHANNELS"
if [ -n "$BUNDLE_GENERATE_EXTRA_ARGS" ]; then
    opsdk_gen_bundle_args="$opsdk_gen_bundle_args $BUNDLE_GENERATE_EXTRA_ARGS"
fi

# operator-sdk generate bundle creates a bundle.Dockerfile relative
# to where it's called and it cannot be changed as of right now.
# For the time being, keep the bundle.Dockerfile local to the actual
# bundle assets.
# TODO(): determine if it makes sense to keep the bundle.Dockerfiles
# in the controller-specific repositories.
mkdir -p $BUNDLE_OUTPUT_PATH
pushd $BUNDLE_OUTPUT_PATH 1> /dev/null
kustomize build $tmp_kustomize_config_dir/manifests | operator-sdk generate bundle $opsdk_gen_bundle_args 
popd 1> /dev/null

operator-sdk bundle validate $BUNDLE_OUTPUT_PATH/bundle