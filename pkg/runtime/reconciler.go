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

package runtime

import (
	"fmt"

	ctrlrt "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-service-operator-k8s/pkg/types"
)

// reconciler is responsible for reconciling the state of a SINGLE KIND of
// Kubernetes custom resources (CRs) that represent AWS service API resources.
// It implements the upstream controller-runtime `Reconciler` interface.
//
// The upstream controller-runtime.Manager object ends up managing MULTIPLE
// controller-runtime.Controller objects (each containing a single reconciler
// object)s and sharing watch and informer queues across those controllers.
type reconciler struct {
	rmf types.AWSResourceManagerFactory
}

// GroupKind returns the string containing the API group and kind reconciled by
// this reconciler
func (r *reconciler) GroupKind() string {
	if r.rmf == nil {
		return ""
	}
	return r.rmf.GroupKind()
}

// BindControllerManager sets up the AWSResourceReconciler with an instance
// of an upstream controller-runtime.Manager
func (r *reconciler) BindControllerManager(mgr ctrlrt.Manager) error {
	if r.rmf == nil {
		return fmt.Errorf("tried binding controller manager to reconciler before resource manager factory")
	}
	return ctrlrt.NewControllerManagedBy(
		mgr,
	).For(
		r.rmf.ObjectPrototype(),
	).Complete(r)
}

// Reconcile implements `controller-runtime.Reconciler` and handles reconciling
// a CR CRUD request
func (r *reconciler) Reconcile(req ctrlrt.Request) (ctrlrt.Result, error) {
	return ctrlrt.Result{}, nil
}

func NewReconciler(rmf types.AWSResourceManagerFactory) *reconciler {
	return &reconciler{
		rmf: rmf,
	}
}
