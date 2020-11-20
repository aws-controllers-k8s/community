#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# lambda_function_exist() returns 0 if a lambda Function with the supplied name
# exists, 1 otherwise.
#
# Usage:
#
#   if ! lambda_function_exist "$repo_name"; then
#       echo "Repo $repo_name does not exist!"
#   fi
lambda_function_exists() {
    __function_name="$1"
    daws lambda get-function --function-names "$__function_name" --output json >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}

lambda_function_jq() {
    __lambda_name="$1"
    __jq_query="$2"
    json=$( daws ecr get-lambda --function-name "$__lambda_name" --output json || exit 1 )
    echo "$json" | jq --raw-output $__jq_query
}
