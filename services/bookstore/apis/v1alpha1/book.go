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

// TODO(jaypipes): Move the identifier and account information into a common
// acktypes.Metadata struct

// BookSpec defines the desired state of Book
type BookSpec struct {
	// Name is the Bookstore API Book object's name.
	// If unspecified or empty, it defaults to be "${name}" of k8s Book
	// +optional
	Name *string `json:"name,omitempty"`
	// The AWS IAM account ID of the book owner.
	// Required if the account ID is not your own.
	// +optional
	Owner *string `json:"owner,omitempty"`
}

// BookStatus defines the observed state of Book
type BookStatus struct {
	// ARN is the Bookstore API Book object's Amazon Resource Name
	// +optional
	ARN *string `json:"ARN,omitempty"`
}

// Book implements sigs.k8s.io/apimachinery/pkg/runtime.Object
type Book struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookSpec   `json:"spec,omitempty"`
	Status BookStatus `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Book{})
}
