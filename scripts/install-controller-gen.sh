#!/usr/bin/env bash

# ./scripts/install-controller-gen.sh
#
# Checks that the `controller-gen` binary is available on the host system and
# if it is, that it matches the exact version that we require in order to
# standardize the YAML manifests for CRDs and Kubernetes Roles.
#
# If the locally-installed controller-gen does not match the required version,
# prints an error message asking the user to uninstall it.
#
# NOTE: We use this technique of building using `go build` within a temp
# directory because controller-tools does not have a binary release artifact
# for controller-gen.
#
# See: https://github.com/kubernetes-sigs/controller-tools/issues/500

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
CONTROLLER_TOOLS_VERSION="v0.4.0"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"

if ! is_installed controller-gen; then
    # GOBIN and GOPATH are not always set, so default to GOPATH from `go env`
    __GOPATH=$(go env GOPATH)
    __install_dir=${GOBIN:-$__GOPATH/bin}
    __install_path="$__install_dir/controller-gen"
    __work_dir=$(mktemp -d /tmp/controller-gen-XXX)

    echo -n "installing controller-gen ${CONTROLLER_TOOLS_VERSION} ... "
    cd "$__work_dir"

    go mod init tmp 1>/dev/null 2>&1
    go get -d "sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_TOOLS_VERSION}" 1>/dev/null 2>&1
    go build -o "$__work_dir/controller-gen" sigs.k8s.io/controller-tools/cmd/controller-gen 1>/dev/null 2>&1
    mv "$__work_dir/controller-gen" "$__install_path"

    rm -rf "$WORK_DIR"
    echo "ok."
fi
