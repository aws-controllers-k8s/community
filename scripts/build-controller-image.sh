#!/usr/bin/env bash

set -E

DIR=$(cd "$(dirname "$0")"; pwd)
SCRIPTS_DIR=$DIR
DOCKERFILE_PATH=$DIR/../Dockerfile
BUILD_CONTEXT=$DIR/..
QUIET=false
OPTIND=1
VERSION=$(git describe --tags --always --dirty || echo "unknown")
export DOCKER_BUILDKIT=${DOCKER_BUILDKIT:-1}

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

USAGE="
Usage:
  $(basename "$0") [-q] [-s <AWS_SERVICE>] [-i <Docker image tag>]

Builds the Docker image for an ACK service controller.

Example: $(basename "$0") -q -s ecr

Options:
  -s          Provide AWS Service name (ecr, sns, sqs, petstore, bookstore)
  -i          Controller container image tag (Default: ack-<service>-controller:$VERSION)
  -q          Build controller container image quietly
"

# Process our input arguments
while getopts "qs:i:" opt; do
  case ${opt} in
    q ) # Build image quietly
        QUIET=true
      ;;
    s ) # AWS Service name
        AWS_SERVICE=$(echo "${OPTARG}" | tr '[:upper:]' '[:lower:]')
      ;;
    i ) # Controller image tag
        AWS_SERVICE_DOCKER_IMG="${OPTARG}"
      ;;
    \? )
        echo "${USAGE}" 1>&2
        exit
      ;;
  esac
done

if [ -z "$AWS_SERVICE" ]; then
  echo "AWS_SERVICE is not defined. Use flag -s <AWS_SERVICE> to build a container image of that service"
  echo "(Example: $(basename "$0") -q -s ecr)"
  exit  1
fi

DEFAULT_AWS_SERVICE_DOCKER_IMG="ack-${AWS_SERVICE}-controller:${VERSION}"
: "${AWS_SERVICE_DOCKER_IMG:="$DEFAULT_AWS_SERVICE_DOCKER_IMG"}"
: "${DOCKERFILE:="$DOCKERFILE_PATH"}"

echo "Building '$AWS_SERVICE' controller docker image with tag: ${AWS_SERVICE_DOCKER_IMG}"

docker build \
  --quiet=${QUIET} \
  -t ${AWS_SERVICE_DOCKER_IMG} \
  -f ${DOCKERFILE} \
  --build-arg service_alias=${AWS_SERVICE} \
  ${BUILD_CONTEXT}

if [ $? -ne 0 ]; then
  exit 2
fi
