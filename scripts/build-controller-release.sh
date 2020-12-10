#!/usr/bin/env bash

# A script that builds release artifacts for a single ACK service controller
# for an AWS service API

set -eo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
BIN_DIR="$ROOT_DIR/bin"
TEMPLATES_DIR="$ROOT_DIR/templates"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/helm.sh"

check_is_installed controller-gen "You can install controller-gen with the helper scripts/install-controller-gen.sh"
check_is_installed helm "You can install Helm with the helper scripts/install-helm.sh"

if ! k8s_controller_gen_version_equals "$CONTROLLER_TOOLS_VERSION"; then
    echo "FATAL: Existing version of controller-gen "`controller-gen --version`", required version is $CONTROLLER_TOOLS_VERSION."
    echo "FATAL: Please uninstall controller-gen and install the required version with scripts/install-controller-gen.sh."
    exit 1
fi

: "${ACK_GENERATE_CACHE_DIR:=~/.cache/aws-controllers-k8s}"
: "${ACK_GENERATE_BIN_PATH:=$BIN_DIR/ack-generate}"
: "${ACK_GENERATE_API_VERSION:="v1alpha1"}"
: "${ACK_GENERATE_CONFIG_PATH:=""}"
: "${ACK_GENERATE_OUTPUT_PATH:=""}"
: "${ACK_GENERATE_IMAGE_REPOSITORY:="amazon/aws-controllers-k8s"}"

USAGE="
Usage:
  $(basename "$0") <service> <release_version>

<service> should be an AWS service API aliases that you wish to build -- e.g.
's3' 'sns' or 'sqs'

<release_version> should be the SemVer version tag for the release -- e.g.
'v0.1.3'

Environment variables:
  ACK_GENERATE_CACHE_DIR                Overrides the directory used for caching
                                        AWS API models used by the ack-generate
                                        tool.
                                        Default: $ACK_GENERATE_CACHE_DIR
  ACK_GENERATE_BIN_PATH:                Overrides the path to the the ack-generate
                                        binary.
                                        Default: $ACK_GENERATE_BIN_PATH
  ACK_GENERATE_CONFIG_PATH:             Specify a path to the generator config YAML
                                        file to instruct the code generator for the
                                        service.
                                        Default: services/{SERVICE}/generator.yaml
  ACK_GENERATE_OUTPUT_PATH:             Specify a path for the generator to output
                                        to.
                                        Default: services/{SERVICE}
  ACK_GENERATE_IMAGE_REPOSITORY:        Specify a Docker image repository to use
                                        for release artifacts
                                        Default: amazon/aws-controllers-k8s
  ACK_GENERATE_SERVICE_ACCOUNT_NAME:    Name of the Kubernetes Service Account and
                                        Cluster Role to use in Helm chart.
                                        Default: $ACK_GENERATE_SERVICE_ACCOUNT_NAME
  K8S_RBAC_ROLE_NAME:                   Name of the Kubernetes Role to use when
                                        generating the RBAC manifests for the
                                        custom resource definitions.
                                        Default: $K8S_RBAC_ROLE_NAME
"

if [ $# -ne 2 ]; then
    echo "ERROR: $(basename "$0") accepts exactly two parameters, the SERVICE and the RELEASE_VERSION" 1>&2
    echo "$USAGE"
    exit 1
fi

if [ ! -f $ACK_GENERATE_BIN_PATH ]; then
    if is_installed "ack-generate"; then
        ACK_GENERATE_BIN_PATH=$(which "ack-generate")
    else
        echo "ERROR: Unable to find an ack-generate binary.
Either set the ACK_GENERATE_BIN_PATH to a valid location or
run:
 
   make build-ack-generate
 
from the root directory or install ack-generate using:

   go get -u github.com/aws/aws-controllers-k8s/cmd/ack-generate" 1>&2
        exit 1;
    fi
fi
SERVICE=$(echo "$1" | tr '[:upper:]' '[:lower:]')
RELEASE_VERSION="$2"
: "${K8S_RBAC_ROLE_NAME:="ack-$SERVICE-controller"}"
: "${ACK_GENERATE_SERVICE_ACCOUNT_NAME:="ack-$SERVICE-controller"}"

# If there's a generator.yaml in the service's directory and the caller hasn't
# specified an override, use that.
if [ -z "$ACK_GENERATE_CONFIG_PATH" ]; then
    if [ -f "$ROOT_DIR/services/$SERVICE/generator.yaml" ]; then
        ACK_GENERATE_CONFIG_PATH="$ROOT_DIR/services/$SERVICE/generator.yaml"
    fi
fi

helm_output_dir="$ROOT_DIR/services/$SERVICE/helm"
ag_args="$SERVICE $RELEASE_VERSION"
if [ -n "$ACK_GENERATE_CACHE_DIR" ]; then
    ag_args="$ag_args --cache-dir $ACK_GENERATE_CACHE_DIR"
fi
if [ -n "$ACK_GENERATE_OUTPUT_PATH" ]; then
    ag_args="$ag_args --output $ACK_GENERATE_OUTPUT_PATH"
    helm_output_dir="$ACK_GENERATE_OUTPUT_PATH/helm"
fi
if [ -n "$ACK_GENERATE_CONFIG_PATH" ]; then
    ag_args="$ag_args --generator-config-path $ACK_GENERATE_CONFIG_PATH"
fi
if [ -n "$ACK_GENERATE_IMAGE_REPOSITORY" ]; then
    ag_args="$ag_args --image-repository $ACK_GENERATE_IMAGE_REPOSITORY"
fi
if [ -n "$ACK_GENERATE_SERVICE_ACCOUNT_NAME" ]; then
    ag_args="$ag_args --service-account-name $ACK_GENERATE_SERVICE_ACCOUNT_NAME"
fi

echo "Building release artifacts for $SERVICE-$RELEASE_VERSION"
$ACK_GENERATE_BIN_PATH release $ag_args

pushd $ROOT_DIR/services/$SERVICE/apis/$ACK_GENERATE_API_VERSION 1>/dev/null

echo "Generating custom resource definitions for $SERVICE"
controller-gen crd:allowDangerousTypes=true paths=./... output:crd:artifacts:config=$helm_output_dir/crds

popd 1>/dev/null

pushd $ROOT_DIR/services/$SERVICE/pkg/resource 1>/dev/null

echo "Generating RBAC manifests for $SERVICE"
controller-gen rbac:roleName=$K8S_RBAC_ROLE_NAME paths=./... output:rbac:artifacts:config=$helm_output_dir/templates
# controller-gen rbac outputs a ClusterRole definition in a
# $config_output_dir/rbac/role.yaml file. We have some other standard Role
# files for a reader and writer role, so here we rename the `role.yaml` file to
# `cluster-role-controller.yaml` to better reflect what is in that file.
mv $helm_output_dir/templates/role.yaml $helm_output_dir/templates/cluster-role-controller.yaml

popd 1>/dev/null
