#!/usr/bin/env bash

# A script that ensures the KinD tool is installed

set -Eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."

source "$SCRIPTS_DIR/lib/kind.sh"

ensure_kind
