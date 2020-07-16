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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"

	mocks "github.com/aws/aws-controllers-k8s/mocks/pkg/types"
	bookstoretypes "github.com/aws/aws-controllers-k8s/services/bookstore/apis/v1alpha1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = bookstoretypes.AddToScheme(scheme)
}

type fakeManager struct{}

func (m *fakeManager) Add(ctrlmanager.Runnable) error                                 { return nil }
func (m *fakeManager) Elected() <-chan struct{}                                       { return nil }
func (m *fakeManager) SetFields(interface{}) error                                    { return nil }
func (m *fakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error { return nil }
func (m *fakeManager) AddHealthzCheck(name string, check healthz.Checker) error       { return nil }
func (m *fakeManager) AddReadyzCheck(name string, check healthz.Checker) error        { return nil }
func (m *fakeManager) Start(<-chan struct{}) error                                    { return nil }
func (m *fakeManager) GetConfig() *rest.Config                                        { return nil }
func (m *fakeManager) GetScheme() *runtime.Scheme                                     { return scheme }
func (m *fakeManager) GetClient() client.Client                                       { return nil }
func (m *fakeManager) GetFieldIndexer() client.FieldIndexer                           { return nil }
func (m *fakeManager) GetCache() cache.Cache                                          { return nil }
func (m *fakeManager) GetEventRecorderFor(name string) record.EventRecorder           { return nil }
func (m *fakeManager) GetRESTMapper() meta.RESTMapper                                 { return nil }
func (m *fakeManager) GetAPIReader() client.Reader                                    { return nil }
func (m *fakeManager) GetWebhookServer() *webhook.Server                              { return nil }

func TestServiceController(t *testing.T) {
	require := require.New(t)

	rd := &mocks.AWSResourceDescriptor{}
	rd.On("GroupKind").Return(
		&metav1.GroupKind{
			Group: "bookstore.services.k8s.aws",
			Kind:  "Book",
		},
	)
	rd.On("EmptyRuntimeObject").Return(
		&bookstoretypes.Book{},
	)

	rmf := &mocks.AWSResourceManagerFactory{}
	rmf.On("ResourceDescriptor").Return(rd)

	reg := ackrt.NewRegistry()
	reg.RegisterResourceManagerFactory(rmf)

	sc := ackrt.NewServiceController("bookstore", "bookstore.services.k8s.aws")
	require.NotNil(sc)

	sc.WithResourceManagerFactories(reg.GetResourceManagerFactories())

	recons := sc.GetReconcilers()

	// Before we bind to a controller manager, there are no reconcilers in the
	// service controller
	require.Empty(recons)

	mgr := &fakeManager{}
	err := sc.BindControllerManager(mgr)
	require.Nil(err)

	recons = sc.GetReconcilers()
	require.NotEmpty(recons)

	foundBookRecon := false
	for _, recon := range recons {
		if recon.GroupKind().String() == "Book.bookstore.services.k8s.aws" {
			foundBookRecon = true
		}
	}
	require.True(foundBookRecon)
	rd.AssertCalled(t, "EmptyRuntimeObject")
}
