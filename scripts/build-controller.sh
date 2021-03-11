#!/usr/bin/env bash

# A script that builds a single ACK service controller for an AWS service API

COMMUNITY_SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$COMMUNITY_SCRIPTS_DIR/.."
TARGET_SCRIPT_PATH="$ROOT_DIR/../../aws-controllers-k8s/code-generator/scripts/build-controller.sh"

if [ ! -e $TARGET_SCRIPT_PATH ]; then
  echo "Could not find script located at $TARGET_SCRIPT_PATH"
  exit 1
fi

$TARGET_SCRIPT_PATH "$@"