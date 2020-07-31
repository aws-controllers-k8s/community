#!/bin/bash
set -euo pipefail

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
KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
KIND_VERSION="0.8.1"
HELM_VERSION="3.2.4"

echoerr() { echo "$@" 1>&2; }

USAGE=$(cat << 'EOM'
  Usage: provision-cluster  [-b <BASE_CLUSTER_NAME>] [-i <TEST_IDENTIFIER>] [-v K8s_VERSION] [-o]
  Executes the spot termination integration test for the Node Termination Handler.
  Outputs the cluster context directory to stdout on successful completion

  Example: provision-cluster -b my-test -i 123 -v 1.16

          Optional:
            -b          Base Name of cluster
            -i          Test Identifier to suffix Cluster Name and tmp dir
            -v          K8s version to use in this test
            -k          Kind cluster config file
            -o          Override path w/ your own kubectl and kind binaries
EOM
)

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
    o ) # Override path with your own kubectl and kind binaries
	    OVERRIDE_PATH=1
      ;;
    \? )
        echoerr "${USAGE}" 1>&2
        exit
      ;;
  esac
done

CLUSTER_NAME="$CLUSTER_NAME_BASE"-"${TEST_ID}"
TMP_DIR=$SCRIPT_PATH/build/tmp-$CLUSTER_NAME

echoerr "🐳 Using Kubernetes $K8_VERSION"
mkdir -p "${TMP_DIR}"

deps=("docker")

for dep in "${deps[@]}"; do
    path_to_executable=$(which $dep)
    if [ ! -x "$path_to_executable" ]; then
        echoerr "You are required to have $dep installed on your system..."
        echoerr "Please install $dep and try again. "
        exit 3
    fi
done

## Append to the end of PATH so that the user can override the executables if they want
if [[ OVERRIDE_PATH -eq 1 ]]; then
   export PATH=$PATH:$TMP_DIR
else
  if [ ! -x "$TMP_DIR/kubectl" ]; then
      echoerr "🥑 Downloading the \"kubectl\" binary"
      curl -Lo $TMP_DIR/kubectl "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/$PLATFORM/amd64/kubectl"
      chmod +x $TMP_DIR/kubectl
      echoerr "👍 Downloaded the \"kubectl\" binary"
  fi

  if [ ! -x "$TMP_DIR/kind" ]; then
      echoerr "🥑 Downloading the \"kind\" binary"
      curl -Lo $TMP_DIR/kind https://github.com/kubernetes-sigs/kind/releases/download/v$KIND_VERSION/kind-$PLATFORM-amd64
      chmod +x $TMP_DIR/kind
      echoerr "👍 Downloaded the \"kind\" binary"
  fi

  if [ ! -x "$TMP_DIR/helm" ]; then
      echoerr "🥑 Downloading the \"helm\" binary"
      curl -L https://get.helm.sh/helm-v$HELM_VERSION-$PLATFORM-amd64.tar.gz | tar zxf - -C $TMP_DIR
      mv $TMP_DIR/$PLATFORM-amd64/helm $TMP_DIR/.
      chmod +x $TMP_DIR/helm
      echoerr "👍 Downloaded the \"helm\" binary"
  fi
  export PATH=$TMP_DIR:$PATH
fi

echoerr "🥑 Creating k8s cluster using \"kind\""
for i in $(seq 0 5); do
  if [[ -z $(kind get clusters | grep $CLUSTER_NAME) ]]; then
      kind create cluster -q --name "$CLUSTER_NAME" --image $K8_VERSION --config "$SCRIPT_PATH/kind-two-node-cluster.yaml" --kubeconfig $TMP_DIR/kubeconfig 1>&2 || :
  else
      break
  fi
done

echo "$CLUSTER_NAME" > $TMP_DIR/clustername
echoerr "👍 Created k8s cluster using \"kind\""

kubectl apply -f "$SCRIPT_PATH/psp-default.yaml" --context kind-$CLUSTER_NAME --kubeconfig $TMP_DIR/kubeconfig 1>&2
kubectl apply -f "$SCRIPT_PATH/psp-privileged.yaml" --context kind-$CLUSTER_NAME --kubeconfig $TMP_DIR/kubeconfig 1>&2

echo $TMP_DIR
