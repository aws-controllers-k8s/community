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

package types

import (
	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
)

// AWSResourceIdentifiers has methods that returns common identifying
// information about a resource
type AWSResourceIdentifiers interface {
	// OwnerAccountID returns the AWS account identifier in which the
	// backend AWS resource resides, or nil if this information is not known
	// for the resource
	OwnerAccountID() *ackv1alpha1.AWSAccountID
}
