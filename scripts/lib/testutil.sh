#!/usr/bin/env bash

# assert_equal returns 0 if the first two supplied arguments are equal, 1
# otherwise after prining a failure message (optional third argument)
#
# Usage:
#
#   assert_equal "a" "b" "Expected a but got b!" || exit 1
assert_equal() {
    local __expected="$1"
    local __actual="$2"
    local __msg="$3"
    if [ ! -n "$__msg" ]; then
        __msg="Expected '$__expected' to equal '$__actual'"
    fi
    if [ "$__expected" != "$__actual" ]; then
        echo "FAIL: $__msg"
        return 1
    fi
    return 0
}

# controller_pod_id returns the ID of the pod running the ACK service
# controller for the supplied service
#
# Usage:
#
#   echo controller_pod_id
controller_pod_id() {
    local x=0
    while true; do
        pod_id=$( kubectl get pods -n ack-system --field-selector="status.phase=Running" \
            --sort-by=.metadata.creationTimestamp \
            --output jsonpath='{.items[-1].metadata.name}' 2>/dev/null )
        if [[ $? -eq 0 ]]; then
            break
        else
            if [[ $x -gt 2 ]]; then
                echo "FAIL: Could not get ACK service controller Pod ID"
                exit 1
            else
                x=$(( x + 1 ))
                sleep 2
            fi
        fi
    done
    echo "$pod_id"
}

# assert_pod_not_restarted ensures the supplied Pod has not been restarted
# (being restarted indicates there was a panic/segfault in the controller code)
#
# Usage:
#
#   assert_pod_not_restarted controller_pod_id
assert_pod_not_restarted() {
    local __pod_id="$1"
    if [ ! -n "$__pod_id" ]; then
        echo "ERROR: assert_pod_not_restarted requires a single argument, the ID of the Pod to check"
        exit 127
    fi
    local __ns=${2:-}
    if [ ! -n "$__ns" ]; then
        __ns="ack-system"
    fi
    restartCount=$( kubectl get pods -n "$__ns" "$__pod_id" --output jsonpath='{.status.containerStatuses[0].restartCount}' )
    if [ "$restartCount" != "0" ]; then
        echo "FAIL: Expected pod $__pod_id to not have been restarted but it has been restarted $restartCount times."
        echo "****************************** logs from previous controller pod ************************************"
        kubectl logs -n ack-system --previous "$__pod_id"
        return 1
    fi
}

# k8s_wait_resource_synced checks the given resource for an ACK.ResourceSynced condition in its
#   k8s status.conditions property. Times out if condition has not been met for a long time. This function
#   is intended to be used after yaml application to await creation of a resource.
# k8s_wait_resource_synced requires 3 arguments:
#   k8s_resource_name: the name of the resource, e.g. "snapshots/test-snapshot"
#   wait_periods: the number of 60-second periods to wait for the resource before timing out
#   service_name: the aws service name. It is used to determine the controller deployment, in case credentials need
#                 to be rotated during the wait.
k8s_wait_resource_synced() {
  if [[ $# -ne 3 ]]; then
    echo "FATAL: Wrong number of arguments passed to ${FUNCNAME[0]}"
    echo "Usage: ${FUNCNAME[0]} k8s_resource_name wait_periods service_name"
    exit 1
  fi

  local k8s_resource_name="$1"
  local wait_periods="$2"
  local service_name="$3"

  kubectl get "$k8s_resource_name" 1>/dev/null 2>&1
  assert_equal "0" "$?" "Resource $k8s_resource_name doesn't exist in k8s cluster" || exit 1

  local wait_failed="true"
  for i in $(seq 1 "$wait_periods"); do
    # Test role credentials expire after 15 minutes (Refer: aws.sh::aws_generate_temp_creds)
    # Ensure that credentials are reloaded after first iteration and after every 15 minutes.
    # else the controller fails to get the latest details from aws service api
    # and the test fails on sync status.
    if [[ "$((i % 15))" == "0" || "$i" == "2" ]]; then
      k8s_controller_reload_credentials "$service_name"
    fi
    debug_msg "waiting for resource $k8s_resource_name to be synced ($i)"
    sleep 60

    # ensure we at least have .status.conditions
    local conditions=$(kubectl get "$k8s_resource_name" -o json | jq -r -e ".status.conditions[]")
    assert_equal "0" "$?" "Expected .status.conditions property to exist for $k8s_resource_name" || exit 1

    # this condition should probably always exist, regardless of the value
    local synced_cond=$(echo $conditions | jq -r -e 'select(.type == "ACK.ResourceSynced")')
    assert_equal "0" "$?" "Expected ACK.ResourceSynced condition to exist for $k8s_resource_name" || exit 1

    # check value of condition; continue if not yet set True
    local cond_status=$(echo $synced_cond | jq -r -e ".status")
    if [[ "$cond_status" == "True" ]]; then
      wait_failed="false"
      debug_msg "resource $k8s_resource_name is synced, continuing.."
      break
    fi
  done

  assert_equal "false" "$wait_failed" "Wait for resource $k8s_resource_name to be synced timed out" || exit 1
}

# k8s_check_resource_terminal_condition_true asserts that the terminal condition of the given resource
#   exists, has status "True", and that the message associated with the terminal condition matches the
#   one provided.
# k8s_check_resource_terminal_condition_true requires 2 arguments:
#   k8s_resource_name: the name of the resource, e.g. "snapshots/test-snapshot"
#   expected_substring: a substring of the expected message associated with the terminal condition
k8s_check_resource_terminal_condition_true() {
  if [[ $# -ne 2 ]]; then
    echo "FATAL: Wrong number of arguments passed to ${FUNCNAME[0]}"
    echo "Usage: ${FUNCNAME[0]} replication_group_id expected_substring"
    exit 1
  fi
  local k8s_resource_name="$1"
  local expected_substring="$2"

  local resource_json=$(kubectl get "$k8s_resource_name" -o json)
  assert_equal "0" "$?" "Expected $k8s_resource_name to exist in k8s cluster" || exit 1

  local terminal_cond=$(echo $resource_json | jq -r -e ".status.conditions[]" | jq -r -e 'select(.type == "ACK.Terminal")')
  assert_equal "0" "$?" "Expected resource $k8s_resource_name to have a terminal condition" || exit 1

  local status=$(echo $terminal_cond | jq -r ".status")
  assert_equal "True" "$status" "expected status of terminal condition to be True for resource $k8s_resource_name" || exit 1

  local cond_msg=$(echo $terminal_cond | jq -r ".message")
  if [[ $cond_msg != *"$expected_substring"* ]]; then
    echo "FAIL: resource $k8s_resource_name has terminal condition set True, but with message different than expected"
    exit 1
  fi
}