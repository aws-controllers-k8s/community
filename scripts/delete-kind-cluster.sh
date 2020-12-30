#!/bin/bash
set -eo pipefail

SCRIPTPATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

USAGE="
Usage:
  $(basename "$0") <Cluster_context_directory>

Deletes a kind cluster and context dir

Example: delete-cluster build/tmp-cluster-1234

<cluster_context_directory> Cluster context directory
"

if [ $# -ne 1 ]; then
    echo "Context directory is not defined" 1>&2
    echo "${USAGE}"
    exit 1
fi

TMP_DIR="$1"
CLUSTER_NAME=$(cat $TMP_DIR/clustername)

echo "🥑 Deleting k8s cluster using \"kind\""
kind delete cluster --name "$CLUSTER_NAME"
rm -r $TMP_DIR
