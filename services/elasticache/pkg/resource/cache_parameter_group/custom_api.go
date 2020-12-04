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
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The number of minutes worth of events to retrieve.
	// 14 days in minutes
	eventsDuration = 20160
)

// customSetOutputDescribeCacheParameters queries cache parameters for given cache parameter group
// and sets parameter name, value for 'user' source type parameters in supplied ko.Spec
// and sets detailed parameters for both 'user', 'system' source types parameters in supplied ko.Status
func (rm *resourceManager) customSetOutputDescribeCacheParameters(
	ctx context.Context,
	cacheParameterGroupName *string,
	ko *svcapitypes.CacheParameterGroup,
) error {
	// Populate latest.ko.Spec.ParameterNameValues with latest 'user' parameter values
	source := "user"
	parameters, err := rm.describeCacheParameters(ctx, cacheParameterGroupName, &source)
	if err != nil {
		return err
	}
	parameterNameValues := []*svcapitypes.ParameterNameValue{}
	for _, p := range parameters {
		sp := svcapitypes.ParameterNameValue{
			ParameterName:  p.ParameterName,
			ParameterValue: p.ParameterValue,
		}
		parameterNameValues = append(parameterNameValues, &sp)
	}
	ko.Spec.ParameterNameValues = parameterNameValues

	// Populate latest.ko.Status.Parameters with latest all (user, system) detailed parameters
	parameters, err = rm.describeCacheParameters(ctx, cacheParameterGroupName, nil)
	if err != nil {
		return err
	}
	ko.Status.Parameters = parameters
	err = rm.customSetOutputSupplementAPIs(ctx, cacheParameterGroupName, ko)
	if err != nil {
		return err
	}
	return nil
}

func (rm *resourceManager) customSetOutputSupplementAPIs(
	ctx context.Context,
	cacheParameterGroupName *string,
	ko *svcapitypes.CacheParameterGroup,
) error {
	events, err := rm.provideEvents(ctx, cacheParameterGroupName, 20)
	if err != nil {
		return err
	}
	ko.Status.Events = events
	return nil
}

func (rm *resourceManager) provideEvents(
	ctx context.Context,
	cacheParameterGroupName *string,
	maxRecords int64,
) ([]*svcapitypes.Event, error) {
	input := &svcsdk.DescribeEventsInput{}
	input.SetSourceType("cache-parameter-group")
	input.SetSourceIdentifier(*cacheParameterGroupName)
	input.SetMaxRecords(maxRecords)
	input.SetDuration(eventsDuration)
	resp, err := rm.sdkapi.DescribeEventsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	events := []*svcapitypes.Event{}
	if resp.Events != nil {
		for _, respEvent := range resp.Events {
			event := &svcapitypes.Event{}
			if respEvent.Message != nil {
				event.Message = respEvent.Message
			}
			if respEvent.Date != nil {
				eventDate := metav1.NewTime(*respEvent.Date)
				event.Date = &eventDate
			}
			// Not copying redundant source id (replication id)
			// and source type (replication group)
			// into each event object
			events = append(events, event)
		}
	}
	return events, nil
}

// describeCacheParameters returns Cache Parameters for given Cache Parameter Group name and source
func (rm *resourceManager) describeCacheParameters(
	ctx context.Context,
	cacheParameterGroupName *string,
	source *string,
) ([]*svcapitypes.Parameter, error) {
	parameters := []*svcapitypes.Parameter{}
	var paginationMarker *string = nil
	for {
		input, err := rm.newDescribeCacheParametersRequestPayload(cacheParameterGroupName, source, paginationMarker)
		if err != nil {
			return nil, err
		}
		response, respErr := rm.sdkapi.DescribeCacheParametersWithContext(ctx, input)
		rm.metrics.RecordAPICall("READ_MANY", "DescribeCacheParameters", respErr)
		if respErr != nil {
			if awsErr, ok := ackerr.AWSError(respErr); ok && awsErr.Code() == "CacheParameterGroupNotFound" {
				return nil, ackerr.NotFound
			}
			return nil, respErr
		}

		if response.Parameters == nil || len(response.Parameters) == 0 {
			break
		}
		for _, p := range response.Parameters {
			sp := svcapitypes.Parameter{
				ParameterName:        p.ParameterName,
				ParameterValue:       p.ParameterValue,
				Source:               p.Source,
				Description:          p.Description,
				IsModifiable:         p.IsModifiable,
				DataType:             p.DataType,
				AllowedValues:        p.AllowedValues,
				MinimumEngineVersion: p.MinimumEngineVersion,
			}
			parameters = append(parameters, &sp)
		}
		paginationMarker = response.Marker
		if paginationMarker == nil || *paginationMarker == "" ||
			response.Parameters == nil || len(response.Parameters) == 0 {
			break
		}
	}

	return parameters, nil
}

// newDescribeCacheParametersRequestPayload returns SDK-specific struct for the HTTP request
// payload of the DescribeCacheParameters API to get properties that have
// given cacheParameterGroupName and given source
func (rm *resourceManager) newDescribeCacheParametersRequestPayload(
	cacheParameterGroupName *string,
	source *string,
	paginationMarker *string,
) (*svcsdk.DescribeCacheParametersInput, error) {
	res := &svcsdk.DescribeCacheParametersInput{}

	if cacheParameterGroupName != nil {
		res.SetCacheParameterGroupName(*cacheParameterGroupName)
	}
	if source != nil {
		res.SetSource(*source)
	}
	if paginationMarker != nil {
		res.SetMarker(*paginationMarker)
	}
	return res, nil
}

// resetAllParameters resets cache parameters for given CacheParameterGroup in desired custom resource.
func (rm *resourceManager) resetAllParameters(
	ctx context.Context,
	desired *resource,
) (bool, error) {
	input := &svcsdk.ResetCacheParameterGroupInput{}
	if desired.ko.Spec.CacheParameterGroupName != nil {
		input.SetCacheParameterGroupName(*desired.ko.Spec.CacheParameterGroupName)
	}
	input.SetResetAllParameters(true)

	_, err := rm.sdkapi.ResetCacheParameterGroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ResetCacheParameterGroup-ResetAllParameters", err)
	if err != nil {
		return false, err
	}
	return true, nil
}

// resetParameters resets given cache parameters for given CacheParameterGroup in desired custom resource.
func (rm *resourceManager) resetParameters(
	ctx context.Context,
	desired *resource,
	parameters []*svcapitypes.ParameterNameValue,
) (bool, error) {
	input := &svcsdk.ResetCacheParameterGroupInput{}
	if desired.ko.Spec.CacheParameterGroupName != nil {
		input.SetCacheParameterGroupName(*desired.ko.Spec.CacheParameterGroupName)
	}
	if parameters != nil && len(parameters) > 0 {
		parametersToReset := []*svcsdk.ParameterNameValue{}
		for _, parameter := range parameters {
			parameterToReset := &svcsdk.ParameterNameValue{}
			if parameter.ParameterName != nil {
				parameterToReset.SetParameterName(*parameter.ParameterName)
			}
			parametersToReset = append(parametersToReset, parameterToReset)
		}
		input.SetParameterNameValues(parametersToReset)
	}

	_, err := rm.sdkapi.ResetCacheParameterGroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ResetCacheParameterGroup", err)
	if err != nil {
		return false, err
	}
	return true, nil
}

// saveParameters saves given cache parameters for given CacheParameterGroup in desired custom resource.
// This invokes the modify API in the batches of 20 parameters.
func (rm *resourceManager) saveParameters(
	ctx context.Context,
	desired *resource,
	parameters []*svcapitypes.ParameterNameValue,
) (bool, error) {
	modifyApiBatchSize := 20
	// Paginated save: 20 parameters in single api call
	parametersToSave := []*svcsdk.ParameterNameValue{}
	for _, parameter := range parameters {
		parameterToSave := &svcsdk.ParameterNameValue{}
		if parameter.ParameterName != nil {
			parameterToSave.SetParameterName(*parameter.ParameterName)
		}
		if parameter.ParameterValue != nil {
			parameterToSave.SetParameterValue(*parameter.ParameterValue)
		}
		parametersToSave = append(parametersToSave, parameterToSave)

		if len(parametersToSave) == modifyApiBatchSize {
			done, err := rm.modifyCacheParameterGroup(ctx, desired, parametersToSave)
			if !done || err != nil {
				return false, err
			}
			// re-init to save next set of parameters
			parametersToSave = []*svcsdk.ParameterNameValue{}
		}
	}
	if len(parametersToSave) > 0 { // when len(parameters) % modifyApiBatchSize != 0
		done, err := rm.modifyCacheParameterGroup(ctx, desired, parametersToSave)
		if !done || err != nil {
			return false, err
		}
	}
	return true, nil
}

// modifyCacheParameterGroup saves given cache parameters for given CacheParameterGroup in desired custom resource.
// see 'saveParameters' method for paginated API call
func (rm *resourceManager) modifyCacheParameterGroup(
	ctx context.Context,
	desired *resource,
	parameters []*svcsdk.ParameterNameValue,
) (bool, error) {
	input := &svcsdk.ModifyCacheParameterGroupInput{}
	if desired.ko.Spec.CacheParameterGroupName != nil {
		input.SetCacheParameterGroupName(*desired.ko.Spec.CacheParameterGroupName)
	}
	if parameters != nil && len(parameters) > 0 {
		input.SetParameterNameValues(parameters)
	}
	_, err := rm.sdkapi.ModifyCacheParameterGroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ModifyCacheParameterGroup", err)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Helper method to set Condition on custom resource.
func (rm *resourceManager) setCondition(
	ko *svcapitypes.CacheParameterGroup,
	cType ackv1alpha1.ConditionType,
	cStatus corev1.ConditionStatus,
) {
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
	var condition *ackv1alpha1.Condition = nil
	for _, c := range ko.Status.Conditions {
		if c.Type == cType {
			condition = c
			break
		}
	}
	if condition == nil {
		condition = &ackv1alpha1.Condition{
			Type:   cType,
			Status: cStatus,
		}
		ko.Status.Conditions = append(ko.Status.Conditions, condition)
	} else {
		condition.Status = cStatus
	}
}
