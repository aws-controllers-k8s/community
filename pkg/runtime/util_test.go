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

package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"

	mocks "github.com/aws/aws-controllers-k8s/mocks/pkg/types"
)

func TestIsAdopted(t *testing.T) {
	require := require.New(t)

	res := &mocks.AWSResource{}
	res.On("MetaObject").Return(&metav1.ObjectMeta{
		Annotations: map[string]string{
			ackv1alpha1.AnnotationARN: "arn:aws:lambda:eu-west-1:0123456789010:function:mylambdafunction-7UXYMW16MLXP",
		},
	})
	require.True(ackrt.IsAdopted(res))

	res = &mocks.AWSResource{}
	res.On("MetaObject").Return(&metav1.ObjectMeta{})
	require.False(ackrt.IsAdopted(res))
}

func TestIsSynced(t *testing.T) {
	require := require.New(t)

	res := &mocks.AWSResource{}
	res.On("Conditions").Return([]*ackv1alpha1.Condition{
		&ackv1alpha1.Condition{
			Type:   ackv1alpha1.ConditionTypeResourceSynced,
			Status: corev1.ConditionTrue,
		},
	})
	require.True(ackrt.IsSynced(res))

	res = &mocks.AWSResource{}
	res.On("Conditions").Return([]*ackv1alpha1.Condition{
		&ackv1alpha1.Condition{
			Type:   ackv1alpha1.ConditionTypeResourceSynced,
			Status: corev1.ConditionUnknown,
		},
		&ackv1alpha1.Condition{
			Type:   ackv1alpha1.ConditionTypeResourceSynced,
			Status: corev1.ConditionFalse,
		},
	})
	require.False(ackrt.IsSynced(res))
}
