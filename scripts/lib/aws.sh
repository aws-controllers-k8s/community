#!/usr/bin/env bash

DEFAULT_AWS_CLI_VERSION="2.0.52"

check_aws_credentials() {
    aws sts get-caller-identity --query "Account" ||
        ( echo "No AWS credentials found. Please run \`aws configure\` to set up the CLI for your credentials." && exit 1)
}

ensure_ecr_repo() {
    local __registry_account_id="$1"
    local __repo_name="$2"
    if ! `aws ecr describe-repositories --registry-id "$__registry_account_id" --repository-names "$__repo_name" >/dev/null 2>&1`; then
        echo "creating ECR repo with name $__repo_name in registry account $__registry_account_id"
        aws ecr create-repository --repository-name "$__repo_name"
    fi
}

ensure_aws_k8s_tester() {
    TESTER_RELEASE=${TESTER_RELEASE:-v1.2.6}
    TESTER_DOWNLOAD_URL=https://github.com/aws/aws-k8s-tester/releases/download/$TESTER_RELEASE/aws-k8s-tester-$TESTER_RELEASE-$OS-$ARCH

    # Download aws-k8s-tester if not yet
    if [[ ! -e $TESTER_PATH ]]; then
        mkdir -p "$TESTER_DIR"
        echo "Downloading aws-k8s-tester from $TESTER_DOWNLOAD_URL to $TESTER_PATH"
        curl -s -L -X GET "$TESTER_DOWNLOAD_URL" -o "$TESTER_PATH"
        chmod +x "$TESTER_PATH"
    fi
}

ensure_ecr_image() {
  local __ack_service_image_tag="$1" #consist of ack-service_name-commit_sha
  local __dockerfile_path="$2"

  if `aws ecr describe-images --repository-name "$AWS_ECR_REPO_NAME" --image-ids imageTag="$__ack_service_image_tag" >/dev/null 2>&1`; then
     echo "ACK image $IMAGE_NAME:$__ack_service_image_tag already exists in repository. Skipping image build..."
  else
    START=$SECONDS
    echo "Building Docker image for $__ack_service_image_tag"
    docker build -t "$AWS_ECR_REPO_NAME":"$__ack_service_image_tag" -f "$__dockerfile_path" .
    docker tag "$AWS_ECR_REPO_NAME":"$__ack_service_image_tag" "$IMAGE_NAME":"$__ack_service_image_tag"
    docker push "$IMAGE_NAME":"$__ack_service_image_tag"
    echo "pushed successfully to ECR"

    DOCKER_BUILD_DURATION=$((SECONDS - START))
    echo "TIMELINE: Docker build took $DOCKER_BUILD_DURATION seconds."
  fi
}

# generate_aws_temp_creds function will generate temporary AWS CREDENTIALS which are valid for 900 seconds
generate_aws_temp_creds() {
  __uuid=$(uuidgen | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')

  if [ -z "$AWS_ROLE_ARN" ]; then
    printf "Missing input Role ARN, exiting...\n"
    exit 1
  fi

  printf "Running aws sts assume-role --role-arn %s --role-session-name tmp-role-%s --duration-seconds 900", "$AWS_ROLE_ARN", "$__uuid "

  JSON=$(aws sts assume-role \
           --role-arn "$AWS_ROLE_ARN"  \
           --role-session-name tmp-role-"$__uuid" \
           --duration-seconds 900 \
           --output json || exit 1)

      AWS_ACCESS_KEY_ID=$(echo "${JSON}" | jq --raw-output ".Credentials[\"AccessKeyId\"]")
      AWS_SECRET_ACCESS_KEY=$(echo "${JSON}" | jq --raw-output ".Credentials[\"SecretAccessKey\"]")
      AWS_SESSION_TOKEN=$(echo "${JSON}" | jq --raw-output ".Credentials[\"SessionToken\"]")
      printf "\nTemporary credentials generated\n"
}

aws_account_id() {
    JSON=$(aws sts get-caller-identity --output json || exit 1)
    echo "${JSON}" | jq --raw-output ".Account"
}

# daws() executes the AWS Python CLI tool from a Docker container.
#
# Instead of relying on developers having a particular version of the AWS
# Python CLI tool, this method allows a specific version of the CLI tool to be
# executed within a Docker container.
#
# You call the daws function just like you were calling the `aws` CLI tool.
#
# Usage:
#
#   daws SERVICE COMMAND [OPTIONS]
#
# Example:
#
#   daws ecr describe-repositories --repository-name my-repo
#
# To use a specific version of the AWS CLI, set the ACK_AWS_CLI_IMAGE_VERSION
# environment variable, otherwise the value of DEFAULT_AWS_CLI_VERSION is used.
daws() {
    aws_cli_img_version=${ACK_AWS_CLI_IMAGE_VERSION:-$DEFAULT_AWS_CLI_VERSION}
    aws_cli_img="amazon/aws-cli:$aws_cli_img_version"
    docker run --rm -it -v ~/.aws:/root/.aws "$aws_cli_img" "$@"
}
