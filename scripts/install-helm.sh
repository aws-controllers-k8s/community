#!/usr/bin/env bash

# ./scripts/install-helm.sh
#
# Installs Helm if not installed. Optional parameter specifies the version of
# Helm to install. Defaults to the value of the environment variable
# "HELM_VERSION" and if that is not set, the value of the DEFAULT_HELM_VERSION
# variable.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/helm

set -Eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
DEFAULT_HELM_VERSION="3.2.4"

source "$SCRIPTS_DIR/lib/common.sh"

__helm_version="$1"
if [ "x$__helm_version" == "x" ]; then
    __helm_version=${HELM_VERSION:-$DEFAULT_HELM_VERSION}
fi
if ! is_installed helm; then
    __platform=$(uname | tr '[:upper:]' '[:lower:]')
    __tmp_install_dir=$(mktemp -d -t install-helm-XXX)
    __helm_url="https://get.helm.sh/helm-v$__helm_version-$__platform-amd64.tar.gz"
    echo -n "installing helm from $__helm_url ... "
    curl -q -L $__helm_url | tar zxf - -C $__tmp_install_dir
    mv $__tmp_install_dir/$__platform-amd64/helm $__tmp_install_dir/.
    chmod +x $__tmp_install_dir/helm
    sudo mv $__tmp_install_dir/helm /usr/local/bin/helm
    echo "ok."
fi
