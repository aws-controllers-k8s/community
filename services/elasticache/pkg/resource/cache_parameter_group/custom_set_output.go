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

package cache_parameter_group

import (
	"context"
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	corev1 "k8s.io/api/core/v1"
)

func (rm *resourceManager) CustomDescribeCacheParameterGroupsSetOutput(
	ctx context.Context,
	r *resource,
	resp *svcsdk.DescribeCacheParameterGroupsOutput,
	ko *svcapitypes.CacheParameterGroup,
) (*svcapitypes.CacheParameterGroup, error) {
	// Retrieve parameters using DescribeCacheParameters API and populate ko.Status.ParameterNameValues
	if len(resp.CacheParameterGroups) == 0 {
		return ko, nil
	}
	cpg := resp.CacheParameterGroups[0]
	// Populate latest.ko.Spec.ParameterNameValues with latest parameter values
	// Populate latest.ko.Status.Parameters with latest detailed parameters
	error := rm.customSetOutputDescribeCacheParameters(ctx, cpg.CacheParameterGroupName, ko)
	if error != nil {
		return nil, error
	}
	return ko, nil
}

func (rm *resourceManager) CustomCreateCacheParameterGroupSetOutput(
	ctx context.Context,
	r *resource,
	resp *svcsdk.CreateCacheParameterGroupOutput,
	ko *svcapitypes.CacheParameterGroup,
) (*svcapitypes.CacheParameterGroup, error) {
	if r.ko.Spec.ParameterNameValues != nil && len(r.ko.Spec.ParameterNameValues) != 0 {
		// Spec has parameters name and values. Create API does not save these, but Modify API does.
		// Thus, Create needs to be followed by Modify call to save parameters from Spec.
		// Setting synched condition to false, so that reconciler gets invoked again
		// and modify logic gets executed.
		rm.setCondition(ko, ackv1alpha1.ConditionTypeResourceSynced, corev1.ConditionFalse)
	}
	return ko, nil
}
