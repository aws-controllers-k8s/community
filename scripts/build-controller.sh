#!/usr/bin/env bash

# A script that builds a single ACK service controller for an AWS service API

set -eo pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$SCRIPTS_DIR/.."
BIN_DIR="$ROOT_DIR/bin"
TEMPLATES_DIR="$ROOT_DIR/templates"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"

check_is_installed controller-gen "You can install controller-gen with the helper scripts/install-controller-gen.sh"

if ! k8s_controller_gen_version_equals "$CONTROLLER_TOOLS_VERSION"; then
    echo "FATAL: Existing version of controller-gen "`controller-gen --version`", required version is $CONTROLLER_TOOLS_VERSION."
    echo "FATAL: Please uninstall controller-gen and install the required version with scripts/install-controller-gen.sh."
    exit 1
fi

ACK_GENERATE_CACHE_DIR=${ACK_GENERATE_CACHE_DIR:-"~/.cache/aws-controllers-k8s"}
ACK_GENERATE_BIN_PATH=${ACK_GENERATE_BIN_PATH:-"$BIN_DIR/ack-generate"}
ACK_GENERATE_API_VERSION=${ACK_GENERATE_API_VERSION:-"v1alpha1"}
ACK_GENERATE_CONFIG_PATH=${ACK_GENERATE_CONFIG_PATH:-""}

USAGE="
Usage:
  $(basename "$0") <service>

<service> should be an AWS service API aliases that you wish to build -- e.g.
's3' 'sns' or 'sqs'

Environment variables:
  ACK_GENERATE_CACHE_DIR    Overrides the directory used for caching AWS API
                            models used by the ack-generate tool.
                            Default: $ACK_GENERATE_CACHE_DIR
  ACK_GENERATE_BIN_PATH:    Overrides the path to the the ack-generate binary.
                            Default: $ACK_GENERATE_BIN_PATH
  ACK_GENERATE_API_VERSION: Overrides the version of the Kubernetes API objects
                            generated by the ack-generate apis command. If not
                            specified, and the service controller has been
                            previously generated, the latest generated API
                            version is used. If the service controller has yet
                            to be generated, 'v1alpha1' is used.
  ACK_GENERATE_CONFIG_PATH: Specify a path to the generator config YAML file to
                            instruct the code generator for the service.
                            Default: services/{SERVICE}/generator.yaml
  K8S_RBAC_ROLE_NAME:       Name of the Kubernetes Role to use when generating
                            the RBAC manifests for the custom resource
                            definitions.
                            Default: $K8S_RBAC_ROLE_NAME
"

if [ $# -ne 1 ]; then
    echo "ERROR: $(basename "$0") only accepts a single parameter" 1>&2
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

K8S_RBAC_ROLE_NAME=${K8S_RBAC_ROLE_NAME:-"ack-$SERVICE-controller"}

# If there's a generator.yaml in the service's directory and the caller hasn't
# specified an override, use that.
if [ -z "$ACK_GENERATE_CONFIG_PATH" ]; then
    if [ -f "$ROOT_DIR/services/$SERVICE/generator.yaml" ]; then
        ACK_GENERATE_CONFIG_PATH="$ROOT_DIR/services/$SERVICE/generator.yaml"
    fi
fi

ag_args="$SERVICE"
if [ -n "$ACK_GENERATE_CACHE_DIR" ]; then
    ag_args="$ag_args --cache-dir $ACK_GENERATE_CACHE_DIR"
fi

apis_args="apis $ag_args"
if [ -n "$ACK_GENERATE_API_VERSION" ]; then
    apis_args="$apis_args --version $ACK_GENERATE_API_VERSION"
fi

if [ -n "$ACK_GENERATE_CONFIG_PATH" ]; then
    ag_args="$ag_args --generator-config-path $ACK_GENERATE_CONFIG_PATH"
    apis_args="$apis_args --generator-config-path $ACK_GENERATE_CONFIG_PATH"
fi

echo "Building common Kubernetes API objects"

common_config_output_dir=$ROOT_DIR/config

controller-gen paths=$ROOT_DIR/apis/... \
    crd:trivialVersions=true object:headerFile=$TEMPLATES_DIR/boilerplate.txt \
    output:crd:artifacts:config=$common_config_output_dir/crd/bases

echo "Building Kubernetes API objects for $SERVICE"
$ACK_GENERATE_BIN_PATH $apis_args
if [ $? -ne 0 ]; then
    exit 2
fi

config_output_dir="$ROOT_DIR/services/$SERVICE/config/"

pushd services/$SERVICE/apis/$ACK_GENERATE_API_VERSION 1>/dev/null

echo "Generating deepcopy code for $SERVICE"
controller-gen object:headerFile=$TEMPLATES_DIR/boilerplate.txt paths=./...

echo "Generating custom resource definitions for $SERVICE"
# Latest version of controller-gen (master) is required for following two reasons
# a) support for pointer values in map https://github.com/kubernetes-sigs/controller-tools/pull/317
# b) support for float type (allowDangerousTypes) https://github.com/kubernetes-sigs/controller-tools/pull/449
controller-gen crd:allowDangerousTypes=true paths=./... output:crd:artifacts:config=$config_output_dir/crd/bases

popd 1>/dev/null

echo "Building service controller for $SERVICE"
controller_args="controller $ag_args"
$ACK_GENERATE_BIN_PATH $controller_args
if [ $? -ne 0 ]; then
    exit 2
fi

pushd services/$SERVICE/pkg/resource 1>/dev/null

echo "Generating RBAC manifests for $SERVICE"
controller-gen rbac:roleName=$K8S_RBAC_ROLE_NAME paths=./... output:rbac:artifacts:config=$config_output_dir/rbac
# controller-gen rbac outputs a ClusterRole definition in a
# $config_output_dir/rbac/role.yaml file. We have some other standard Role
# files for a reader and writer role, so here we rename the `role.yaml` file to
# `cluster-role-controller.yaml` to better reflect what is in that file.
mv $config_output_dir/rbac/role.yaml $config_output_dir/rbac/cluster-role-controller.yaml

popd 1>/dev/null

echo "Running gofmt against generated code for $SERVICE"
gofmt -w "services/$SERVICE"
