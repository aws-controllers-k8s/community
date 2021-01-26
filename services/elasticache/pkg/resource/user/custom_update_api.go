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

package user

import (
	"context"
	"github.com/pkg/errors"

	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	"github.com/aws/aws-controllers-k8s/pkg/requeue"
)

// If a requeue is needed or the access string needs to be modified, execute the necessary actions.
// If not, return (nil, nil), which defers to the generated code in SdkUpdate.
func (rm *resourceManager) CustomModifyUser(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {

	// requeue if necessary
	latestStatus := latest.ko.Status.Status
	if latestStatus == nil || *latestStatus != "active" {
		return nil, requeue.NeededAfter(
			errors.New("User cannot be modified as its status is not 'available'."),
			requeue.DefaultRequeueAfterDuration)
	}

	// no recent change in desired access string; do nothing and return 'latest' unmodified
	if *desired.ko.Spec.AccessString == *desired.ko.Status.LastAppliedAccessString {
		return latest, nil
	}

	// desired access string changed; defer to generated code in SdkUpdate and rely on
	// custom set output to update last applied/response access strings
	return nil, nil
}
