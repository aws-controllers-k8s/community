#!/usr/bin/env bash

# ./scripts/install-kustomize.sh
#
# Installs the latest version kustomize if not installed.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/kustomize

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."

source "$SCRIPTS_DIR/lib/common.sh"

if ! is_installed kustomize ; then
    __kustomize_url="https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
    echo -n "installing kustomize from $__kustomize_url ... "
    curl --silent "$__kustomize_url" | bash 1>/dev/null
    chmod +x kustomize
    sudo mv kustomize /usr/local/bin/kustomize
    echo "ok."
fi
