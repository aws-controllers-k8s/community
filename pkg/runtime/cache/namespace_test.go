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

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	ctrlrtzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackrtcache "github.com/aws/aws-controllers-k8s/pkg/runtime/cache"
)

const (
	testNamespace1 = "production"
)

func TestNamespaceCache(t *testing.T) {
	// create a fake k8s client and fake watcher
	k8sClient := k8sfake.NewSimpleClientset()
	watcher := watch.NewFake()
	k8sClient.PrependWatchReactor("production", k8stesting.DefaultWatchReactor(watcher, nil))

	// New logger writing to specific buffer
	zapOptions := ctrlrtzap.Options{
		Development: true,
		Level:       zapcore.InfoLevel,
	}
	fakeLogger := ctrlrtzap.New(ctrlrtzap.UseFlagOptions(&zapOptions))

	// initlizing account cache
	namespaceCache := ackrtcache.NewNamespaceCache(k8sClient, fakeLogger)
	stopCh := make(chan struct{})

	namespaceCache.Run(stopCh)

	// Test create events
	k8sClient.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "production",
				Annotations: map[string]string{
					ackv1alpha1.AnnotationDefaultRegion:  "us-west-2",
					ackv1alpha1.AnnotationOwnerAccountID: "012345678912",
				},
			},
		},
		metav1.CreateOptions{},
	)

	time.Sleep(time.Second)

	defaultRegion, ok := namespaceCache.GetDefaultRegion("production")
	require.True(t, ok)
	require.Equal(t, "us-west-2", defaultRegion)

	ownerAccountID, ok := namespaceCache.GetOwnerAccountID("production")
	require.True(t, ok)
	require.Equal(t, "012345678912", ownerAccountID)

	// Test update events
	k8sClient.CoreV1().Namespaces().Update(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "production",
				Annotations: map[string]string{
					ackv1alpha1.AnnotationDefaultRegion:  "us-est-1",
					ackv1alpha1.AnnotationOwnerAccountID: "21987654321",
				},
			},
		},
		metav1.UpdateOptions{},
	)

	time.Sleep(time.Second)

	defaultRegion, ok = namespaceCache.GetDefaultRegion("production")
	require.True(t, ok)
	require.Equal(t, "us-est-1", defaultRegion)

	ownerAccountID, ok = namespaceCache.GetOwnerAccountID("production")
	require.True(t, ok)
	require.Equal(t, "21987654321", ownerAccountID)

	// Test delete events
	k8sClient.CoreV1().Namespaces().Delete(
		context.Background(),
		"production",
		metav1.DeleteOptions{},
	)

	time.Sleep(time.Second)

	_, ok = namespaceCache.GetDefaultRegion(testNamespace1)
	require.False(t, ok)
}
