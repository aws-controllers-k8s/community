#!/usr/bin/env bash

# ./scripts/install-operator-sdk.sh
#
#
# Installs Operator SDK if not installed. Optional parameters specifies the
# version of Operator SDK to install. Defaults tot eh value of the environment
# variable OPERATOR_SDK_VERSION and if that is not set, the value of the
# DEFAULT_OPERATOR_SDK_VERSION variable.
#
# NOTE: uses `sudo mv` to relocate a downloaded binary to /usr/local/bin/operator-sdk

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
DEFAULT_OPERATOR_SDK_VERSION="1.7.1"

source "${SCRIPTS_DIR}/lib/common.sh"

__operator_sdk_version="${1}"
if [ "x${__operator_sdk_version}" == "x" ]; then
    __operator_sdk_version=${OPERATOR_SDK_VERSION:-$DEFAULT_OPERATOR_SDK_VERSION}
fi
if ! is_installed operator-sdk; then
    __platform=$(uname | tr '[:upper:]' '[:lower:]')
    __tmp_install_dir=$(mktemp -d -t install-operator-sdk-XXX)
    __operator_sdk_url="https://github.com/operator-framework/operator-sdk/releases/download/v${__operator_sdk_version}/operator-sdk_${__platform}_amd64"
    echo -n "installing operator-sdk from ${__operator_sdk_url} ... "
    curl -sq -L ${__operator_sdk_url} --output ${__tmp_install_dir}/operator-sdk_${__platform}_amd64
    chmod +x ${__tmp_install_dir}/operator-sdk_${__platform}_amd64
    sudo mv ${__tmp_install_dir}/operator-sdk_${__platform}_amd64 /usr/local/bin/operator-sdk
    echo "ok."
fi