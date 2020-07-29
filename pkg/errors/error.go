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

package errors

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

var (
	NotFound                  = fmt.Errorf("resource not found")
	NilResourceManagerFactory = fmt.Errorf(
		"error binding controller manager to reconciler before " +
			"setting resource manager factory",
	)
	AdoptedResourceNotFound = fmt.Errorf("adopted resource not found")
)

// AWSError returns the type conversion for the supplied error to an aws-sdk-go
// Error interface
func AWSError(err error) (awserr.Error, bool) {
	awsErr, ok := err.(awserr.Error)
	return awsErr, ok
}
