#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# s3_bucket_exists() returns 0 if an S3 Bucket with the supplied name
# exists, 1 otherwise.
#
# s3_bucket_exists BUCKET_NAME [ AWS_REGION ] [ AWS_PROFILE ]
#
# Arguments:
#
#   BUCKET_NAME     required string for the name of the bucket to check
#   AWS_REGION      alternate region to use
#   AWS_PROFILE     alternate profile to use
#
# Usage:
#
#   if ! s3_bucket_exists "$bucket_name"; then
#       echo "Bucket $bucket_name does not exist!"
#   fi
s3_bucket_exists() {
    __bucket_name="$1"
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

    jq_expr='.Buckets[] | select(.Name | contains($BUCKET_NAME))'
    daws s3api list-buckets $__region_args $__profile_args --output json | jq -e --arg BUCKET_NAME "$bucket_name" "$jq_expr"
    if [[ $? -eq 4 ]]; then
        return 1
    else
        return 0
    fi
}
