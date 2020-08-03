#!/usr/bin/env bash

# A script that provisions a KinD Kubernetes cluster for local development and
# testing

set -Eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."

source "$SCRIPTS_DIR"/lib/common.sh
source "$SCRIPTS_DIR"/lib/kind.sh
source "$SCRIPTS_DIR"/lib/k8s.sh
source "$SCRIPTS_DIR"/lib/helm.sh

SCRIPT_PATH="$(cd "$(dirname "$0")" ; pwd -P )"
PLATFORM=$(uname | tr '[:upper:]' '[:lower:]')
CLUSTER_CREATION_TIMEOUT_IN_SEC=300
TEST_ID=$(uuidgen | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
CLUSTER_NAME_BASE=$(uuidgen | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
OVERRIDE_PATH=0
KIND_CONFIG_FILE=$SCRIPT_PATH/kind-two-node-cluster.yaml

K8_1_18="kindest/node:v1.18.4@sha256:9ddbe5ba7dad96e83aec914feae9105ac1cffeb6ebd0d5aa42e820defe840fd4"
K8_1_17="kindest/node:v1.17.5@sha256:ab3f9e6ec5ad8840eeb1f76c89bb7948c77bbf76bcebe1a8b59790b8ae9a283a"
K8_1_16="kindest/node:v1.16.9@sha256:7175872357bc85847ec4b1aba46ed1d12fa054c83ac7a8a11f5c268957fd5765"
K8_1_15="kindest/node:v1.15.11@sha256:6cc31f3533deb138792db2c7d1ffc36f7456a06f1db5556ad3b6927641016f50"
K8_1_14="kindest/node:v1.14.10@sha256:6cd43ff41ae9f02bb46c8f455d5323819aec858b99534a290517ebc181b443c6"

K8_VERSION="$K8_1_16"

echoerr() { echo "$@" 1>&2; }

USAGE="
Usage:
  $(basename "$0") [-b <BASE_CLUSTER_NAME>] [-i <TEST_IDENTIFIER>] [-v K8s_VERSION]

Provisions a KinD cluster for local development and testing. Outputs the
directory containing the KinD/kubectl cluster context to stdout on successful
completion

Example: $(basename "$0") -b my-test -i 123 -v 1.16

      Optional:
        -b          Base Name of cluster
        -i          Test Identifier to suffix Cluster Name and tmp dir
        -v          K8s version to use in this test
        -k          Kind cluster config file
"

# Process our input arguments
while getopts "b:i:v:k:o" opt; do
  case ${opt} in
    b ) # BASE CLUSTER NAME
        CLUSTER_NAME_BASE=$OPTARG
      ;;
    i ) # Test ID
        TEST_ID=$OPTARG
        echoerr "👉 Test Run: $TEST_ID 👈"
      ;;
    v ) # K8s version to provision
        OPTARG="K8_$(echo "${OPTARG}" | sed 's/\./\_/g')"
        if [ ! -z ${OPTARG+x} ]; then
            K8_VERSION=${!OPTARG}
        else
            echoerr "K8s version not supported"
            exit 2
        fi
      ;;
    k ) # Kind cluster config file
        KIND_CONFIG_FILE="${OPTARG}"
      ;;
    \? )
        echoerr "${USAGE}" 1>&2
        exit
      ;;
  esac
done

check_is_installed docker

ensure_kind
ensure_kubectl
ensure_helm

CLUSTER_NAME="$CLUSTER_NAME_BASE"-"${TEST_ID}"
TMP_DIR=$ROOT_DIR/build/tmp-$CLUSTER_NAME

echoerr "🐳 Using Kubernetes $K8_VERSION"
mkdir -p "${TMP_DIR}"

echoerr "🥑 Creating k8s cluster using \"kind\""
for i in $(seq 0 5); do
  if [[ -z $(kind get clusters | grep $CLUSTER_NAME) ]]; then
      kind create cluster --name "$CLUSTER_NAME" --image $K8_VERSION --config "$SCRIPT_PATH/kind-two-node-cluster.yaml" --kubeconfig $TMP_DIR/kubeconfig 1>&2 || :
  else
      break
  fi
done

echo "$CLUSTER_NAME" > $TMP_DIR/clustername
echoerr "👍 Created k8s cluster using \"kind\""
echo $TMP_DIR
