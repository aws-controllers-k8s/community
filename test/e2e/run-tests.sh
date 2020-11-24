#!/usr/bin/env bash

set -eo pipefail

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"
BIN_DIR="$ROOT_DIR/bin"

source "$SCRIPTS_DIR/lib/common.sh"

USAGE="
Usage:
  $(basename "$0") <service>

<service> should be an AWS service for which you wish to run tests -- e.g.
's3' 'sns' or 'sqs'

Environment variables:
  DEBUG:        Set to any value to enable debug logging in the tests
"

if [ $# -ne 1 ]; then
    echo "ERROR: $(basename "$0") only accepts a single parameter" 1>&2
    echo "$USAGE"
    exit 1
fi

SERVICE="$1"

service_test_dir="$THIS_DIR/$SERVICE"

if [ ! -d "$service_test_dir" ]; then
    echo "No tests for service $SERVICE"
    exit 0
fi

# find all files except under helper directory
service_test_files=$( find "$service_test_dir" -name helper -prune -false -o -type f ! -name '.*' | sort )

for service_test_file in $service_test_files; do
    test_name=$( filenoext "$service_test_file" )
    test_start_time=$( date +%s )
    bash $service_test_file
    test_end_time=$( date +%s )
    echo "$test_name took $( expr $test_end_time - $test_start_time ) second(s)"
done
