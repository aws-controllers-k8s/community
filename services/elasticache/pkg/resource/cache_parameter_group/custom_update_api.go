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
	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
)

// Implements specialized logic for update CacheParameterGroup.
func (rm *resourceManager) customUpdateCacheParameterGroup(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	desiredParameters := desired.ko.Spec.ParameterNameValues
	latestParameters := latest.ko.Spec.ParameterNameValues

	updated := false
	var err error
	// Update
	if (desiredParameters == nil || len(desiredParameters) == 0) &&
		(latestParameters != nil && len(latestParameters) > 0) {
		updated, err = rm.resetAllParameters(ctx, desired)
		if !updated || err != nil {
			return nil, err
		}
	} else {
		removedParameters, modifiedParameters, addedParameters := rm.provideDelta(desiredParameters, latestParameters)
		if removedParameters != nil && len(removedParameters) > 0 {
			updated, err = rm.resetParameters(ctx, desired, removedParameters)
			if !updated || err != nil {
				return nil, err
			}
		}
		if modifiedParameters != nil && len(modifiedParameters) > 0 {
			updated, err = rm.saveParameters(ctx, desired, modifiedParameters)
			if !updated || err != nil {
				return nil, err
			}
		}
		if addedParameters != nil && len(addedParameters) > 0 {
			updated, err = rm.saveParameters(ctx, desired, addedParameters)
			if !updated || err != nil {
				return nil, err
			}
		}
	}
	if updated {
		rm.setStatusDefaults(latest.ko)
		// Populate ko.Spec.ParameterNameValues with latest parameter values
		source := "user"
		parameterNameValues, err := rm.describeCacheParameters(ctx, desired.ko.Spec.CacheParameterGroupName, &source)
		if err != nil {
			return nil, err
		}
		latest.ko.Spec.ParameterNameValues = parameterNameValues
	}
	return latest, nil
}

// provideDelta compares given desired and latest Parameters and returns
// removedParameters, modifiedParameters, addedParameters
func (rm *resourceManager) provideDelta(
	desiredParameters []*svcapitypes.ParameterNameValue,
	latestParameters []*svcapitypes.ParameterNameValue,
) ([]*svcapitypes.ParameterNameValue, []*svcapitypes.ParameterNameValue, []*svcapitypes.ParameterNameValue) {

	desiredPametersMap := map[string]*svcapitypes.ParameterNameValue{}
	for _, parameter := range desiredParameters {
		p := *parameter
		desiredPametersMap[*p.ParameterName] = &p
	}
	latestPametersMap := map[string]*svcapitypes.ParameterNameValue{}
	for _, parameter := range latestParameters {
		p := *parameter
		latestPametersMap[*p.ParameterName] = &p
	}

	removedParameters := []*svcapitypes.ParameterNameValue{}  // available in latest but not found in desired
	modifiedParameters := []*svcapitypes.ParameterNameValue{} // available in both desired, latest but values differ
	addedParameters := []*svcapitypes.ParameterNameValue{}    // available in desired but not found in latest
	for latestParameterName, latestParameterNameValue := range latestPametersMap {
		desiredParameterNameValue, found := desiredPametersMap[latestParameterName]
		if found && desiredParameterNameValue != nil &&
			desiredParameterNameValue.ParameterValue != nil && *desiredParameterNameValue.ParameterValue != ""{
			if *desiredParameterNameValue.ParameterValue != *latestParameterNameValue.ParameterValue {
				// available in both desired, latest but values differ
				modified := *desiredParameterNameValue
				modifiedParameters = append(modifiedParameters, &modified)
			}
		} else {
			// available in latest but not found in desired
			removed := *latestParameterNameValue
			removedParameters = append(removedParameters, &removed)
		}
	}
	for desiredParameterName, desiredParameterNameValue := range desiredPametersMap {
		_, found := latestPametersMap[desiredParameterName]
		if !found && desiredParameterNameValue != nil {
			// available in desired but not found in latest
			added := *desiredParameterNameValue
			if added.ParameterValue != nil && *added.ParameterValue != "" {
				addedParameters = append(addedParameters, &added)
			}
		}
	}
	return removedParameters, modifiedParameters, addedParameters
}
