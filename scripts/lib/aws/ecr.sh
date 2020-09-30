#!/usr/bin/env bash

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
