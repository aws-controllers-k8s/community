#!/usr/bin/env bash

# A script that provisions a KinD Kubernetes cluster for local development and
# testing

set -eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."

source "$SCRIPTS_DIR"/lib/common.sh

check_is_installed uuidgen
check_is_installed wget
check_is_installed docker
check_is_installed kind "You can install kind with the helper scripts/install-kind.sh"

KIND_CONFIG_FILE="$SCRIPTS_DIR/kind-two-node-cluster.yaml"

K8_1_18="kindest/node:v1.18.4@sha256:9ddbe5ba7dad96e83aec914feae9105ac1cffeb6ebd0d5aa42e820defe840fd4"
K8_1_17="kindest/node:v1.17.5@sha256:ab3f9e6ec5ad8840eeb1f76c89bb7948c77bbf76bcebe1a8b59790b8ae9a283a"
K8_1_16="kindest/node:v1.16.9@sha256:7175872357bc85847ec4b1aba46ed1d12fa054c83ac7a8a11f5c268957fd5765"
K8_1_15="kindest/node:v1.15.11@sha256:6cc31f3533deb138792db2c7d1ffc36f7456a06f1db5556ad3b6927641016f50"
K8_1_14="kindest/node:v1.14.10@sha256:6cd43ff41ae9f02bb46c8f455d5323819aec858b99534a290517ebc181b443c6"

USAGE="
Usage:
  $(basename "$0") CLUSTER_NAME

Provisions a KinD cluster for local development and testing.

Example: $(basename "$0") my-test

Environment variables:
  K8S_VERSION               Kubernetes Version [1.14, 1.15, 1.16, 1.17, and 1.18]           
                            Default: 1.16
"

cluster_name="$1"
if [[ -z "$cluster_name" ]]; then
    echo "FATAL: required cluster name argument missing."
    echo "${USAGE}" 1>&2
    exit 1
fi

# Process K8_VERSION env var

if [ ! -z ${K8_VERSION} ]; then
    K8_VER="K8_$(echo "${K8_VERSION}" | sed 's/\./\_/g')"
    K8_VERSION=${!K8_VER}
    
    # Check if version is supported:
    if [ -z $K8_VERSION ]; then
        echo "Version set: $K8_VER"
        echo "K8s version not supported" 1>&2
        exit 2
    fi
else
    K8_VERSION=${K8_1_16}
fi

TMP_DIR=$ROOT_DIR/build/tmp-$cluster_name
mkdir -p "${TMP_DIR}"

debug_msg "kind: using Kubernetes $K8_VERSION"
echo -n "creating kind cluster $cluster_name ... "
for i in $(seq 0 5); do
  if [[ -z $(kind get clusters 2>/dev/null | grep $cluster_name) ]]; then
      kind create cluster -q --name "$cluster_name" --image $K8_VERSION --config "$KIND_CONFIG_FILE" --kubeconfig $TMP_DIR/kubeconfig 1>&2 || :
  else
      break
  fi
done
echo "ok."

echo "$cluster_name" > $TMP_DIR/clustername
