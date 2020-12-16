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
	"context"
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/service/elasticache"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The number of minutes worth of events to retrieve.
	// 14 days in minutes
	eventsDuration = 20160
)

func (rm *resourceManager) CustomDescribeReplicationGroupsSetOutput(
	ctx context.Context,
	r *resource,
	resp *elasticache.DescribeReplicationGroupsOutput,
	ko *svcapitypes.ReplicationGroup,
) (*svcapitypes.ReplicationGroup, error) {
	if len(resp.ReplicationGroups) == 0 {
		return ko, nil
	}
	elem := resp.ReplicationGroups[0]
	rm.customSetOutput(r, elem, ko)
	err := rm.customSetOutputSupplementAPIs(ctx, r, elem, ko)
	if err != nil {
		return nil, err
	}
	return ko, nil
}

func (rm *resourceManager) CustomCreateReplicationGroupSetOutput(
	ctx context.Context,
	r *resource,
	resp *elasticache.CreateReplicationGroupOutput,
	ko *svcapitypes.ReplicationGroup,
) (*svcapitypes.ReplicationGroup, error) {
	rm.customSetOutput(r, resp.ReplicationGroup, ko)
	return ko, nil
}

func (rm *resourceManager) CustomModifyReplicationGroupSetOutput(
	ctx context.Context,
	r *resource,
	resp *elasticache.ModifyReplicationGroupOutput,
	ko *svcapitypes.ReplicationGroup,
) (*svcapitypes.ReplicationGroup, error) {
	rm.customSetOutput(r, resp.ReplicationGroup, ko)
	return ko, nil
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

	if rgStatus != nil && (*rgStatus == "available" || *rgStatus == "snapshotting") {
		input, err := rm.newListAllowedNodeTypeModificationsPayLoad(respRG)

		if err == nil {
			resp, apiErr := rm.sdkapi.ListAllowedNodeTypeModifications(input)
			rm.metrics.RecordAPICall("READ_MANY", "ListAllowedNodeTypeModifications", apiErr)
			// Overwrite the values for ScaleUp and ScaleDown
			if apiErr == nil {
				ko.Status.AllowedScaleDownModifications = resp.ScaleDownModifications
				ko.Status.AllowedScaleUpModifications = resp.ScaleUpModifications
			}
		}
	} else {
		ko.Status.AllowedScaleDownModifications = nil
		ko.Status.AllowedScaleUpModifications = nil
	}
}

// newListAllowedNodeTypeModificationsPayLoad returns an SDK-specific struct for the HTTP request
// payload of the ListAllowedNodeTypeModifications API call.
func (rm *resourceManager) newListAllowedNodeTypeModificationsPayLoad(respRG *elasticache.ReplicationGroup) (
	*svcsdk.ListAllowedNodeTypeModificationsInput, error) {
	res := &svcsdk.ListAllowedNodeTypeModificationsInput{}

	if respRG.ReplicationGroupId != nil {
		res.SetReplicationGroupId(*respRG.ReplicationGroupId)
	}

	return res, nil
}

func (rm *resourceManager) customSetOutputSupplementAPIs(
	ctx context.Context,
	r *resource,
	respRG *elasticache.ReplicationGroup,
	ko *svcapitypes.ReplicationGroup,
) error {
	events, err := rm.provideEvents(ctx, r.ko.Spec.ReplicationGroupID, 20)
	if err != nil {
		return err
	}
	ko.Status.Events = events
	return nil
}

func (rm *resourceManager) provideEvents(
	ctx context.Context,
	replicationGroupId *string,
	maxRecords int64,
) ([]*svcapitypes.Event, error) {
	input := &elasticache.DescribeEventsInput{}
	input.SetSourceType("replication-group")
	input.SetSourceIdentifier(*replicationGroupId)
	input.SetMaxRecords(maxRecords)
	input.SetDuration(eventsDuration)
	resp, err := rm.sdkapi.DescribeEventsWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "DescribeEvents-ReplicationGroup", err)
	if err != nil {
		return nil, err
	}
	events := []*svcapitypes.Event{}
	if resp.Events != nil {
		for _, respEvent := range resp.Events {
			event := &svcapitypes.Event{}
			if respEvent.Message != nil {
				event.Message = respEvent.Message
			}
			if respEvent.Date != nil {
				eventDate := metav1.NewTime(*respEvent.Date)
				event.Date = &eventDate
			}
			// Not copying redundant source id (replication id)
			// and source type (replication group)
			// into each event object
			events = append(events, event)
		}
	}
	return events, nil
}
