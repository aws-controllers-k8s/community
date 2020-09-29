#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$THIS_DIR/../../.."
SCRIPTS_DIR="$ROOT_DIR/scripts"

source "$SCRIPTS_DIR/lib/common.sh"
source "$SCRIPTS_DIR/lib/aws.sh"
source "$SCRIPTS_DIR/lib/aws/ecr.sh"
source "$SCRIPTS_DIR/lib/k8s.sh"
source "$SCRIPTS_DIR/lib/testutil.sh"

wait_seconds=5
test_name="$( filenoext "${BASH_SOURCE[0]}" )"
service_name="ecr"
ack_ctrl_pod_id=$( controller_pod_id )
debug_msg "executing test: $service_name/$test_name"

# This smoke test creates and deletes a set of ECR repositories. It creates
# more than 1 repository in order to ensure that the ReadMany code paths and
# the associated generated code that do object lookups for a single object work
# when >1 object are returned in various List operations.

# PRE-CHECKS

for x in a b c; do

    repo_name="ack-test-smoke-$service_name-$x"
    resource_name="repositories/$repo_name"

    if ecr_repo_exists "$repo_name"; then
        echo "FAIL: expected $repo_name to not exist in ECR. Did previous test run cleanup?"
        exit 1
    fi

    if k8s_resource_exists "$resource_name"; then
        echo "FAIL: expected $resource_name to not exist. Did previous test run cleanup?"
        exit 1
    fi

done

# TEST ACTIONS and ASSERTIONS

for x in a b c; do

    repo_name="ack-test-smoke-$service_name-$x"

    cat <<EOF | kubectl apply -f -
apiVersion: ecr.services.k8s.aws/v1alpha1
kind: Repository
metadata:
  name: $repo_name
spec:
  repositoryName: $repo_name
EOF

done

sleep $wait_seconds

for x in a b c; do

    repo_name="ack-test-smoke-$service_name-$x"
    resource_name="repositories/$repo_name"

    debug_msg "checking repository $repo_name created in ECR"
    if ! ecr_repo_exists "$repo_name"; then
        echo "FAIL: expected $repo_name to have been created in ECR"
        kubectl logs -n ack-system "$ack_ctrl_pod_id"
        exit 1
    fi

    kubectl delete "$resource_name" 2>/dev/null
    assert_equal "0" "$?" "Expected success from kubectl delete but got $?" || exit 1

done

sleep $wait_seconds

for x in a b c; do

    repo_name="ack-test-smoke-$service_name-$x"

    if ecr_repo_exists "$repo_name"; then
        echo "FAIL: expected $repo_name to be deleted in ECR"
        kubectl logs -n ack-system "$ack_ctrl_pod_id"
        exit 1
    fi
done

assert_pod_not_restarted $ack_ctrl_pod_id
