#!/usr/bin/env bash

##############################################
# Tests for AWS SageMaker
##############################################

set -u
set -x

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/aws/sagemaker.sh"

sagemaker_setup_common_test_resources

# Test Model
THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$THIS_DIR/helper/model/xgboost-model.sh"
sagemaker_test_create_model
sagemaker_test_delete_model
