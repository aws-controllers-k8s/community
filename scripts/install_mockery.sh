#!/usr/bin/env bash

# A script that installs the mockery CLI tool that is used to build Go mocks
# for our interfaces to use in unit testing. This script installs mockery into
# the bin/mockery path and really should just be used in testing scripts.

set -Eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
BIN_DIR="$ROOT_DIR/bin"
MOCKERY_RELEASE_URL="https://github.com/vektra/mockery/releases/download/v2.1.0/mockery_2.1.0_Linux_x86_64.tar.gz"

source "$SCRIPTS_DIR/lib/common.sh"

if ! is_installed mockery; then
    mkdir -p $BIN_DIR
    cd $BIN_DIR
    wget -q $MOCKERY_RELEASE_URL -O mockery.tar.gz
    tar -xvf mockery.tar.gz
    export PATH="$PATH:$BIN_DIR"
fi
