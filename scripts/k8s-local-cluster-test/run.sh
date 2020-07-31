#!/bin/bash
set -eo pipefail

AEMM_URL="amazon-ec2-metadata-mock-service.default.svc.cluster.local"
AEMM_VERSION="1.2.0"
AEMM_DL_URL="https://github.com/aws/amazon-ec2-metadata-mock/releases/download/v${AEMM_VERSION}/amazon-ec2-metadata-mock-${AEMM_VERSION}.tgz"
CLUSTER_NAME_BASE="nth-test"
DELETE_CLUSTER_ARGS=""
K8S_VERSION="1.16"
OVERRIDE_PATH=0
PRESERVE=false
PROVISION_CLUSTER_ARGS=""
START=$(date +%s)
SCRIPT_PATH="$( cd "$(dirname "$0")" ; pwd -P )"
TMP_DIR=""
# VERSION is the source revision that executables and images are built from.
VERSION=$(git describe --tags --always --dirty || echo "unknown")

function timeout() { perl -e 'alarm shift; exec @ARGV' "$@"; }

function relpath() {
  perl -e 'use File::Spec; print File::Spec->abs2rel(@ARGV) . "\n"' "${1}" "${2}"
}

function clean_up {
    if [[ "$PRESERVE" == false ]]; then
        "${SCRIPT_PATH}"/delete-cluster.sh $DELETE_CLUSTER_ARGS || :
        return
    fi
    echo "To resume test with the same cluster use: \"-c $TMP_DIR\""""
}

function exit_and_fail {
    local pod_id=$(get_nth_worker_pod || :)
    kubectl logs "${pod_id}" --namespace kube-system || :
    END=$(date +%s)
    echo "⏰ Took $(expr "${END}" - "${START}")sec"
    echo "❌ NTH Integration Test FAILED $CLUSTER_NAME! ❌"
    exit 1
}

function get_nth_worker_pod {
    kubectl get pods -n kube-system \
      --selector 'app.kubernetes.io/name=aws-node-termination-handler' \
      --field-selector="spec.nodeName=$CLUSTER_NAME-worker,status.phase=Running" \
      --sort-by=.metadata.creationTimestamp \
      --output jsonpath='{.items[-1].metadata.name}'
}

USAGE=$(cat << 'EOM'
  Usage: bash run.sh [-p] [-s] [-o] [-b <TEST_BASE_NAME>] [-c <CLUSTER_CONTEXT_DIR>] [-i <AWS Docker image name>] [-s] [-v K8S_VERSION]
  Executes a test within a provisioned kubernetes cluster with NTH and IMDS pre-loaded.

  Example: bash run.sh -p -s petstore

          Optional:
            -b          Base name of test (will be used for cluster too)
            -c          Cluster context directory, if operating on an existing cluster
            -p          Preserve kind k8s cluster for inspection
            -i          Provide AWS Service docker image
            -s          Provide AWS Service name (ecr, sns, sqs, petstore, bookstore)
            -o          Override path w/ your own kubectl and kind binaries
            -v          Kubernetes Version (Default: 1.16) [1.14, 1.15, 1.16, 1.17, and 1.18]

EOM
)

# Process our input arguments
while getopts "ps:ioc:b:v:" opt; do
  case ${opt} in
    p ) # PRESERVE K8s Cluster
        echo "❄️  This run will preserve the cluster as you requested"
        PRESERVE=true
      ;;
    s ) # AWS Service name
        echo "Running Docker build as you requested for ${OPTARG} service"
        AWS_SERVICE=$(echo "${OPTARG}" | tr '[:upper:]' '[:lower:]')
        echo $AWS_SERVICE
      ;;
    i ) # AWS Service Docker Image
        AWS_SERVICE_DOCKER_IMG="${OPTARG}"
      ;;
    o ) # Override path with your own kubectl and kind binaries
        DELETE_CLUSTER_ARGS="${DELETE_CLUSTER_ARGS} -o"
        PROVISION_CLUSTER_ARGS="${PROVISION_CLUSTER_ARGS} -o"
        OVERRIDE_PATH=1
      ;;
    c ) # Cluster context directory to operate on existing cluster
        TMP_DIR="${OPTARG}"
      ;;
    b ) # Base cluster name
        CLUSTER_NAME_BASE="${OPTARG}"
      ;;
    v ) # K8s VERSION
        K8S_VERSION="${OPTARG}"
      ;;
    \? )
        echo "${USAGE}" 1>&2
        exit
      ;;
  esac
done

if [ -z $TMP_DIR ]; then
    TMP_DIR=$("${SCRIPT_PATH}"/../k8s-local-cluster-test/provision-cluster.sh -b "${CLUSTER_NAME_BASE}" -v "${K8S_VERSION}" "${PROVISION_CLUSTER_ARGS}")
fi

if [ $OVERRIDE_PATH == 0 ]; then
  export PATH=$TMP_DIR:$PATH
else
  export PATH=$PATH:$TMP_DIR
fi

CLUSTER_NAME=$(cat $TMP_DIR/clustername)

## Build and Load Docker Images

if [ -z "$AWS_SERVICE_DOCKER_IMG" ]; then
    echo "🥑 Building ${AWS_SERVICE} docker image"
    DEFAULT_AWS_SERVICE_DOCKER_IMG="${AWS_SERVICE}:${VERSION}"
    docker build -f services/"${AWS_SERVICE}"/Dockerfile -t "${DEFAULT_AWS_SERVICE_DOCKER_IMG}" .
    AWS_SERVICE_DOCKER_IMG="${DEFAULT_AWS_SERVICE_DOCKER_IMG}"
    echo "👍 Built the ${AWS_SERVICE} docker image"
else
    echo "🥑 Skipping building the ${AWS_SERVICE} docker image, since one was specified ${AWS_SERVICE_DOCKER_IMG}"
fi
echo "$AWS_SERVICE_DOCKER_IMG" > "${TMP_DIR}"/nth-docker-img

echo "🥑 Loading the images into the cluster"
kind load docker-image --name "${CLUSTER_NAME}" --nodes="${CLUSTER_NAME}"-worker,"${CLUSTER_NAME}"-control-plane "${AWS_SERVICE_DOCKER_IMG}"
echo "👍 Loaded image(s) into the cluster"

export KUBECONFIG="${TMP_DIR}/kubeconfig"

trap "exit_and_fail" INT TERM ERR
trap "clean_up" EXIT

echo "======================================================================================================"
echo "To poke around your test manually:"
echo "export KUBECONFIG=$TMP_DIR/kubeconfig"
echo "export PATH=$TMP_DIR:\$PATH"
echo "kubectl get pods -A"
echo "======================================================================================================"


### exported vars and funcs that tests can use
export TMP_DIR
export CLUSTER_NAME
export AEMM_URL
export AEMM_VERSION
export AEMM_DL_URL
export -f timeout
export -f relpath
export -f get_nth_worker_pod
export NTH_WORKER_LABEL="kubernetes\.io/hostname=${CLUSTER_NAME}-worker"
###

echo "======================================================================================================"
echo "✅ All tests passed! ✅"
echo "======================================================================================================"
