#!/usr/bin/env bash

##############################################
# Tests for AWS SageMaker Training
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

# Create a random trainingjob name
# Expects RANDOM_ID env var to be set
setup_sagemaker_training_test_inputs() {
  if [[ -z ${RANDOM_ID+x} || -z ${AWS_REGION+x} || -z ${SAGEMAKER_DATA_BUCKET+x} ]]; then
    echo "[FAIL] Usage setup_sagemaker_training_test_inputs. Expects RANDOM_ID, SAGEMAKER_DATA_BUCKET and AWS_REGION to be set"
    exit 1
  fi

  TRAINING_JOB_NAME="ack-trainingjob-${RANDOM_ID}"
  # TrainingJobName must have length less than 64
  TRAINING_JOB_NAME=`echo $TRAINING_JOB_NAME|cut -c1-64`
  TRAINING_IMAGE_REGISTRY=`echo $(get_xgboost_registry $AWS_REGION)`
  TRAINING_IMAGE_USWEST2="$TRAINING_IMAGE_REGISTRY.dkr.ecr.$AWS_REGION.amazonaws.com/xgboost:1"
}

get_xgboost_training_yaml() {
  cat <<EOF
apiVersion: sagemaker.services.k8s.aws/v1alpha1
kind: TrainingJob
metadata:
  name: $TRAINING_JOB_NAME
spec:
  trainingJobName: $TRAINING_JOB_NAME
  roleARN: $SAGEMAKER_EXECUTION_ROLE_ARN
  hyperParameters:
    max_depth: "5"
    gamma: "4"
    eta: "0.2"
    min_child_weight: "6"
    silent: "0"
    objective: "multi:softmax"
    num_class: "10"
    num_round: "10"
  algorithmSpecification:
    trainingImage: $TRAINING_IMAGE_USWEST2
    trainingInputMode: File
  outputDataConfig:
    s3OutputPath: s3://$SAGEMAKER_DATA_BUCKET/sagemaker/training/output
  resourceConfig:
    instanceCount: 1
    instanceType: ml.m4.xlarge
    volumeSizeInGB: 5
  stoppingCondition:
    maxRuntimeInSeconds: 86400
  inputDataConfig:
    - channelName: train
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          s3URI: s3://$SAGEMAKER_DATA_BUCKET/sagemaker/training/train
          s3DataDistributionType: FullyReplicated
      contentType: text/csv
      compressionType: None
    - channelName: validation
      dataSource:
        s3DataSource:
          s3DataType: S3Prefix
          s3URI: s3://$SAGEMAKER_DATA_BUCKET/sagemaker/training/validation
          s3DataDistributionType: FullyReplicated
      contentType: text/csv
      compressionType: None
  tags:
    - key: key
      value: value
EOF
}

#################################################
# PRE-CHECKS
#################################################
sagemaker_trainingjob_prechecks() {
    setup_sagemaker_training_test_inputs
    if !is_aws_sagemaker_trainingjob_exists "$TRAINING_JOB_NAME"; then
        echo "[FAIL] expected $TRAINING_JOB_NAME to not exist in SageMaker. Did previous test run cleanup?"
        exit 1
    fi

    if k8s_resource_exists "$TRAINING_JOB_NAME"; then
        echo "[FAIL] expected $TRAINING_JOB_NAME to not exist. Did previous test run cleanup?"
        exit 1
    fi
}

#################################################
# Create training job
#################################################
# Assertions for trainingJob creation
# Parameter:
#   $1: training_job_name
verify_trainingjob_created() {
  if [[ "$#" -lt 1 ]]; then
    echo "[FAIL] Usage: verify_trainingjob_created training_job_name"
    exit 1
  fi

  local __training_job_name="$1"
  # Verify TrainingJob ARN was populated in Status
  local __k8s_trainingjob_arn=$(get_field_from_status "trainingjob/$__training_job_name" "ackResourceMetadata.arn")
  # Verify model exists on SageMaker and matches ARN populated in Status
  assert_aws_sagemaker_trainingjob_arn "$__training_job_name" "$__k8s_trainingjob_arn" || exit 1
  assert_aws_sagemaker_trainingjob_created "$__training_job_name" || exit 1
}

sagemaker_test_create_trainingjob() {
    setup_sagemaker_training_test_inputs
    debug_msg "Testing create trainingJob: $TRAINING_JOB_NAME."
    __yaml="$(get_xgboost_training_yaml)"
    echo "$__yaml" | kubectl apply -f -
    debug_msg "checking trainingJob $TRAINING_JOB_NAME created in SageMaker"
    verify_trainingjob_created "$TRAINING_JOB_NAME"
}

#################################################
# Stop (Delete) training job
#################################################
sagemaker_test_delete_trainingjob() {
    setup_sagemaker_training_test_inputs
    debug_msg "Testing Stopping TrainingJob: $TRAINING_JOB_NAME."
    kubectl delete TrainingJob/"$TRAINING_JOB_NAME" 2>/dev/null
    assert_equal "0" "$?" "FAILED: Expected success from kubectl delete but got $?" || exit 1
    assert_aws_sagemaker_trainingjob_stopped "$TRAINING_JOB_NAME" || exit 1
}

#################################################
# Post-Checks
#################################################
sagemaker_trainingjob_postchecks() {
    debug_msg "Check that the controller pod was not restarted."
    assert_pod_not_restarted $ack_ctrl_pod_id
}
