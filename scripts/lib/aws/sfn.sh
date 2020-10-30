#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

. $SCRIPTS_DIR/lib/common.sh
. $SCRIPTS_DIR/lib/aws.sh

# sfn_statemachine_exists() returns 0 if a Step Functions state machine with the supplied ARN
# exists, 1 otherwise.
#
# Usage:
#
#   if ! sfn_statemachine_exists "$statemachine_arn"; then
#       echo "State machine $statemachine_arn does not exist!"
#   fi
sfn_statemachine_exists() {
    __statemachine_arn="$1"
    daws stepfunctions describe-state-machine --region $AWS_REGION --state-machine-arn "$__statemachine_arn" >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}

# sfn_activity_exists() returns 0 if a Step Functions activity with the supplied ARN
# exists, 1 otherwise.
#
# Usage:
#
#   if ! sfn_activity_exists "$activity_arn"; then
#       echo "State machine $activity_arn does not exist!"
#   fi
sfn_activity_exists() {
    __activity_arn="$1"
    daws stepfunctions describe-activity --region $AWS_REGION --activity-arn "$__activity_arn" >/dev/null 2>&1
    if [[ $? -eq 254 ]]; then
        return 1
    else
        return 0
    fi
}


# State machines require an execution role; they assume that role when making calls to other services.
# sfn_setup_iam_resources_for_executionrole creates an iam-role and attaches the AWSDenyAll policy to it.
# sfn_setup_iam_resources_for_executionrole accepts only one required parameter: role_name
sfn_setup_iam_resources_for_executionrole() {
  if [[ $# -ne 1 ]]; then
    echo "FATAL: Wrong number of arguments passed to sfn_setup_iam_resources_for_executionrole"
    echo "Usage: sfn_setup_iam_resources_for_executionrole role_name"
    exit 1
  fi

  local __role_name="$1"
  daws iam create-role --role-name "$__role_name" --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "states.amazonaws.com"}, "Action": "sts:AssumeRole"}]}' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam create-role --role-name $__role_name but got $?" || exit 1
  daws iam attach-role-policy --role-name "$__role_name" --policy-arn 'arn:aws:iam::aws:policy/AWSDenyAll' >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam attach-role-policy --role-name $__role_name but got $?" || exit 1
}

## sfn_clean_up_iam_resources_for_executionrole deletes the execution role created for state machines.
## sfn_clean_up_iam_resources_for_executionrole accepts only one required parameter: role_name
sfn_clean_up_iam_resources_for_executionrole() {
  local __role_name="$1"
  daws iam detach-role-policy --role-name "$__role_name" --policy-arn arn:aws:iam::aws:policy/AWSDenyAll >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam detach-role-policy --role-name $__role_name but got $?" || exit 1

  daws iam delete-role --role-name "$__role_name" >/dev/null
  assert_equal "0" "$?" "Expected success from aws iam delete-role --role-name $__role_name but got $?" || exit 1
}
