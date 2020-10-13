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
	// NotImplemented is returned when a code path isn't implemented yet
	NotImplemented = fmt.Errorf("not implemented")
	// NotFound is returned when an expected resource was not found
	NotFound = fmt.Errorf("resource not found")
	// NilResourceManagerFactory is returned when a resource manager factory
	// that has not been properly initialized is bound to a controller manager
	NilResourceManagerFactory = fmt.Errorf(
		"error binding controller manager to reconciler before " +
			"setting resource manager factory",
	)
	// AdoptedResourceNotFound is like NotFound but provides the caller with
	// information that the resource being checked for existence was
	// previously-created out of band from ACK
	AdoptedResourceNotFound = fmt.Errorf("adopted resource not found")
	// TemporaryOutOfSync is to indicate the error isn't really an error
	// but more of a marker that the status check will be performed
	// after some wait time
	TemporaryOutOfSync = fmt.Errorf(
		"Temporary out of sync, reconcile after some time")
)

// AWSError returns the type conversion for the supplied error to an aws-sdk-go
// Error interface
func AWSError(err error) (awserr.Error, bool) {
	awsErr, ok := err.(awserr.Error)
	return awsErr, ok
}
