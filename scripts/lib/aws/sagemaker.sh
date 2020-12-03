
#!/usr/bin/env bash

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
  ack_ctrl_pod_id=$( controller_pod_id )
}

# print_k8s_ack_controller_pod_logs prints kubernetes ack controller pod logs
# this function depends upon testutil.sh
print_k8s_ack_controller_pod_logs() {
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
  kubectl delete -n "$__namespace" trainingjob --all
}

# Create an IAM Role for SageMaker to assume on your behalf.
# Parameter:
#   $1: role_name
sagemaker_create_execution_role() {
  if [[ "$#" -lt 1 ]]; then
    echo "[FAIL] Usage: sagemaker_create_execution_role role_name"
    exit 1
  fi

  local __role_name="$1"
  daws iam create-role --role-name "$__role_name" --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "sagemaker.amazonaws.com"}, "Action": "sts:AssumeRole"}]}' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam create-role --role-name $__role_name, got $?" || exit 1

  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonSageMakerFullAccess' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name, got $?" || exit 1
  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AmazonS3FullAccess' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name, got $?" || exit 1
  
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
  daws s3 sync s3://$S3_DATA_SOURCE_BUCKET s3://$__bucket_name --only-show-errors 
  assert_equal "0" "$?" "Expected success from aws s3 sync s3://$S3_DATA_SOURCE_BUCKET s3://$__bucket_name but got $?" || exit 1
}

# Delete an S3 bucket
# Parameter:
#   $1: bucket_name
sagemaker_delete_data_bucket() {
  local __bucket_name="$1"
  aws s3 rm --recursive s3://$__bucket_name --only-show-errors
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
    debug_msg "[ERROR] ModelArn did not match/not found for $model_name"
    return 1
  fi

  return 0
}

# is_aws_sagemaker_trainingjob_exists() returns 0 if an TrainingJob with the supplied name
# exists, 1 otherwise.
#
# training_job_created TRAINING_JOB_NAME
# Arguments:
#
#   TRAINING_JOB_NAME  required string for the name of the bucket to check
#
# Usage:
#
#   if ! training_job_created "$training_job_name"; then
#       echo "Training Job $training_job_name does not exist!"
#   fi
is_aws_sagemaker_trainingjob_exists() {
  if [[ "$#" -lt 1 || -z ${AWS_REGION+x} ]]; then
    echo "[FAIL] Usage: is_aws_sagemaker_trainingjob_exists training_job_name. Expects AWS_REGION to be set"
    exit 1
  fi
  local __training_job_name="$1"

  local __training_job_status="$(daws sagemaker describe-training-job --region "$AWS_REGION" --training-job-name $__training_job_name --output json | jq .TrainingJobStatus)"

  if [[ $? -eq 0 ]]; then
      echo "$__training_job_name found!"
      return 0
  else
      echo "$__training_job_name not found!"
      return 1
  fi
}

# assert_aws_sagemaker_trainingjob_arn() verifies the SageMaker ARN of the TrainingJob with the supplied name.
#
# assert_aws_sagemaker_trainingjob_arn TRAINING_JOB_NAME
# Arguments:
#
#   TRAINING_JOB_NAME  required string for the name of the trainingJob to check
#   EXPECTED_TRAINING_JOB_ARN The expected training job arn from the k8s status
assert_aws_sagemaker_trainingjob_arn() {
  if [[ "$#" -lt 1 || -z ${AWS_REGION+x} ]]; then
    echo "[FAIL] Usage: assert_aws_sagemaker_trainingjob_arn training_job_name expected_training_job_arn. Expects AWS_REGION to be set"
    exit 1
  fi

  local __training_job_name="$1"
  local __expected_training_job_arn="$2"
  local training_job_response=$(daws sagemaker describe-training-job --region "$AWS_REGION" --training-job-name $__training_job_name)
  local __aws_training_job_arn=$(echo "$training_job_response" | jq -r '.TrainingJobArn')
  if [[ ( -z "$__aws_training_job_arn" || $__expected_training_job_arn != $__aws_training_job_arn ) ]]; then
    debug_msg "[ERROR] TrainingJobArn did not match/not found for $__training_job_name"
    return 1
  fi

  return 0
}

# get_aws_sagemaker_trainingjob_status() prints the SageMaker status of the TrainingJob with the supplied name.
# If the job is not found, it returns a 1.
#
# get_aws_sagemaker_trainingjob_status TRAINING_JOB_NAME
# Arguments:
#
#   TRAINING_JOB_NAME  required string for the name of the trainingJob to check
#
# Usage:
#
#   local training_job_status=$(get_aws_sagemaker_trainingjob_status "${__training_job_name}")
get_aws_sagemaker_trainingjob_status() {
  if [[ "$#" -lt 1 || -z ${AWS_REGION+x} ]]; then
    echo "[FAIL] Usage: get_aws_sagemaker_trainingjob_status training_job_name. Expects AWS_REGION to be set"
    exit 1
  fi

  local __training_job_name="$1"
  local __training_job_status="$(daws sagemaker describe-training-job --region "$AWS_REGION" --training-job-name $__training_job_name --output json | jq .TrainingJobStatus)"

  if [ -z "${__training_job_status}" ]; then
    echo "[ERROR] trainingJob $__training_job_status not found"
    return 1
  else
    echo "${__training_job_status}"
    return 0
  fi
}

# assert_aws_sagemaker_trainingjob_created() checks if the SageMaker status of the TrainingJob with the supplied name is in creating or created.
# If not, it returns a 1.
#
# assert_aws_sagemaker_trainingjob_created TRAINING_JOB_NAME
# Arguments:
#
#   TRAINING_JOB_NAME  required string for the name of the trainingJob to check
#
# Usage:
#
#   if ! assert_aws_sagemaker_trainingjob_created "$training_job_name"; then
#       echo "Training Job $training_job_name was not created, Check logs!"
#   fi
assert_aws_sagemaker_trainingjob_created() {
  local __training_job_name="$1"
  local __training_job_status=$(echo $(get_aws_sagemaker_trainingjob_status "${__training_job_name}"))
   
  if [[ ( \"InProgress\" = $__training_job_status || \"Completed\" = $__training_job_status ) ]]; then
    debug_msg "[SUCCESS] TrainingJob $__training_job_name created and has status ${__training_job_status}"
    return 0
  else
    debug_msg "[ERROR] TrainingJob $__training_job_name was not created and has status ${__training_job_status}"
    return 1
  fi
}

# assert_aws_sagemaker_trainingjob_stopped() checks if the SageMaker status of the TrainingJob with the supplied name is in creating or created.
# If not, it returns a 1.
#
# assert_aws_sagemaker_trainingjob_stopped TRAINING_JOB_NAME
# Arguments:
#
#   TRAINING_JOB_NAME  required string for the name of the trainingJob to check
#
# Usage:
#
#   if ! assert_aws_sagemaker_trainingjob_stopped "$training_job_name"; then
#       echo "Training Job $training_job_name was not stopped, Check logs!"
#   fi
assert_aws_sagemaker_trainingjob_stopped() {
  local __training_job_name="$1"
  local __training_job_status=$(get_aws_sagemaker_trainingjob_status "${__training_job_name}")  
  
  # TODO: should the first condition be considered a failure ? 
  if [[ \"Completed\" = $__training_job_status ]]; then
    debug_msg "[WARNING] TrainingJob $__training_job_status has ${__training_job_status} status, cannot stop now."
    return 0
  elif [[ ( \"Stopping\" = $__training_job_status || \"Stopped\" = $__training_job_status ) ]]; then
    debug_msg "[SUCCESS] TrainingJob $__training_job_name stopped and has status ${__training_job_status}"
    return 0
  else 
    debug_msg "[ERROR] TrainingJob $__training_job_name was not stopped and has status ${__training_job_status}"
    return 1
  fi
}
