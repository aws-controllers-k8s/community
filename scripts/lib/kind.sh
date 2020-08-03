#!/usr/bin/env bash

DEFAULT_KIND_VERSION="0.8.1"

# ensure_kind [<kind version>]
#
# Ensures that KinD is installed. Optional parameter specifies the version of
# KinD to install. Defaults to the value of the environment variable
# "KIND_VERSION" and if that is not set, the value of the DEFAULT_KIND_VERSION
# variable.
ensure_kind() {
    local __kind_version="$1"
    if [ "x$__kind_version" == "x" ]; then
        __kind_version=${KIND_VERSION:-$DEFAULT_KIND_VERSION}
    fi
    if ! is_installed kind; then
        go get "sigs.k8s.io/kind@v$__kind_version"
    fi
}
