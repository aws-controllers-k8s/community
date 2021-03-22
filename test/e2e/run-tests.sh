#!/usr/bin/env bash

# This script runs the existing bash tests for a service controller.

set -eo pipefail

E2E_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$E2E_DIR/../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

# set environment variables
SKIP_PYTHON_TESTS=${SKIP_PYTHON_TESTS:-"false"}
RUN_PYTEST_LOCALLY=${RUN_PYTEST_LOCALLY:="false"}
PYTEST_LOG_LEVEL="${PYTEST_LOG_LEVEL:-"INFO"}"

USAGE="
Usage:
  $(basename "$0") <service>

<service> should be an AWS service for which you wish to run tests -- e.g.
's3' 'sns' or 'sqs'

Environment variables:
  DEBUG:                    Set to any value to enable debug logging in the bash tests
  SKIP_PYTHON_TESTS         Whether to skip python tests and run bash tests instead for
                            the service controller (<true|false>)
                            Default: false
  RUN_PYTEST_LOCALLY        If python tests exist, whether to run them locally instead of
                            inside Docker (<true|false>)
                            Default: false
  PYTEST_LOG_LEVEL:         Set to any Python logging level for the Python tests.
                            Default: INFO
"

if [ $# -ne 1 ]; then
    echo "ERROR: $(basename "$0") only accepts a single parameter" 1>&2
    echo "$USAGE"
    exit 1
fi

# construct and validate service directory path
SERVICE="$1"
service_test_dir="$E2E_DIR/$SERVICE"
if [ ! -d "$service_test_dir" ]; then
    echo "No tests for service $SERVICE"
    exit 0
fi

# check if python tests exist for the service
[[ -f "$service_test_dir/__init__.py" ]] && python_tests_exist="true" || python_tests_exist="false"

# run tests
if [[ "$python_tests_exist" == "false" ]] || [[ "$SKIP_PYTHON_TESTS" == "true" ]]; then
  source "$SCRIPTS_DIR/lib/common.sh"

  echo "running bash tests..."
  service_test_files=$( find "$service_test_dir" -type f -name '*.sh' | sort )
  for service_test_file in $service_test_files; do
      test_name=$( filenoext "$service_test_file" )
      test_start_time=$( date +%s )
      bash $service_test_file
      test_end_time=$( date +%s )
      echo "$test_name took $( expr $test_end_time - $test_start_time ) second(s)"
  done

elif [[ "$RUN_PYTEST_LOCALLY" == "true" ]]; then
  echo "running python tests locally..."
  python bootstrap.py "${SERVICE}"
  set +e
  PYTHONPATH=. pytest -n auto --dist loadfile --log-cli-level "${PYTEST_LOG_LEVEL}" "${SERVICE}"
  python cleanup.py "${SERVICE}"
  set -eo pipefail
else
  echo "running python tests in Docker..."
  $E2E_DIR/build-run-test-dockerfile.sh $SERVICE
fi
