
#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/testutil.sh
. $SCRIPTS_DIR/lib/aws.sh

service_name="sagemaker"

#################################################
# functions for tests
#################################################

# print_k8s_ack_controller_pod_logs prints kubernetes ack controller pod logs
# this function depends upon testutil.sh
print_k8s_ack_controller_pod_logs() {
  local ack_ctrl_pod_id=$( controller_pod_id )
  kubectl logs -n ack-system "$ack_ctrl_pod_id"
}

setup_sagemaker_common_test_inputs() {
  # uses non local variable for later use in tests
  sagemaker_data_bucket="REPLACE_ME"
  sagemaker_execution_role_arn="REPLACE_ME"
}

setup_sagemaker_common_test_inputs

assert_aws_sagemaker_model_created() {
  local model_name="$1"
  local model_arn="$(daws sagemaker describe-model --model-name "$model_name" | jq .ModelArn)"
  if [ -z "$model_arn" ]; then
    echo "ERROR: ModelArn not found for $model_name"
    return 1
  fi
  return 0
}