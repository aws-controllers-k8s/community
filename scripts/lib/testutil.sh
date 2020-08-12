#!/usr/bin/env bash

DEFAULT_DEBUG_PREFIX="DEBUG: "

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

# debug_msg prints out a supplied message if the DEBUG environs variable is
# set. An optional second argument indicates the "indentation level" for the
# message. If the indentation level argument is missing, we look for the
# existence of an environs variable called "indent_level" and use that
debug_msg() {
    local __msg="$1"
    local __indent_level="$2"
    local __debug="${DEBUG:-""}"
    local __debug_prefix="${DEBUG_PREFIX:-$DEFAULT_DEBUG_PREFIX}"
    if [ ! -n "$__debug" ]; then
        return 0
    fi
    __indent=""
    if [ -n "$__indent_level" ]; then
        __indent="$( for each in $( seq 0 $__indent_level ); do printf " "; done )"
    fi
    echo "$__debug_prefix$__indent$__msg"
}

# controller_pod_id returns the ID of the pod running the ACK service
# controller for the supplied service
#
# Usage:
#
#   echo controller_pod_id "ecr"
controller_pod_id() {
    local __msg="$1"
    if [ ! -n "$__msg" ]; then
        echo "ERROR: controller_pod_id requires a single argument, the name of the service"
        exit 127
    fi
    kubectl get pods -n ack-system --field-selector="status.phase=Running" \
        --sort-by=.metadata.creationTimestamp \
        --output jsonpath='{.items[-1].metadata.name}'
}
