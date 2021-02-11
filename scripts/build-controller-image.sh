#!/usr/bin/env bash

set -eo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SCRIPTS_DIR=$DIR
ROOT_DIR=$DIR/..
DOCKERFILE_PATH=$ROOT_DIR/Dockerfile
ACK_DIR=$ROOT_DIR/..
DOCKERFILE=${DOCKERFILE:-"$DOCKERFILE_PATH"}
LOCAL_MODULES=${LOCAL_MODULES:-"false"}
BUILD_DATE=$(date +%Y-%m-%dT%H:%M)
QUIET=${QUIET:-"false"}

export DOCKER_BUILDKIT=${DOCKER_BUILDKIT:-1}

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

USAGE="
Usage:
  $(basename "$0") <aws_service>

Builds the Docker image for an ACK service controller. 

Example: $(basename "$0") ecr

<aws_service> should be an AWS Service name (ecr, sns, sqs, petstore, bookstore)

Environment variables:
  QUIET:                            Build controller container image quietly (<true|false>)
                                    Default: false
  LOCAL_MODULES:                    Enables use of local modules during AWS Service controller docker image build
                                    Default: false
  AWS_SERVICE_DOCKER_IMG:           Controller container image tag
                                    Default: aws-controllers-k8s:\$AWS_SERVICE-\$VERSION
  SERVICE_CONTROLLER_SOURCE_PATH:   Directory to find the service controller to build an image for.
                                    Default: ../\$AWS_SERVICE-controller
"

if [ $# -ne 1 ]; then
    echo "AWS_SERVICE is not defined. Script accepts one parameter, the <AWS_SERVICE> to build a container image of that service" 1>&2
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

pushd $SERVICE_CONTROLLER_SOURCE_PATH 1>/dev/null

SERVICE_CONTROLLER_GIT_VERSION=`git describe --tags --always --dirty || echo "unknown"`
SERVICE_CONTROLLER_GIT_COMMIT=`git rev-parse HEAD`

popd 1>/dev/null

DEFAULT_AWS_SERVICE_DOCKER_IMG="aws-controllers-k8s:$AWS_SERVICE-$SERVICE_CONTROLLER_GIT_VERSION"
AWS_SERVICE_DOCKER_IMG=${AWS_SERVICE_DOCKER_IMG:-"$DEFAULT_AWS_SERVICE_DOCKER_IMG"}

if [[ $QUIET = "false" ]]; then
    echo "building '$AWS_SERVICE' controller docker image with tag: ${AWS_SERVICE_DOCKER_IMG}"
    echo " git commit: $SERVICE_CONTROLLER_GIT_COMMIT"
fi

# if local build
# then use Dockerfile which allows references to local modules from service controller
DOCKER_BUILD_CONTEXT="$ACK_DIR"
if [[ "$LOCAL_MODULES" = "true" ]]; then
  DOCKERFILE="${ROOT_DIR}"/Dockerfile.local
fi

docker build \
  --quiet=${QUIET} \
  -t "${AWS_SERVICE_DOCKER_IMG}" \
  -f "${DOCKERFILE}" \
  --build-arg service_alias=${AWS_SERVICE} \
  --build-arg service_controller_git_version="$SERVICE_CONTROLLER_GIT_VERSION" \
  --build-arg service_controller_git_commit="$SERVICE_CONTROLLER_GIT_COMMIT" \
  --build-arg build_date="$BUILD_DATE" \
  "${DOCKER_BUILD_CONTEXT}"

if [ $? -ne 0 ]; then
  exit 2
fi
