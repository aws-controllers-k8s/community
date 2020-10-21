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
