#!/usr/bin/env bash

set -eo pipefail

DIR=$(cd "$(dirname "$0")"; pwd)
SCRIPTS_DIR=$DIR
DEFAULT_DOCKER_REPOSITORY="amazon/aws-controllers-k8s"
DOCKER_REPOSITORY=${DOCKER_REPOSITORY:-$DEFAULT_DOCKER_REPOSITORY}
VERSION=$(git describe --tags --always --dirty || echo "unknown")

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

USAGE="
Usage:
  $(basename "$0") <AWS_SERVICE> 

Publishes the Docker image for an ACK service controller. By default, the
repository will be $DEFAULT_DOCKER_REPOSITORY and the image tag for the
specific ACK service controller will be ":\$SERVICE-\$VERSION".

<AWS_SERVICE> AWS Service name (ecr, sns, sqs)

Example: 
export DOCKER_REPOSITORY=aws-controllers-k8s
$(basename "$0") ecr 

Environment variables:
  DOCKER_REPOSITORY:        Name for the Docker repository to push to 
                            Default: $DEFAULT_DOCKER_REPOSITORY
  AWS_SERVICE_DOCKER_IMG:   Controller container image tag 
                            Default: aws-controllers-k8s:$AWS_SERVICE-$VERSION
"

if [ $# -ne 1 ]; then
    echo "AWS_SERVICE is not defined. Script accepts one parameter, <AWS_SERVICE> to build that docker images of that service" 1>&2
    echo "${USAGE}"
    exit 1
fi

AWS_SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')

DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG="${AWS_SERVICE}-${VERSION}"
AWS_SERVICE_DOCKER_IMG_TAG=${AWS_SERVICE_DOCKER_IMG_TAG:-"$DOCKER_REPOSITORY:$DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG"}
AWS_SERVICE_DOCKER_IMG=${AWS_SERVICE_DOCKER_IMG:-"$DOCKER_REPOSITORY:$DEFAULT_AWS_SERVICE_DOCKER_IMG_TAG"}

export AWS_SERVICE_DOCKER_IMG
${SCRIPTS_DIR}/build-controller-image.sh ${AWS_SERVICE}

echo "Pushing '$AWS_SERVICE' controller docker image with tag: ${AWS_SERVICE_DOCKER_IMG_TAG}"

docker push ${AWS_SERVICE_DOCKER_IMG_TAG}

if [ $? -ne 0 ]; then
  exit 2
fi
