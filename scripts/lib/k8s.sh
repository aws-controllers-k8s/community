#!/usr/bin/env bash

CONTROLLER_TOOLS_VERSION="v0.4.0"

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/aws.sh

# controller_gen_version_equals accepts a string version and returns 0 if the
# installed version of controller-gen matches the supplied version, otherwise
# returns 1
#
# Usage:
#
#   if controller_gen_version_equals "v0.4.0"; then
#       echo "controller-gen is at version 0.4.0"
#   fi
k8s_controller_gen_version_equals() {
    currentver="$(controller-gen --version | cut -d' ' -f2 | tr -d '\n')";
    requiredver="$1";
    if [ "$currentver" = "$requiredver" ]; then
        return 0
    else
        return 1
    fi;
}

# resource_exists returns 0 when the supplied resource can be found, 1
# otherwise. An optional second parameter overrides the Kubernetes namespace
# argument
k8s_resource_exists() {
    local __res_name=${1:-}
    local __namespace=${2:-}
    local __args=""
    if [ -n "$__namespace" ]; then
        __args="$__args-n $__namespace"
    fi
    kubectl get $__args "$__res_name" >/dev/null 2>&1
}

# get_field_from_status returns the field from status of a K8s resource
# get_field_from_status accepts the following parameters:
#   $1: resource_name, format: CR_Kind/Instance
#   $2: status_field
#   $3 (optional): Namespace
#   $4 (optional): Timeout to wait for the field to populate
# To pass only timout and skip namespace use:
# get_field_from_status resource_name status_field "" [timeout]
get_field_from_status() {

  if [[ "$#" -lt 2 || "$#" -gt 5 ]]; then
    echo "[FAIL] Usage: get_field_from_status resource_name status_field [namespace] [timeout]"
    exit 1
  fi

  local __resource_name="$1"
  local __status_field="$2"
  local __namespace="${3:-default}"
  local __timeout="${4:-20}"
  local __retry_interval=5

  local __args=""
  if [ -n "$__namespace" ]; then
      __args="$__args-n $__namespace"
  fi

  local __id=""
  while [ "$__timeout" -gt 0 ]; do
    local __id=$(kubectl get $__args "$__resource_name" -o=json | jq -r .status."$__status_field")
    if [[ ( -n "$__id" && "$__id" != "null" ) ]];then
      break
    fi
    sleep "$__retry_interval"
    __timeout=$(($__timeout-$__retry_interval))
  done

  if [[ ( -z "$__id" || "$__id" == "null") ]];then
    echo "FAIL: $__resource_name resource's status does not have $__status_field field"
    exit 1
  else
    echo "$__id"
  fi
}

# k8s_controller_reload_credentials generates AWS temporary credentials
# and adds them to service controller running on kubernetes cluster.
# it requires 1 argument: service name
# it depends upon:
#   $AWS_ACCESS_KEY_ID
#   $AWS_SECRET_ACCESS_KEY
#   $AWS_SESSION_TOKEN
k8s_controller_reload_credentials() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to k8s_controller_reload_credentials"
    echo "Usage: k8s_controller_reload_credentials service_name"
    exit 1
  fi
  local service_name="$1"
  echo -n "generating AWS temporary credentials and adding to env vars map ... "
  aws_generate_temp_creds
  kubectl -n ack-system set env deployment/ack-"$service_name"-controller \
    AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" 1>/dev/null
  sleep 15
  echo "ok."
}

# k8_get_pod_status returns status of the supplied pod name
# it requires 1 argument: pod name
k8_get_pod_status() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to ${FUNCNAME[0]}"
    echo "Usage: ${FUNCNAME[0]} pod_name"
    exit 1
  fi
  local pod_name="$1"
  echo $(kubectl get pods -n ack-system -o json | jq -r -e ".items[] | select(.metadata.name | contains(\"$pod_name\")) | .status.phase")
}

# k8_wait_for_pod_status waits for pod status for the given timeout interval (seconds)
# it requires 3 arguments:
# - pod name
# - expected status
# - timeout interval (in seconds)
k8_wait_for_pod_status() {
  if [[ $# -ne 3 ]]; then
    echo "FATAL: Wrong number of arguments passed to ${FUNCNAME[0]}"
    echo "Usage: ${FUNCNAME[0]} pod_name expected_status timeout_interval"
    exit 1
  fi
  local pod_name="$1"
  local expected_status="$2"
  local timeout_interval="$3"

  local x=0
  while true; do
    sleep 1
    x=$(( x + 1 ))
    local actual_status=$(k8_get_pod_status "$pod_name")
    if [ "$expected_status" == "$actual_status" ]; then
      return 0
    fi
    if [[ $x -gt $timeout_interval ]]; then
      break
    fi
  done
  return 1
}