#!/usr/bin/env bash

# A script that builds the controllers for one or more AWS services

set -E

DIR=$(cd "$(dirname "$0")"; pwd)
source "$DIR"/lib/common.sh

ACK_GENERATE_CACHE_DIR=${ACK_GENERATE_CACHE_DIR:-"~/.cache/aws-controllers-k8s"}
export ACK_GENERATE_CACHE_DIR

USAGE="
Usage:
  $(basename "$0") <services>

<services> should be a space-delimited list of AWS service API
aliases that you wish to build -- e.g. 's3 sns sqs'

Environment variables:
  ACK_GENERATE_CACHE_DIR:         Overrides the directory used for caching AWS API model.
                                  Default: ~/.cache/aws-controllers-k8s
"

shift $(( OPTIND - 1))

SERVICES=()

while [ -n "$1" ]; do
    SERVICES+=( $1 )
    shift
done

if [[ ${#SERVICES[@]} -eq 0 ]]; then
    echo "ERROR: Specify at least one service to build a controller for" 1>&2
    echo "$USAGE"
    exit 1
fi

for SERVICE in ${SERVICES[@]}; do
    SERVICE=$(echo "$SERVICE" | tr '[:upper:]' '[:lower:]')
    $DIR/build-controller.sh $SERVICE
    if [ $? -ne 0 ]; then
        exit 2
    fi
done
