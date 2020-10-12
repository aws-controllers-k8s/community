#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# ecr_repo_exists() returns 0 if an ECR Repository with the supplied name
# exists, 1 otherwise.
#
# Usage:
#
#   if ! ecr_repo_exists "$repo_name"; then
#       echo "Repo $repo_name does not exist!"
#   fi
ecr_repo_exists() {
    __repo_name="$1"
    daws ecr describe-repositories --repository-names "$__repo_name" --output json >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}

ecr_repo_jq() {
    __repo_name="$1"
    __jq_query="$2"
    json=$( daws ecr describe-repositories --repository-names "$__repo_name" --output json || exit 1 )
    echo "$json" | jq --raw-output $__jq_query
}
