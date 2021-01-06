#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# kms_key_exists() returns 0 if an kms key with the supplied id
# exists, 1 otherwise.
#
# kms_key_exists KEY_ID [ AWS_REGION ] [ AWS_PROFILE ]
#
# Arguments:
#
#   KEY_ID          required string for the kms key id to check
#   AWS_REGION      alternate region to use
#   AWS_PROFILE     alternate profile to use
#
# Usage:
#
#   if ! kms_key_exists "$KEY_ID"; then
#       echo "Key $KEY_ID does not exist!"
#   fi

kms_key_exists() {
    __key_id="$1"
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

    daws kms describe-key --key-id $__key_id $__region_args $__profile_args --output json
    if [[ $? -ne 0 ]]; then
        return 1
    else
        return 0
    fi
}
