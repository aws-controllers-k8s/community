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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdoptedResourceSpec defines the desired state of the AdoptedResource.
type AdoptedResourceSpec struct {
	// +kubebuilder:validation:Required
	Kubernetes *TargetKubernetesResource `json:"kubernetes"`
	// +kubebuilder:validation:Required
	AWS *AWSIdentifiers `json:"aws"`
}

// AdoptedResourceStatus defines the observed status of the AdoptedResource.
type AdoptedResourceStatus struct {
	AdoptionStatus *AdoptionStatus `json:"adoptionStatus,omitempty"`
}

// AdoptedResource is the schema for the AdoptedResource API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type AdoptedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AdoptedResourceSpec   `json:"spec,omitempty"`
	Status            AdoptedResourceStatus `json:"status,omitempty"`
}

// AdoptedResourceList defines a list of AdoptedResources.
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="AdoptionStatus",type=string,JSONPath=`.status.adoptionStatus`
type AdoptedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdoptedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdoptedResource{}, &AdoptedResourceList{})
}
