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

package snapshot

import (
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/service/elasticache"
	corev1 "k8s.io/api/core/v1"
)

func (rm *resourceManager) CustomDescribeSnapshotSetOutput(
	r *resource,
	resp *elasticache.DescribeSnapshotsOutput,
	ko *svcapitypes.Snapshot,
) *svcapitypes.Snapshot {
	if len(resp.Snapshots) == 0 {
		return ko
	}
	elem := resp.Snapshots[0]
	rm.customSetOutput(r, elem, ko)
	return ko
}

func (rm *resourceManager) CustomCreateSnapshotSetOutput(
	r *resource,
	resp *elasticache.CreateSnapshotOutput,
	ko *svcapitypes.Snapshot,
) *svcapitypes.Snapshot {
	rm.customSetOutput(r, resp.Snapshot, ko)
	return ko
}

func (rm *resourceManager) CustomCopySnapshotSetOutput(
	r *resource,
	resp *elasticache.CopySnapshotOutput,
	ko *svcapitypes.Snapshot,
) *svcapitypes.Snapshot {
	rm.customSetOutput(r, resp.Snapshot, ko)
	return ko
}

func (rm *resourceManager) customSetOutput(
	r *resource,
	respSnapshot *elasticache.Snapshot,
	ko *svcapitypes.Snapshot,
) {
	if respSnapshot.ReplicationGroupId != nil {
		ko.Spec.ReplicationGroupID = respSnapshot.ReplicationGroupId
	}

	if respSnapshot.KmsKeyId != nil {
		ko.Spec.KMSKeyID = respSnapshot.KmsKeyId
	}

	if respSnapshot.CacheClusterId != nil {
		ko.Spec.CacheClusterID = respSnapshot.CacheClusterId
	}

	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
	snapshotStatus := respSnapshot.SnapshotStatus
	syncConditionStatus := corev1.ConditionUnknown
	if snapshotStatus != nil {
		if *snapshotStatus == "available" ||
			*snapshotStatus == "failed" {
			syncConditionStatus = corev1.ConditionTrue
		} else {
			// resource in "creating", "restoring","exporting"
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
