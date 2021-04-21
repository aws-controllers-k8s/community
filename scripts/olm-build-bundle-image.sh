#!/usr/bin/env bash

set -eo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SCRIPTS_DIR=$DIR
ROOT_DIR=$DIR/..
BUILD_DATE=$(date +%Y-%m-%dT%H:%M)
QUIET=${QUIET:-"false"}
DEFAULT_DOCKER_REPOSITORY="amazon/aws-controllers-k8s"
DOCKER_REPOSITORY=${DOCKER_REPOSITORY:-$DEFAULT_DOCKER_REPOSITORY}

export DOCKER_BUILDKIT=${DOCKER_BUILDKIT:-1}

source $SCRIPTS_DIR/lib/common.sh

check_is_installed docker

# red hat operator certification requires additional labels and metadata
# in the bundle container image. 

# details on the required bundle labels
# https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/bundle-directory
DEFAULT_ADD_RH_CERTIFICATION_LABELS="false"
ADD_RH_CERTIFICATION_LABELS=${ADD_RH_CERTIFICATION_LABELS:-$DEFAULT_ADD_RH_CERTIFICATION_LABELS}

# details on how versions are reflected by this label.
# https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/bundle-directory/managing-openshift-versions
cert_label_openshift_supported_version="com.redhat.openshift.versions"
DEFAULT_SUPPORTED_OPENSHIFT_VERSIONS="v4.6" # v4.6 and on
SUPPORTED_OPENSHIFT_VERSIONS=${SUPPORTED_OPENSHIFT_VERSIONS:-$DEFAULT_SUPPORTED_OPENSHIFT_VERSIONS}

# Identifies that the delivery mechanism is the bundle format
cert_label_operator_bundle_delivery="com.redhat.delivery.operator.bundle"
DEFAULT_RED_HAT_DELIVERY_BUNDLE="true"
RED_HAT_DELIVERY_BUNDLE=${RED_HAT_DELIVERY_BUNDLE:-$DEFAULT_RED_HAT_DELIVERY_BUNDLE}

# Do not backport to OpenShift versions prior to v4.5 by default.
cert_label_deliver_backport="com.redhat.deliver.backport"
DEFAULT_RED_HAT_DELIVER_BACKPORT="false" # do not list in openshift <v4.5
RED_HAT_DELIVER_BACKPORT=${RED_HAT_DELIVER_BACKPORT:-$DEFAULT_RED_HAT_DELIVER_BACKPORT}

USAGE="
Usage:
  $(basename "$0") <aws_service> <bundle_version>

Builds the Docker image for an ACK service controller. 

Example: $(basename "$0") ecr

<aws_service> should be an AWS Service name (ecr, sns, sqs, petstore, bookstore)
<bundle_version> an operator lifecycle manager bundle version in semver (0.0.1, 1.0.0)


Environment variables:
  QUIET:                            Build bundle container image quietly (<true|false>)
                                    Default: false
  SERVICE_CONTROLLER_SOURCE_PATH:   Path to the service controller source code
                                    repository.
                                    Default: ../{SERVICE}-controller
  DOCKER_REPOSITORY:                Name for the Docker repository to push to 
                                    Default: $DEFAULT_DOCKER_REPOSITORY
  BUNDLE_DOCKERFILE_DIR:            Specify the directory where the bundle.Dockerfile exists.
                                    Default: {SERVICE_CONTROLLER_SOURCE_PATH}/olm
  BUNDLE_DOCKER_IMG_TAG:            Bundle container image tag
                                    Default: \$AWS_SERVICE-bundle-\$BUNDLE_VERSION
  BUNDLE_DOCKER_IMG:                The bundle container image (including the tag).
                                    Supercedes the use of BUNDLE_DOCKER_IMAGE_TAG
                                    and DOCKER_REPOSITORY if set.
                                    Default: $DEFAULT_DOCKER_REPOSITORY:\$AWS_SERVICE-bundle-\$BUNDLE_VERSION
  ADD_RH_CERTIFICATION_LABELS       Adds the certification labels required by Red Hat
                                    container certification (<true|false>)
                                    Default: $DEFAULT_ADD_RH_CERTIFICATION_LABELS
  SUPPORTED_OPENSHIFT_VERSIONS:     Indicates what versions of OpenShift are supported
                                    Only used if the ADD_RH_CERTIFICATION_LABELS is
                                    set to true.
                                    Default: $DEFAULT_SUPPORTED_OPENSHIFT_VERSIONS
  RED_HAT_DELIVERY_BUNDLE:          Red Hat should deliver the operator as a bundle.
                                    Only used if the ADD_RH_CERTIFICATION_LABELS is
                                    set to true.
                                    Default: $DEFAULT_RED_HAT_DELIVERY_BUNDLE
  RED_HAT_DELIVER_BACKPORT:         Red Hat should backport the operator to versions of
                                    OpenShift prior to v4.5. Only used if the
                                    ADD_RH_CERTIFICATION_LABELS is set to true.
                                    Default: $DEFAULT_RED_HAT_DELIVER_BACKPORT
"

if [ $# -ne 2 ]; then
    echo "AWS_SERVICE or BUNDLE_VERSION is not defined. Script accepts two parameters, the <AWS_SERVICE> and the <BUNDLE_VERSION> to build." 1>&2
    echo "${USAGE}"
    exit 1
fi

AWS_SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')
BUNDLE_VERSION="$2"

DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH="$ROOT_DIR/../$AWS_SERVICE-controller"
SERVICE_CONTROLLER_SOURCE_PATH=${SERVICE_CONTROLLER_SOURCE_PATH:-$DEFAULT_SERVICE_CONTROLLER_SOURCE_PATH}
DEFAULT_BUNDLE_DOCKERFILE_DIR="${SERVICE_CONTROLLER_SOURCE_PATH}/olm"
BUNDLE_DOCKERFILE_DIR=${BUNDLE_DOCKERFILE_DIR:-$DEFAULT_BUNDLE_DOCKERFILE_DIR}
BUNDLE_DOCKERFILE="$BUNDLE_DOCKERFILE_DIR/bundle.Dockerfile"

# stop if the dockerfile was not found
if [ ! -f $BUNDLE_DOCKERFILE ]; then
  echo "The bundle.Dockerfile was not found at expected path $BUNDLE_DOCKERFILE."
  exit 1
fi

DEFAULT_BUNDLE_DOCKER_IMG_TAG="$AWS_SERVICE-bundle-$BUNDLE_VERSION"
BUNDLE_DOCKER_IMG_TAG=${BUNDLE_DOCKER_IMG_TAG:-$DEFAULT_BUNDLE_DOCKER_IMG_TAG}
BUNDLE_DOCKER_IMG=${BUNDLE_DOCKER_IMAGE:-$DOCKER_REPOSITORY:$BUNDLE_DOCKER_IMG_TAG}

build_args="--quiet=${QUIET} -t ${BUNDLE_DOCKER_IMG} -f ${BUNDLE_DOCKERFILE} --build-arg build_date=\"$BUILD_DATE\""

if [[ $ADD_RH_CERTIFICATION_LABELS = "true" ]]; then 
  # add additional labels with values for certification purposes.
  build_args="$build_args --label=$cert_label_openshift_supported_version=$SUPPORTED_OPENSHIFT_VERSIONS --label=$cert_label_deliver_backport=$RED_HAT_DELIVER_BACKPORT --label=$cert_label_operator_bundle_delivery=$RED_HAT_DELIVERY_BUNDLE"
fi

if [[ $QUIET = "false" ]]; then
    echo "building '$AWS_SERVICE' OLM bundle container image with tag: ${BUNDLE_DOCKER_IMG}"
fi

pushd $BUNDLE_DOCKERFILE_DIR 1>/dev/null
docker build $build_args .

if [ $? -ne 0 ]; then
  exit 2
fi

popd 1>/dev/null