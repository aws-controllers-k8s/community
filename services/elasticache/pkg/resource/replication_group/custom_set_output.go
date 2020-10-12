// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package replication_group

import (
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/service/elasticache"
	corev1 "k8s.io/api/core/v1"
)

func (rm *resourceManager) CustomDescribeReplicationGroupsSetOutput(
	r *resource,
	resp *elasticache.DescribeReplicationGroupsOutput,
	ko *svcapitypes.ReplicationGroup,
) *svcapitypes.ReplicationGroup {
	if len(resp.ReplicationGroups) == 0 {
		return ko
	}
	elem := resp.ReplicationGroups[0]
	rm.customSetOutput(r, elem, ko)
	return ko
}

func (rm *resourceManager) CustomCreateReplicationGroupSetOutput(
	r *resource,
	resp *elasticache.CreateReplicationGroupOutput,
	ko *svcapitypes.ReplicationGroup,
) *svcapitypes.ReplicationGroup {
	rm.customSetOutput(r, resp.ReplicationGroup, ko)
	return ko
}

func (rm *resourceManager) CustomModifyReplicationGroupSetOutput(
	r *resource,
	resp *elasticache.ModifyReplicationGroupOutput,
	ko *svcapitypes.ReplicationGroup,
) *svcapitypes.ReplicationGroup {
	rm.customSetOutput(r, resp.ReplicationGroup, ko)
	return ko
}

func (rm *resourceManager) customSetOutput(
	r *resource,
	respRG *elasticache.ReplicationGroup,
	ko *svcapitypes.ReplicationGroup,
) {
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
	rgStatus := respRG.Status
	syncConditionStatus := corev1.ConditionUnknown
	if rgStatus != nil {
		if *rgStatus == "available" ||
			*rgStatus == "create-failed" {
			syncConditionStatus = corev1.ConditionTrue
		} else {
			// resource in "creating", "modifying" , "deleting", "snapshotting"
			// states is being modified at server end
			// thus current status is considered out of sync.
			syncConditionStatus = corev1.ConditionFalse
		}
	}
	var resourceSyncedCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			resourceSyncedCondition = condition
			break
		}
	}
	if resourceSyncedCondition == nil {
		resourceSyncedCondition = &ackv1alpha1.Condition{
			Type:   ackv1alpha1.ConditionTypeResourceSynced,
			Status: syncConditionStatus,
		}
		ko.Status.Conditions = append(ko.Status.Conditions, resourceSyncedCondition)
	} else {
		resourceSyncedCondition.Status = syncConditionStatus
	}

}
