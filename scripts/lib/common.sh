#!/usr/bin/env bash

# setting the -x option if debugging is true
if [[ "${DEBUG:-"false"}" = "true" ]]; then
    set -x
fi

# check_is_installed checks to see if the supplied executable is installed and
# exits if not. An optional second argument is an extra message to display when
# the supplied executable is not installed.
#
# Usage:
#
#   check_is_installed PROGRAM [ MSG ]
#
# Example:
#
#   check_is_installed kind "You can install kind with the helper scripts/install-kind.sh"
check_is_installed() {
    local __name="$1"
    local __extra_msg="$2"
    if ! is_installed "$__name"; then
        echo "FATAL: Missing requirement '$__name'"
        echo "Please install $__name before running this script."
        if [[ -n $__extra_msg ]]; then
            echo ""
            echo "$__extra_msg"
            echo ""
        fi
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

DEFAULT_DEBUG_PREFIX="DEBUG: "

# debug_msg prints out a supplied message if the DEBUG environs variable is
# set. An optional second argument indicates the "indentation level" for the
# message. If the indentation level argument is missing, we look for the
# existence of an environs variable called "indent_level" and use that.
debug_msg() {
    local __msg=${1:-}
    local __indent_level=${2:-}
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
