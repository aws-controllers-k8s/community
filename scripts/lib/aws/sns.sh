#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# sns_topic_exists() returns 0 if an SNS topic with the supplied ARN
# exists, 1 otherwise.
#
# sns_topic_exists TOPIC_NAME [ AWS_REGION ] [ AWS_PROFILE ]
#
# Arguments:
#
#   TOPIC_NAME      required string for the name of the topic to check
#   AWS_REGION      alternate region to use
#   AWS_PROFILE     alternate profile to use
#
# Usage:
#
#   if ! sns_topic_exists "$topic_arn"; then
#       echo "Topic $topic_arn does not exist!"
#   fi
sns_topic_exists() {
    __topic_arn="$1"
    __region_args=""
    __region="$2"
    if [[ -n "$__region" ]]; then
        __region_args=" --region $__region"
    fi
    __profile_args=""
    __profile="$3"
    if [[ -n "$__profile" ]]; then
        __profile_args=" --profile $__profile"
    fi

    daws sns get-topic-attributes --topic-arn "$__topic_arn" $__region_args $__profile_args --output json >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}

