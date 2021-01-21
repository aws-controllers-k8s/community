#!/usr/bin/env bash

# This test script generates the Helm Chart for a supplied ACK service
# controller and specified release version, installs the ACK service controller
# for S3 using the generated Helm chart and then uninstalls the controller
# using Helm.
#
# You should have already created a Kubernetes cluster (perhaps using
# ./scripts/provision-kind-cluster.sh) and exported the KUBECONFIG
# appropriately before running this script.

set -eo pipefail

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"
BUILD_DIR="$ROOT_DIR/build"

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed helm "You can install helm with the helper scripts/install-helm.sh"

USAGE="
Usage:
  $(basename "$0") <service> <release_version>

<service> should be an AWS service for which you wish to run tests -- e.g.
's3' 'sns' or 'sqs'

<release_version> should be the SemVer version string to build a Helm chart
for. This release version string should match the name of the Docker image that
has previously been built.
"

SERVICE="$1"
RELEASE_VERSION="$2"
K8S_NAMESPACE="ack-system-test-helm"

CHART_DIR="$BUILD_DIR/release/$RELEASE_VERSION/$SERVICE"

echo "testing Helm release for $SERVICE for release version $RELEASE_VERSION."
mkdir -p $CHART_DIR

# We need to do this to prevent "unable to open go.mod" errors from
# ack-generate...
pushd $ROOT_DIR 1>/dev/null

ACK_GENERATE_IMAGE_REPOSITORY="aws-controllers-k8s" \
    ACK_GENERATE_SERVICE_ACCOUNT_NAME="ack-$SERVICE-controller-helm-test" \
    ACK_GENERATE_OUTPUT_PATH="$CHART_DIR" \
    K8S_RBAC_ROLE_NAME="ack-$SERVICE-controller-helm-test" \
    $SCRIPTS_DIR/build-controller-release.sh "$SERVICE" "$RELEASE_VERSION"

popd 1>/dev/null

kubectl create namespace "$K8S_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

pushd $CHART_DIR/helm 1>/dev/null

echo -n "installing the helm chart for ack-$SERVICE-controller in namespace $K8S_NAMESPACE ... "
helm install --namespace "$K8S_NAMESPACE" ack-$SERVICE-controller-helm-test . 1>/dev/null || exit 1
echo "ok."

echo -n "uninstalling the helm chart for ack-$SERVICE-controller in namespace $K8S_NAMESPACE ... "
helm uninstall --namespace "$K8S_NAMESPACE" ack-$SERVICE-controller-helm-test 1>/dev/null || exit 1
echo "ok."
