#!/usr/bin/env bash

# A script that installs the mockery CLI tool that is used to build Go mocks
# for our interfaces to use in unit testing. This script installs mockery into
# the bin/mockery path and really should just be used in testing scripts.

set -euxo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
BIN_DIR="$ROOT_DIR/bin"

OS=$(uname -s)
ARCH=$(uname -m)
VERSION=2.2.2
MOCKERY_RELEASE_URL="https://github.com/vektra/mockery/releases/download/v${VERSION}/mockery_${VERSION}_${OS}_${ARCH}.tar.gz"

if [[ ! -f $BIN_DIR/mockery ]]; then
    echo -n "Installing mockery into bin/mockery ... "
    mkdir -p $BIN_DIR
    cd $BIN_DIR
    wget -q --no-check-certificate --content-disposition $MOCKERY_RELEASE_URL -O mockery.tar.gz
    tar -xf mockery.tar.gz
    echo "ok."
fi
