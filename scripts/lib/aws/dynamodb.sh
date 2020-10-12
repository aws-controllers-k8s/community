#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# dynamodb_table_exists() returns 0 if a DynamoDB table with the supplied name
# exists, 1 otherwise.
#
# Usage:
#
#   if ! dynamodb_table_exists "$table_name"; then
#       echo "Table $table_name does not exist!"
#   fi
dynamodb_table_exists() {
    __repo_name="$1"
    daws dynamodb describe-table --table-name "$table_name" --output json >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}
