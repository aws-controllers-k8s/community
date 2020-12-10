#!/usr/bin/env bash

# ./scripts/install-kubectl.sh
#
# Installs the latest stable version kubectl if not installed.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/kubectl

set -eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."

source "$SCRIPTS_DIR/lib/common.sh"

if ! is_installed kubectl; then
    __platform=$(uname | tr '[:upper:]' '[:lower:]')
    __stable_k8s_version=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
    __kubectl_url="https://storage.googleapis.com/kubernetes-release/release/$__stable_k8s_version/bin/$__platform/amd64/kubectl"
    echo -n "installing kubectl from $__kubectl_url ... "
    curl --silent -Lo ./kubectl "$__kubectl_url"
    chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin/kubectl
    echo "ok."
fi
