#!/usr/bin/env bash

DEFAULT_KIND_VERSION="0.9.0"

# ensure_kind [<kind version>]
#
# Ensures that KinD is installed. Optional parameter specifies the version of
# KinD to install. Defaults to the value of the environment variable
# "KIND_VERSION" and if that is not set, the value of the DEFAULT_KIND_VERSION
# variable.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/kind
ensure_kind() {
    local __kind_version="$1"
    if [ "x$__kind_version" == "x" ]; then
        __kind_version=${KIND_VERSION:-$DEFAULT_KIND_VERSION}
    fi
    if ! is_installed kind; then
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v${__kind_version}/kind-linux-amd64
        chmod +x ./kind
        sudo mv kind /usr/local/bin/kind
    fi
}
