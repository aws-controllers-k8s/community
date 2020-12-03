#!/usr/bin/env bash

set -euxo pipefail

DIR=$(cd "$(dirname "$0")"; pwd)
SCRIPTS_DIR=$DIR
DEFAULT_DOCKER_REPOSITORY="amazon/aws-controllers-k8s"
DOCKER_REPOSITORY=${DOCKER_REPOSITORY:-$DEFAULT_DOCKER_REPOSITORY}
OPTIND=1
VERSION=$(git describe --tags --always --dirty || echo "unknown")

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

USAGE="
Usage:
  $(basename "$0") -s <AWS_SERVICE> -r <DOCKER_REPOSITORY> [-i <Docker image tag>]

Publishes the Docker image for an ACK service controller. By default, the
repository will be $DEFAULT_DOCKER_REPOSITORY and the image tag for the
specific ACK service controller will be ":\$SERVICE-\$VERSION".

Example: $(basename "$0") -s ecr -r amazon/aws-controllers-k8s

Options:
  -s          Provide AWS Service name (ecr, sns, sqs)
  -i          Controller container image tag (default: ack-<service>-controller:$VERSION)
  -r          Name for the Docker repository to push to (default: $DEFAULT_DOCKER_REPOSITORY)
"

# Process our input arguments
while getopts "r:s:i:" opt; do
  case ${opt} in
    r ) # Docker repository name
        DOCKER_REPOSITORY="${OPTARG}"
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

DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG="${AWS_SERVICE}-${VERSION}"
: "${AWS_SERVICE_DOCKER_IMG_TAG:="$DOCKER_REPOSITORY:$DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG"}"

: "${AWS_SERVICE_DOCKER_IMG:="$DOCKER_REPOSITORY:$DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG"}"
export AWS_SERVICE_DOCKER_IMG
${SCRIPTS_DIR}/build-controller-image.sh ${AWS_SERVICE}

echo "Pushing '$AWS_SERVICE' controller docker image with tag: ${AWS_SERVICE_DOCKER_IMG_TAG}"

docker push ${AWS_SERVICE_DOCKER_IMG_TAG}

if [ $? -ne 0 ]; then
  exit 2
fi
