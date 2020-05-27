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
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"
)

// reconciler is responsible for reconciling the state of a SINGLE KIND of
// Kubernetes custom resources (CRs) that represent AWS service API resources.
// It implements the upstream controller-runtime `Reconciler` interface.
//
// The upstream controller-runtime.Manager object ends up managing MULTIPLE
// controller-runtime.Controller objects (each containing a single reconciler
// object)s and sharing watch and informer queues across those controllers.
type reconciler struct {
	kc  client.Client
	rmf acktypes.AWSResourceManagerFactory
	log logr.Logger
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
		return ReconcilerBindControllerManagerError
	}
	r.kc = mgr.GetClient()
	rf := r.rmf.ResourceFactory()
	return ctrlrt.NewControllerManagedBy(
		mgr,
	).For(
		rf.EmptyObject(),
	).Complete(r)
}

// Reconcile implements `controller-runtime.Reconciler` and handles reconciling
// a CR CRUD request
func (r *reconciler) Reconcile(req ctrlrt.Request) (ctrlrt.Result, error) {
	return r.handleReconcileError(r.reconcile(req))
}

func (r *reconciler) reconcile(req ctrlrt.Request) error {
	res, err := r.getAWSResource(req)
	if err != nil {
		return err
	}

	// TODO(jaypipes): Grab a resource manager from the factory for the AWS
	// account referenced in the object's AWS metadata.

	if res.IsDeleted() {
		// TODO(jaypipes): call rm.Delete()
		return nil
	}
	// TODO(jaypipes): reconcile the state of the object using the resource
	// manager
	return nil
}

// getAWSResource returns an AWSResource representing the requested Kubernetes
// namespaced object
func (r *reconciler) getAWSResource(
	req ctrlrt.Request,
) (acktypes.AWSResource, error) {
	ctx := context.Background()
	rf := r.rmf.ResourceFactory()
	ko := rf.EmptyObject()
	if err := r.kc.Get(ctx, req.NamespacedName, ko); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return rf.ResourceFromObject(ko), nil
}

// handleReconcileError will handle errors from reconcile handlers, which
// respects runtime errors.
func (r *reconciler) handleReconcileError(err error) (ctrlrt.Result, error) {
	if err == nil {
		return ctrlrt.Result{}, nil
	}

	var requeueAfterErr *RequeueAfterError
	if errors.As(err, &requeueAfterErr) {
		r.log.V(1).Info(
			"requeue after due to error",
			"duration", requeueAfterErr.Duration(),
			"error", requeueAfterErr.Unwrap())
		return ctrlrt.Result{RequeueAfter: requeueAfterErr.Duration()}, nil
	}

	var requeueError *RequeueError
	if errors.As(err, &requeueError) {
		r.log.V(1).Info("requeue due to error", "error", requeueError.Unwrap())
		return ctrlrt.Result{Requeue: true}, nil
	}

	return ctrlrt.Result{}, err
}

// NewReconciler returns a new reconciler object that
func NewReconciler(
	rmf acktypes.AWSResourceManagerFactory,
	log logr.Logger,
) acktypes.AWSResourceReconciler {
	return &reconciler{
		rmf: rmf,
		log: log,
	}
}
