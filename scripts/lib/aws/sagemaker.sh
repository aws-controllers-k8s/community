
#!/usr/bin/env bash
set -Eeuxo pipefail

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/testutil.sh
. $SCRIPTS_DIR/lib/aws.sh

#################################################
# functions for tests
#################################################

# Initializes global variables to be used by tests later
init_sagemaker_test_vars() {
  S3_DATA_SOURCE_BUCKET="source-data-bucket-718865417152-us-west-2"
  RANDOM_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
  AWS_SERVICE="sagemaker"
  CLEANUP_EXECUTION_ROLE_ARN=false
  CLEANUP_DATA_BUCKET=false
}

# print_k8s_ack_controller_pod_logs prints kubernetes ack controller pod logs
# this function depends upon testutil.sh
print_k8s_ack_controller_pod_logs() {
  local ack_ctrl_pod_id=$( controller_pod_id )
  kubectl logs -n ack-system "$ack_ctrl_pod_id"
}

# Cleans up all k8s resources created during tests
cleanup() {
  # We want to run every command in this function, even if some fail.
  set +e
  delete_all_resources
  if [[ ("$CLEANUP_EXECUTION_ROLE_ARN" = true) ]]; then
    sagemaker_delete_execution_role $SAGEMAKER_EXECUTION_ROLE_NAME
  fi
  if [[ "$CLEANUP_DATA_BUCKET" = true ]]; then
    sagemaker_delete_data_bucket $SAGEMAKER_DATA_BUCKET
  fi
  print_k8s_ack_controller_pod_logs
}

trap cleanup EXIT

# Cleans up all k8s resources created during tests
# Parameter:
#   $1: namespace of CR
function delete_all_resources()
{
  local __namespace="${1:-default}"
  kubectl delete -n "$__namespace" model --all  
}

# Create a IAM Role for SageMaker to assume on your behalf.
# Parameter:
#   $1: role_name
sagemaker_create_execution_role() {
  if [[ "$#" -lt 1 ]]; then
    echo "[FAIL] Usage: sagemaker_create_execution_role role_name"
    exit 1
  fi

  local __role_name="$1"
  daws iam create-role --role-name "$__role_name" --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "sagemaker.amazonaws.com"}, "Action": "sts:AssumeRole"}]}' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam create-role --role-name $__role_name but got $?" || exit 1

  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonSageMakerFullAccess' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name but got $?" || exit 1
  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonS3FullAccess' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name but got $?" || exit 1
  
  local __role_arn=$(daws iam get-role --role-name "$__role_name" | jq -r ".Role.Arn")
  echo "$__role_arn"
}

# Delete IAM role created for SageMaker tests.
# Parameter:
#   $1: role_name
sagemaker_delete_execution_role() {
  if [[ "$#" -lt 1 ]]; then
    echo "[FAIL] Usage: sagemaker_delete_execution_role role_name"
    exit 1
  fi

  local __role_name="$1"
  daws iam detach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonS3FullAccess' >/dev/null
  daws iam detach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonSageMakerFullAccess' >/dev/null

  daws iam delete-role --role-name "$__role_name" >/dev/null
}

# Create a S3 bucket for dataset used in SageMaker tests
# Expects S3_DATA_SOURCE_BUCKET and AWS_REGION env vars to be set
# Parameter:
#   $1: bucket_name
sagemaker_create_data_bucket() {
  if [[ "$#" -lt 1 || -z ${AWS_REGION+x} || -z ${S3_DATA_SOURCE_BUCKET+x} ]]; then
    echo "[FAIL] Usage: sagemaker_create_data_bucket bucket_name. Expects S3_DATA_SOURCE_BUCKET and AWS_REGION to be set"
    exit 1
  fi

  local __bucket_name="$1"
  if [[ $AWS_REGION != "us-east-1" ]]; then
    daws s3api create-bucket --bucket "$__bucket_name" --region "$AWS_REGION" --create-bucket-configuration LocationConstraint="$AWS_REGION"
  else
    daws s3api create-bucket --bucket "$__bucket_name" --region "$AWS_REGION"
  fi

  assert_equal "0" "$?" "Expected success from aws s3api create-bucket --bucket $__bucket_name but got $?" || exit 1
  daws s3 sync s3://$S3_DATA_SOURCE_BUCKET s3://$__bucket_name
  assert_equal "0" "$?" "Expected success from aws s3 sync s3://$S3_DATA_SOURCE_BUCKET s3://$__bucket_name but got $?" || exit 1
}

# Delete an S3 bucket
# Parameter:
#   $1: bucket_name
sagemaker_delete_data_bucket() {
  local __bucket_name="$1"
  aws s3 rb s3://$__bucket_name --force
}

# Generate a random bucket name
# Expects AWS_ACCOUNT_ID, RANDOM_ID and AWS_REGION env vars to be set
generate_s3_bucket_name() {
  if [[ -z ${AWS_REGION+x} || -z ${AWS_ACCOUNT_ID+x} || -z ${RANDOM_ID+x} ]]; then
    echo "[FAIL] Usage: generate_s3_bucket_name. Expects AWS_ACCOUNT_ID, RANDOM_ID and AWS_REGION to be set"
    exit 1
  fi

  local __bucket_name="data-bucket-${AWS_REGION}-${AWS_ACCOUNT_ID}-${RANDOM_ID}"
  # bucket name must have length less than 63
  echo ${__bucket_name:0:62}
}

# Generate a random iam role name
# Expects RANDOM_ID env var to be set
generate_iam_role_name() {
  if [[ -z ${RANDOM_ID+x} ]]; then
    echo "[FAIL] Usage: generate_iam_role_name. Expects RANDOM_ID and AWS_REGION to be set"
    exit 1
  fi

  local __role_name="sagemaker-execution-role-${RANDOM_ID}"
  # bucket name must have length less than 64
  echo ${__role_name:0:63}
}

# Setup resources to be used by tests
sagemaker_setup_common_test_resources() {
  init_sagemaker_test_vars

  if [[ -z ${SAGEMAKER_DATA_BUCKET+x} ]]; then
    debug_msg "Creating new S3 bucket . . ."
    SAGEMAKER_DATA_BUCKET=$(generate_s3_bucket_name)
    sagemaker_create_data_bucket $SAGEMAKER_DATA_BUCKET
    CLEANUP_DATA_BUCKET=true
  else
    debug_msg "using existing S3 bucket for tests: ${SAGEMAKER_DATA_BUCKET}"
  fi

  if [[ -z ${SAGEMAKER_EXECUTION_ROLE_ARN+x} ]]; then
    debug_msg "Creating new IAM role . . ."
    SAGEMAKER_EXECUTION_ROLE_NAME=$(generate_iam_role_name)
    SAGEMAKER_EXECUTION_ROLE_ARN=$(sagemaker_create_execution_role "$SAGEMAKER_EXECUTION_ROLE_NAME")
    CLEANUP_EXECUTION_ROLE_ARN=true
  else
    debug_msg "using existing role: ${SAGEMAKER_EXECUTION_ROLE_ARN}"
  fi

}

# Get SageMaker built in xgboost image registry based on region
get_xgboost_registry() {
  local __region="$1"
  declare -A registry_map=( ["us-east-1"]="811284229777" ["us-west-2"]="433757028032" ["eu-west-1"]="685385470294" ["us-east-2"]="825641698319")
  echo "${registry_map[$__region]}"
}

# Verify model created in SageMaker and ARN matches the CR
# Expects AWS_REGION env var to be set
# Parameter:
#   $1: model_name to describe
#   $2: expected_model_arn
assert_aws_sagemaker_model_arn() {
  if [[ "$#" -lt 1 || -z ${AWS_REGION+x} ]]; then
    echo "[FAIL] Usage: assert_aws_sagemaker_model_arn model_name expected_model_arn. Expects AWS_REGION to be set"
    exit 1
  fi

  local model_name="$1"
  local expected_model_arn="$2"
  local model_response=$(daws sagemaker describe-model --model-name "$model_name" --region "$AWS_REGION")
  local aws_model_arn=$(echo "$model_response" | jq -r '.ModelArn')
  if [[ ( -z "$aws_model_arn" || "$expected_model_arn" != "$aws_model_arn" ) ]]; then
    debug_msg "ERROR: ModelArn did not match/not found for $model_name"
    return 1
  fi

  return 0
}