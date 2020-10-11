#!/usr/bin/env bash

CONTROLLER_TOOLS_VERSION="v0.4.0"

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
# get_field_from_status accepts three parameters. namespace (which is an optional parameter),
# resource_name and status_field
get_field_from_status() {

  if [[ "$#" -lt 2 || "$#" -gt 3 ]]; then
    echo "[FAIL] Usage: get_field_from_status [namespace] resource_name status_field"
    exit 1
  fi

  local __namespace=""
  local __resource_name=""
  local __status_field=""

  if [[ "$#" -eq 2 ]]; then
    __resource_name="$1"
    __status_field="$2"
  else
    __namespace="$1"
    __resource_name="$2"
    __status_field="$3"
  fi

  local __args=""
  if [ -n "$__namespace" ]; then
      __args="$__args-n $__namespace"
  fi


  local __id=$(kubectl get $__args "$__resource_name" -o=json | jq -r .status."$__status_field")
  if [[ -z "$__id" ]];then
    echo "FAIL: $__resource_name resource's status does not have $__status_field field"
    exit 1
  fi
  echo "$__id"
}
