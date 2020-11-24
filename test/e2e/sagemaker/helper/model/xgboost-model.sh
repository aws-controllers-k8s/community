#!/usr/bin/env bash

##############################################
# Tests for AWS SageMaker Model
##############################################

set -u
set -x

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/sagemaker.sh"

# Create a random model name
# Expects RANDOM_ID env var to be set
setup_sagemaker_model_test_inputs() {
  if [[ -z ${RANDOM_ID+x} ]]; then
    echo "[FAIL] Usage: setup_sagemaker_model_test_inputs. Expects RANDOM_ID to be set"
    exit 1
  fi
  XGBOOST_MODEL_NAME="xgboost-model-${RANDOM_ID}"
  # model name must have length less than 64
  XGBOOST_MODEL_NAME=`echo $XGBOOST_MODEL_NAME|cut -c1-64`
}

# Returns a Model Spec based on XGBOOST_MODEL_NAME, SAGEMAKER_DATA_BUCKET, AWS_REGION
get_xgboost_model_yaml() {
  cat <<EOF
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: Model
metadata:
  name: $XGBOOST_MODEL_NAME
spec:
  modelName: $XGBOOST_MODEL_NAME
  primaryContainer:
    containerHostname: xgboost
    modelDataURL: s3://$SAGEMAKER_DATA_BUCKET/sagemaker/model/xgboost-mnist-model.tar.gz
    image: $(get_xgboost_registry $AWS_REGION).dkr.ecr.$AWS_REGION.amazonaws.com/xgboost:latest
  executionRoleARN: $SAGEMAKER_EXECUTION_ROLE_ARN
  tags:
    - key: key
      value: value
EOF
}

#################################################
# create model
#################################################

# Assertions for model creation
# Parameter:
#   $1: model_name
verify_model_created() {
  if [[ "$#" -lt 1 ]]; then
    echo "[FAIL] Usage: verify_model_created model_name"
    exit 1
  fi

  local __model_name="$1"
  # Verify Model ARN was populated in Status
  local __k8s_model_arn=$(get_field_from_status "model/$__model_name" "ackResourceMetadata.arn")
  # Verify model exists on SageMaker and matches ARN populated in Status
  assert_aws_sagemaker_model_arn "$__model_name" "$__k8s_model_arn" || exit 1
}

sagemaker_test_create_model() {
  setup_sagemaker_model_test_inputs
  debug_msg "Testing create model: $XGBOOST_MODEL_NAME."
  echo "$(get_xgboost_model_yaml)" | kubectl apply -f -
  verify_model_created $XGBOOST_MODEL_NAME
}

#################################################
# delete model
#################################################
sagemaker_test_delete_model() {
  debug_msg "Testing delete model: $XGBOOST_MODEL_NAME."
  kubectl delete Model/"$XGBOOST_MODEL_NAME" 2>/dev/null
  assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1
}