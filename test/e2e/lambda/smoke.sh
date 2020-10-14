#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/aws.sh"
source "$SCRIPTS_DIR/lib/aws/lambda.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

wait_seconds=10
test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="lambda"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

function_name="ack-test-smoke-$service_name-$x" # TODO
resource_name="lambdas/$function_name"

if lambda_function_exists "$function_name"; then
    echo "FAIL: expected $function_name to not exist in Lambda. Did previous test run cleanup?"
    exit 1
fi

if k8s_resource_exists "$resource_name"; then
    echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
    exit 1
fi

# REQUIREMENTS

bucket_name="ack-test-smoke-$service_name-$AWS_ACCOUNT_ID"

# Verify that the lambda function is already in S3

# API CALL

# COMPILE-DEPLOY



# TEST ACTIONS and ASSERTIONS

# Create the function
cat <<EOF | kubectl apply -f -
apiVersion: lambda.services.k8s.aws/v1alpha1
kind: Function
metadata:
  name: $function_name
  namespace: default
spec:
  code:
    # s3 bucket where the lambda function is deployed
    s3Bucket: $bucket_name
    # lambda function binary/script filename
    s3Key: lambda.zip
  functionName: test-ack-lambda-function
  # Lambda handler
  handler: lambda
  # Lambda role arn
  role: arn:aws:iam::771174509839:role/lambda-no-permissions 
  # Lambda function runtime
  runtime: go1.x
EOF

sleep $wait_seconds

# Check the lambda function was created


# delete lambda function


# Check lambda function was deleted