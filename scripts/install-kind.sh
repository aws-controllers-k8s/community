#!/usr/bin/env bash

# ./scripts/install-kind.sh [<kind version>]
#
# Installs KinD if not installed. Optional parameter specifies the version of
# KinD to install. Defaults to the value of the environment variable
# "KIND_VERSION" and if that is not set, the value of the DEFAULT_KIND_VERSION
# variable.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/kind

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
DEFAULT_KIND_VERSION="0.9.0"

source "$SCRIPTS_DIR/lib/common.sh"

__kind_version="$1"
if [ "x$__kind_version" == "x" ]; then
    __kind_version=${KIND_VERSION:-$DEFAULT_KIND_VERSION}
fi

if ! is_installed kind; then
    __kind_url="https://kind.sigs.k8s.io/dl/v${__kind_version}/kind-linux-amd64"
    echo -n "installing kind from $__kind_url ... "
    curl --silent -Lo ./kind "$__kind_url"
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
    echo "ok."
fi
