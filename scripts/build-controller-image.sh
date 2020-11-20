#!/usr/bin/env bash

set -E

DIR=$(cd "$(dirname "$0")"; pwd)
SCRIPTS_DIR=$DIR
DOCKERFILE_PATH=$DIR/../Dockerfile
BUILD_CONTEXT=$DIR/..
OPTIND=1
VERSION=$(git describe --tags --always --dirty || echo "unknown")
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
  QUIET:                    Build controller container image quietly (<true|false>) 
                            Default: false
  AWS_SERVICE_DOCKER_IMG:   Controller container image tag 
                            Default: aws-controllers-k8s:$AWS_SERVICE-$VERSION
"

if [ $# -ne 1 ]; then
    echo "AWS_SERVICE is not defined. Script accepts one parameter, the <AWS_SERVICE> to build a container image of that service" 1>&2
    echo "${USAGE}"
    exit 1
fi

AWS_SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')
QUIET=${QUIET:-"false"}
DEFAULT_AWS_SERVICE_DOCKER_IMG="aws-controllers-k8s:$AWS_SERVICE-$VERSION"
: "${AWS_SERVICE_DOCKER_IMG:="$DEFAULT_AWS_SERVICE_DOCKER_IMG"}"
: "${DOCKERFILE:="$DOCKERFILE_PATH"}"

if [[ $QUIET = "false" ]]; then
    echo "building '$AWS_SERVICE' controller docker image with tag: ${AWS_SERVICE_DOCKER_IMG}"
fi

docker build \
  --quiet=${QUIET} \
  -t ${AWS_SERVICE_DOCKER_IMG} \
  -f ${DOCKERFILE} \
  --build-arg service_alias=${AWS_SERVICE} \
  ${BUILD_CONTEXT}

if [ $? -ne 0 ]; then
  exit 2
fi
