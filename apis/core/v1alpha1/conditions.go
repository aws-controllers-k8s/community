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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType is a category of ResourceCondition that all CRs managed by an
// ACK service controller expose in their `Status.Conditions` collection
type ConditionType string

const (
	// ConditionTypeResourceSynced indicates the state of the resource in the
	// backend service is in sync with the ACK service controller
	ConditionTypeResourceSynced ConditionType = "ACK.ResourceSynced"
	// ConditionTypeTerminal indicates that the custom resource Spec need to be
	// updated before any further sync.
	// Examples include:
	//		- As a result of InvalidArgument in input yaml
	//		- Resource server state is "create-failed"
	ConditionTypeTerminal ConditionType = "ACK.Terminal"
)

// Condition is the common struct used by all CRDs managed by ACK service
// controllers to indicate terminal states  of the CR and its backend AWS
// service API resource
type Condition struct {
	// Type is the type of the Condition
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message *string `json:"message,omitempty"`
}
