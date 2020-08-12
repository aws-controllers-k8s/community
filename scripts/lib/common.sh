#!/usr/bin/env bash

check_is_installed() {
    local __name="$1"
    if ! is_installed "$__name"; then
        echo "Please install $__name before running this script."
        exit 1
    fi
}

is_installed() {
    local __name="$1"
    if $(which $__name >/dev/null 2>&1); then
        return 0
    else
        return 1
    fi
}

display_timelines() {
    echo ""
    echo "Displaying all step durations."
    echo "TIMELINE: Docker build took $DOCKER_BUILD_DURATION seconds."
    echo "TIMELINE: Upping test cluster took $UP_CLUSTER_DURATION seconds."
    echo "TIMELINE: Base image integration tests took $BASE_INTEGRATION_DURATION seconds."
    echo "TIMELINE: Current image integration tests took $LATEST_INTEGRATION_DURATION seconds."
    echo "TIMELINE: Down processes took $DOWN_DURATION seconds."
}

should_execute() {
  if [[ "$TEST_PASS" -ne 0 ]]; then
    echo "NOTE: Skipping operation '$1'. Test is already marked as failed."
    return 1
  else
    return 0
  fi
}

# filenoext returns just the name of the supplied filename without the
# extension
filenoext() {
    local __name="$1"
    local __filename=$( basename "$__name" )
    # How much do I despise Bash?!
    echo "${__filename%.*}"
}
