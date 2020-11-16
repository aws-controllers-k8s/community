#!/usr/bin/env bash

##############################################
# Tests for AWS SageMaker Model
##############################################

set -u
set -x

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"
source "$SCRIPTS_DIR/lib/aws/sagemaker.sh"

test_name="$( filenoext "${BASH_SOURCE[0]}" )"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

setup_sagemaker_model_test_inputs() {
  xgboost_model_name="xgboost-model-"$(date '+%Y-%m-%d-%H-%M-%S')""
}

setup_sagemaker_model_test_inputs
assert_equal "us-west-2" "$AWS_REGION" "Expected $AWS_REGION to be us-west-2" || exit 1

get_xgboost_model_yaml() {
  cat <<EOF
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: Model
metadata:
  name: $xgboost_model_name
spec:
  modelName: $xgboost_model_name
  primaryContainer:
    containerHostname: xgboost
    modelDataURL: s3://$sagemaker_data_bucket/inference/xgboost-mnist/model.tar.gz
    image: 433757028032.dkr.ecr.us-west-2.amazonaws.com/xgboost:latest
  executionRoleARN: $sagemaker_execution_role_arn
  tags:
    - key: key 
      value: value
EOF
}

# k8s_controller_reload_credentials "$service_name"

#################################################
# create model
#################################################
ack_create_model() {
  debug_msg "Testing create model: $xgboost_model_name."
  __yaml="$(get_xgboost_model_yaml)"
  echo "$__yaml" | kubectl apply -f -
}
ack_create_model

# TODO: remove sleep wait for model cr to have ARN
sleep 10
assert_aws_sagemaker_model_created "$xgboost_model_name"|| exit 1

#################################################
# delete model
#################################################
debug_msg "Testing delete model: $xgboost_model_name."
kubectl delete Model/"$xgboost_model_name" 2>/dev/null
assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

