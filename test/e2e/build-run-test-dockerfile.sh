#!/usr/bin/env bash

set -eo pipefail

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

USAGE="
Usage:
  $(basename "$0") <service>

<service> should be an AWS service for which you wish to run tests -- e.g.
's3' 'sns' or 'sqs'
"

if [ $# -ne 1 ]; then
    echo "ERROR: $(basename "$0") only accepts a single parameter" 1>&2
    echo "$USAGE"
    exit 1
fi

SERVICE="$1"

KUBECONFIG_LOCATION="${KUBECONFIG:-"$HOME/.kube/config"}"

# Ensure we are inside the correct build context
pushd "${THIS_DIR}" 1> /dev/null
  # Build the dockerfile first
  TEST_DOCKER_SHA="$(docker build . --quiet)"
popd 1>/dev/null

# Ensure it can connect to KIND cluster on host device by running on host 
# network. 
# Pass AWS credentials and kubeconfig through to Dockerfile.
docker run --rm -it \
    --network="host" \
    -v $KUBECONFIG_LOCATION:/root/.kube/config:z \
    -v $HOME/.aws/credentials:/root/.aws/credentials:z \
    -v $THIS_DIR:/root/tests:z \
    -e AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-"us-west-2"}" \
    -e AWS_ACCESS_KEY_ID \
    -e AWS_SECRET_ACCESS_KEY \
    -e AWS_SESSION_TOKEN \
    -e RUN_PYTEST_LOCALLY="true" \
    $TEST_DOCKER_SHA "${SERVICE}"
