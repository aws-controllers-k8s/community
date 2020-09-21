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

	ackrtcache "github.com/aws/aws-controllers-k8s/pkg/runtime/cache"
)

const (
	testNamespace = "ack-system"

	testAccount1    = "012345678912"
	testAccountARN1 = "arn:aws:iam::012345678912:role/S3Access"
	testAccount2    = "219876543210"
	testAccountARN2 = "arn:aws:iam::012345678912:role/root"
)

func TestAccountCache(t *testing.T) {
	accountsMap1 := map[string]string{
		testAccount1: testAccountARN1,
	}

	accountsMap2 := map[string]string{
		testAccount1: testAccountARN1,
		testAccount2: testAccountARN2,
	}

	// create a fake k8s client and a fake watcher
	k8sClient := k8sfake.NewSimpleClientset()
	watcher := watch.NewFake()
	k8sClient.PrependWatchReactor("configMaps", k8stesting.DefaultWatchReactor(watcher, nil))

	zapOptions := ctrlrtzap.Options{
		Development: true,
		Level:       zapcore.InfoLevel,
	}
	fakeLogger := ctrlrtzap.New(ctrlrtzap.UseFlagOptions(&zapOptions))

	// initlizing account cache
	accountCache := ackrtcache.NewAccountCache(k8sClient, fakeLogger)
	stopCh := make(chan struct{})
	accountCache.Run(stopCh)

	// Test create events
	k8sClient.CoreV1().ConfigMaps(testNamespace).Create(
		context.Background(),
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "random-map",
			},
			Data: accountsMap1,
		},
		metav1.CreateOptions{},
	)

	time.Sleep(time.Second)

	_, ok := accountCache.GetAccountRoleARN("random-account")
	require.False(t, ok)

	k8sClient.CoreV1().ConfigMaps(testNamespace).Create(
		context.Background(),
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ackrtcache.ACKRoleAccountMap,
				Namespace: "ack-system",
			},
			Data: accountsMap1,
		},
		metav1.CreateOptions{},
	)

	time.Sleep(time.Second)

	roleARN, ok := accountCache.GetAccountRoleARN(testAccount1)
	require.True(t, ok)
	require.Equal(t, roleARN, testAccountARN1)

	_, ok = accountCache.GetAccountRoleARN(testAccount2)
	require.False(t, ok)

	// Test update events
	k8sClient.CoreV1().ConfigMaps("ack-system").Update(
		context.Background(),
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ackrtcache.ACKRoleAccountMap,
				Namespace: "ack-system",
			},
			Data: accountsMap2,
		},
		metav1.UpdateOptions{},
	)

	time.Sleep(time.Second)

	roleARN, ok = accountCache.GetAccountRoleARN(testAccount1)
	require.True(t, ok)
	require.Equal(t, roleARN, testAccountARN1)

	roleARN, ok = accountCache.GetAccountRoleARN(testAccount2)
	require.True(t, ok)
	require.Equal(t, roleARN, testAccountARN2)

	// Test delete events
	k8sClient.CoreV1().ConfigMaps("ack-system").Delete(
		context.Background(),
		ackrtcache.ACKRoleAccountMap,
		metav1.DeleteOptions{},
	)

	time.Sleep(time.Second)

	_, ok = accountCache.GetAccountRoleARN(testAccount1)
	require.False(t, ok)
	_, ok = accountCache.GetAccountRoleARN(testAccount2)
	require.False(t, ok)

}
