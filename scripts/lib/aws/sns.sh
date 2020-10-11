#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# sns_topic_exists() returns 0 if an SNS topic with the supplied ARN
# exists, 1 otherwise.
#
# Usage:
#
#   if ! sns_topic_exists "$topic_arn"; then
#       echo "Topic $topic_arn does not exist!"
#   fi
sns_topic_exists() {
    __topic_arn="$1"
    daws sns get-topic-attributes --topic-arn "$topic_arn" --output json >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}

