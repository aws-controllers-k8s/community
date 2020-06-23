#!/usr/bin/env bash

# This run-integration-tests.sh is a copy of run-integration-tests.sh from amazon-vpc-cni-k8s package.
# Most of time taking commands are commented out currently and this file is used for POC of CircleCI run
# and AWS setup.

set -Euo pipefail

trap 'on_error $LINENO' ERR

DIR=$(cd "$(dirname "$0")"; pwd)
source "$DIR"/lib/common.sh
source "$DIR"/lib/aws.sh
source "$DIR"/lib/cluster.sh
source "$DIR"/lib/helm.sh
source "$DIR"/lib/k8s.sh

# Variables used in /lib/aws.sh
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
GO111MODULE=on

: "${AWS_DEFAULT_REGION:=us-west-2}"
: "${K8S_VERSION:=1.16.8}"
: "${PROVISION:=true}"
: "${DEPROVISION:=true}"
: "${BUILD:=true}"
: "${RUN_CONFORMANCE:=false}"
: "${HELM_LOCAL_REPO_NAME:=ack}"
# TODO: HELM_REPO_SOURCE will change to aws org.
: "${HELM_REPO_SOURCE:=https://vijtrip2.github.io/aws-service-operator-k8s/helm/charts/}"
: "${HELM_REPO_CHART_NAME:=start-all-service-controllers}"
: "${HELM_CONTROLLER_NAME_PREFIX:=aws-k8s}"
: "${HELM_LOCAL_CHART_NAME:=ack}"
: "${TEST_PASS:=0}"

__cluster_created=0
__cluster_deprovisioned=0

on_error() {
    # Make sure we destroy any cluster that was created if we hit run into an
    # error when attempting to run tests against the cluster
    if [[ $__cluster_created -eq 1 && $__cluster_deprovisioned -eq 0 && "$DEPROVISION" == true ]]; then
        # prevent double-deprovisioning with ctrl-c during deprovisioning...
        __cluster_deprovisioned=1
        echo "Cluster was provisioned already. Deprovisioning it..."
        down-test-cluster
    fi
    exit 1
}

# test specific config, results location
: "${TEST_ID:=$RANDOM}"
TEST_DIR=/tmp/ack-test/$(date "+%Y%M%d%H%M%S")-$TEST_ID
REPORT_DIR=${TEST_DIR}/report
TEST_CONFIG_DIR="$TEST_DIR/config"

# test cluster config location
# Pass in CLUSTER_ID to reuse a test cluster
: "${CLUSTER_ID:=$RANDOM}"
CLUSTER_NAME=ack-test-$CLUSTER_ID
TEST_CLUSTER_DIR=/tmp/ack-test/cluster-$CLUSTER_NAME
: "${CLUSTER_CONFIG:=${TEST_CLUSTER_DIR}/${CLUSTER_NAME}.yaml}"
: "${KUBECONFIG_PATH:=${TEST_CLUSTER_DIR}/kubeconfig}"

# shared binaries
: "${TESTER_DIR:=/tmp/aws-k8s-tester}"
: "${TESTER_PATH:=$TESTER_DIR/aws-k8s-tester}"
: "${KUBECTL_PATH:=$TESTER_DIR/kubectl}"

## Uncomment
LOCAL_GIT_VERSION=$(git describe --tags --always --dirty)
TEST_IMAGE_VERSION=${IMAGE_VERSION:-$LOCAL_GIT_VERSION}


#Install helm
install_helm

# double-check all our preconditions and requirements have been met
check_is_installed docker
check_is_installed aws
check_is_installed helm
check_aws_credentials
ensure_aws_k8s_tester
# Install controller-gen
ensure_controller_gen

: "${AWS_ACCOUNT_ID:=$(aws sts get-caller-identity --query Account --output text)}"
: "${AWS_ECR_REGISTRY:="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com"}"
: "${AWS_ECR_REPO_NAME:="ack"}"
: "${IMAGE_NAME:="$AWS_ECR_REGISTRY/$AWS_ECR_REPO_NAME"}"
: "${ROLE_CREATE:=true}"
: "${ROLE_ARN:=""}"

# S3 bucket initialization
: "${S3_BUCKET_CREATE:=true}"
: "${S3_BUCKET_NAME:=""}"

# `aws ec2 get-login` returns a docker login string, which we eval here to login to the ECR registry
# shellcheck disable=SC2046
eval $(aws ecr get-login --region $AWS_DEFAULT_REGION --no-include-email) >/dev/null 2>&1
ensure_ecr_repo "$AWS_ACCOUNT_ID" "$AWS_ECR_REPO_NAME"

echo "*******************************************************************************"
echo "Running $TEST_ID on $CLUSTER_NAME in $AWS_DEFAULT_REGION"
echo "+ Cluster config dir: $TEST_CLUSTER_DIR"
echo "+ Result dir:         $TEST_DIR"
echo "+ Tester:             $TESTER_PATH"
echo "+ Kubeconfig:         $KUBECONFIG_PATH"
echo "+ Cluster config:     $CLUSTER_CONFIG"
echo "+ AWS Account ID:     $AWS_ACCOUNT_ID"

mkdir -p "$TEST_DIR"
mkdir -p "$REPORT_DIR"
mkdir -p "$TEST_CLUSTER_DIR"
mkdir -p "$TEST_CONFIG_DIR"

if [[ "$PROVISION" == true ]]; then
    START=$SECONDS
    up-test-cluster
    UP_CLUSTER_DURATION=$((SECONDS - START))
    echo "TIMELINE: Upping test cluster took $UP_CLUSTER_DURATION seconds."
    __cluster_created=1
fi

export KUBECONFIG=$KUBECONFIG_PATH

# Cluster is setup at this point.
# Make sure not to exit the test-run without cleaning the cluster.
# Use should_execute in common.sh to short circuit methods if $TEST_PASS -eq 1

add_helm_repo
BASE_INTEGRATION_DURATION=0
if [[ "$TEST_PASS" -ne 0 ]]; then
  echo "NOTE: Skipping base test run because test is marked as failed"
elif [[ -z "${BASE_GIT_TAG+x}" ]]; then
  echo "NOTE: Skipping base test run because BASE_GIT_TAG is not given"
else
  echo "*******************************************************************************"
  echo "Running integration tests on BASE_GIT_TAG, $BASE_GIT_TAG"
  echo ""

  __ack_source_tmpdir="/tmp/ack-src-$BASE_GIT_TAG"
  echo "Checking out ack source code for $BASE_GIT_TAG ..."
  git clone git@github.com:varun1524/aws-service-operator-k8s.git "$__ack_source_tmpdir" || exit 1

  pushd "$__ack_source_tmpdir" || exit
  git checkout -b "$BASE_GIT_TAG" "$BASE_GIT_TAG"


  for d in ./services/*; do
    if [ -d "$d" ]; then
      __service_name=$(basename "$d")
      __test_crd_path="/tmp/crd/base/__service_name"

      echo "***************************************"
      echo "Running integration test on BASE_GIT_TAG $BASE_GIT_TAG for Service $__service_name"
      echo ""

      ensure_service_controller_running "$d" "$__service_name" "$BASE_GIT_TAG" "$__test_crd_path"

      START=$SECONDS
      pushd ./test/integration/services
      go test -v -timeout 0 "./$__service_name/..." --kubeconfig=$KUBECONFIG --ginkgo.skip="\[Disruptive\]" --ginkgo.randomizeAllSpecs --assets="./$__service_name/assets"
      TEST_PASS=$?
      popd
      BASE_INTEGRATION_DURATION=$((SECONDS - START))

      echo "TIMELINE: Integration Test took $BASE_INTEGRATION_DURATION seconds for service $__service_name on BASE_GIT_TAG $BASE_GIT_TAG."
      echo "***************************************"

      if [[ $TEST_PASS -ne 0 ]]; then
        break
      fi
    fi
  done

  echo "*******************************************************************************"
  echo "Integration Testing on BASE_GIT_TAG $BASE_GIT_TAG Finished"
  echo ""

  popd
fi

if [[ "$TEST_PASS" -ne 0 ]]; then
  echo "NOTE: Skipping latest test run because test is marked as failed"
else
  echo "*******************************************************************************"
  echo "Running integration tests on Latest Commit, $TEST_IMAGE_VERSION"
  echo ""

  __ack_source_tmpdir="/tmp/ack-src-$TEST_IMAGE_VERSION"
  echo "Checking out ack source code for $TEST_IMAGE_VERSION ..."
  git clone --depth=1 --branch e2etest git@github.com:varun1524/aws-service-operator-k8s.git "$__ack_source_tmpdir" || exit 1

  pushd "$__ack_source_tmpdir" || exit

  for d in ./services/*; do
    if [ -d "$d" ]; then
      __service_name=$(basename "$d")
      __test_crd_path="/tmp/crd/test/__service_name"

      echo "***************************************"
      echo "Running integration test on Latest Commit $TEST_IMAGE_VERSION for Service $__service_name."
      echo ""

      ensure_service_controller_running "$d" "$__service_name" "$TEST_IMAGE_VERSION" "$__test_crd_path"

      START=$SECONDS
      pushd ./test/integration/services
      go test -v -timeout 0 "./$__service_name/..." --kubeconfig=$KUBECONFIG --ginkgo.skip="\[Disruptive\]" --ginkgo.randomizeAllSpecs --assets="./$__service_name/assets"
      TEST_PASS=$?
      popd
      LATEST_INTEGRATION_DURATION=$((SECONDS - START))
      echo "TIMELINE: Integration tests on latest took $LATEST_INTEGRATION_DURATION seconds for service $__service_name."
      echo "***************************************"

      if [[ $TEST_PASS -ne 0 ]]; then
        break
      fi
    fi
  done

  echo "*******************************************************************************"
  echo "Integration Testing on Commit $TEST_IMAGE_VERSION Finished"
  echo ""

  popd

fi


if [[ "$DEPROVISION" == true ]]; then
    START=$SECONDS
    down-test-cluster

    DOWN_DURATION=$((SECONDS - START))
    echo "TIMELINE: Down processes took $DOWN_DURATION seconds."
    display_timelines
fi

if [[ $TEST_PASS -ne 0 ]]; then
    exit 1
fi
