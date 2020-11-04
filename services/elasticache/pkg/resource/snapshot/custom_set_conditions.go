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
	corev1 "k8s.io/api/core/v1"
)

// CustomUpdateConditions sets conditions (terminal) on supplied snapshot
// it examines supplied resource to determine conditions.
// It returns true if conditions are updated
func (rm *resourceManager) CustomUpdateConditions(
	ko *svcapitypes.Snapshot,
	r *resource,
	err error,
) bool {
	snapshotStatus := r.ko.Status.SnapshotStatus
	if snapshotStatus == nil || *snapshotStatus != "failed" {
		return false
	}
	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	} else {
		for _, condition := range ko.Status.Conditions {
			if condition.Type == ackv1alpha1.ConditionTypeTerminal {
				terminalCondition = condition
				break
			}
		}
		if terminalCondition != nil && terminalCondition.Status == corev1.ConditionTrue {
			// some other exception already put the resource in terminal condition
			return false
		}
	}
	if terminalCondition == nil {
		terminalCondition = &ackv1alpha1.Condition{
			Type: ackv1alpha1.ConditionTypeTerminal,
		}
		ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
	}
	terminalCondition.Status = corev1.ConditionTrue
	errorMessage := "Snapshot status: failed"
	terminalCondition.Message = &errorMessage
	return true
}
