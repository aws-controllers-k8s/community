#!/usr/bin/env bash

set -eo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SCRIPTS_DIR=$DIR
DEFAULT_DOCKER_REPOSITORY="amazon/aws-controllers-k8s"
DOCKER_REPOSITORY=${DOCKER_REPOSITORY:-$DEFAULT_DOCKER_REPOSITORY}

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

USAGE="
Usage:
  $(basename "$0") <AWS_SERVICE> <BUNDLE_VERSION>

Publishes the Docker image for an ACK service OLM bundle. By default, the
repository will be $DEFAULT_DOCKER_REPOSITORY and the image tag for the
specific ACK service controller will be ":\$SERVICE-bundle-\$VERSION".

<AWS_SERVICE> AWS Service name (ecr, sns, sqs)
<BUNDLE_VERSION> OLM bundle version in SemVer (0.0.1, 1.0.0)

Example: 
export DOCKER_REPOSITORY=aws-controllers-k8s
$(basename "$0") ecr 0.0.1

Environment variables:
  DOCKER_REPOSITORY:        Name for the Docker repository to push to 
                            Default: $DEFAULT_DOCKER_REPOSITORY
  BUNDLE_DOCKER_IMG_TAG:    Bundle container image tag
                            Default: \$AWS_SERVICE-bundle-\$BUNDLE_VERSION
  BUNDLE_DOCKER_IMG:        The bundle container image (including the tag).
                            Supercedes the use of BUNDLE_DOCKER_IMAGE_TAG
                            and DOCKER_REPOSITORY if set.
                            Default: $DEFAULT_DOCKER_REPOSITORY:\$AWS_SERVICE-bundle-\$BUNDLE_VERSION
"

if [ $# -ne 2 ]; then
    echo "AWS_SERVICE or BUNDLE_VERSION is not defined. Script accepts two parameters, the <AWS_SERVICE> and the <BUNDLE_VERSION> to build." 1>&2
    echo "${USAGE}"
    exit 1
fi

AWS_SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')
BUNDLE_VERSION="$2"

DEFAULT_BUNDLE_DOCKER_IMG_TAG="$AWS_SERVICE-bundle-$BUNDLE_VERSION"
BUNDLE_DOCKER_IMG_TAG=${BUNDLE_DOCKER_IMG_TAG:-$DEFAULT_BUNDLE_DOCKER_IMG_TAG}
BUNDLE_DOCKER_IMG=${BUNDLE_DOCKER_IMAGE:-$DOCKER_REPOSITORY:$BUNDLE_DOCKER_IMG_TAG}

export BUNDLE_DOCKER_IMG
${SCRIPTS_DIR}/olm-build-bundle-image.sh ${AWS_SERVICE} ${BUNDLE_VERSION}

echo "Pushing '$AWS_SERVICE' operator lifecycle manager bundle image with tag: ${BUNDLE_DOCKER_IMG_TAG}"

docker push ${BUNDLE_DOCKER_IMG}

if [ $? -ne 0 ]; then
  exit 2
fi
