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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ackerr "github.com/aws/aws-service-operator-k8s/pkg/errors"
	"github.com/aws/aws-service-operator-k8s/pkg/requeue"
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
	rd  acktypes.AWSResourceDescriptor
	log logr.Logger
}

// GroupKind returns the string containing the API group and kind reconciled by
// this reconciler
func (r *reconciler) GroupKind() *metav1.GroupKind {
	if r.rd == nil {
		return nil
	}
	return r.rd.GroupKind()
}

// BindControllerManager sets up the AWSResourceReconciler with an instance
// of an upstream controller-runtime.Manager
func (r *reconciler) BindControllerManager(mgr ctrlrt.Manager) error {
	if r.rmf == nil {
		return ackerr.NilResourceManagerFactory
	}
	r.kc = mgr.GetClient()
	rd := r.rmf.ResourceDescriptor()
	return ctrlrt.NewControllerManagedBy(
		mgr,
	).For(
		rd.EmptyObject(),
	).Complete(r)
}

// Reconcile implements `controller-runtime.Reconciler` and handles reconciling
// a CR CRUD request
func (r *reconciler) Reconcile(req ctrlrt.Request) (ctrlrt.Result, error) {
	return r.handleReconcileError(r.reconcile(req))
}

func (r *reconciler) reconcile(req ctrlrt.Request) error {
	ctx := context.Background()
	res, err := r.getAWSResource(ctx, req)
	if err != nil {
		return err
	}

	acctID := res.AccountID()
	rm, err := r.rmf.ManagerFor(acctID)

	if res.IsDeleted() {
		return r.cleanup(ctx, rm, res)
	}

	return r.sync(ctx, rm, res)
}

// sync ensures that the supplied AWSResource's backing API resource
// matches the supplied desired state
func (r *reconciler) sync(
	ctx context.Context,
	rm acktypes.AWSResourceManager,
	desired acktypes.AWSResource,
) error {
	var latest acktypes.AWSResource // the newly created or mutated resource

	// TODO(jaypipes): Handle all dependent resources. The AWSResource
	// interface needs to get some methods that return schema relationships,
	// first though

	latest, err := rm.ReadOne(ctx, desired)
	if err != nil {
		if err != ackerr.NotFound {
			return err
		}
		latest, err = rm.Create(ctx, desired)
		if err != nil {
			return err
		}
		r.log.V(1).Info(
			"reconciler.sync created new resource",
			"kind", r.rd.GroupKind().String(),
			"account_id", latest.AccountID(),
		)
	} else {
		// Check to see if the latest observed state already matches the
		// desired state and if so, simply return since there's nothing to do
		if r.rd.Equal(desired, latest) {
			return nil
		}
		diff := r.rd.Diff(desired, latest)
		r.log.V(1).Info("desired resource state has changed",
			"kind", r.rd.GroupKind().String(),
			"account_id", latest.AccountID(),
			"diff", diff,
		)
		latest, err = rm.Update(ctx, desired)
		if err != nil {
			return err
		}
		r.log.V(1).Info(
			"reconciler.sync updated resource",
			"kind", r.rd.GroupKind().String(),
			"account_id", latest.AccountID(),
		)
	}
	// TODO(jaypipes): Set the CRD's Status and other stuff to the latest
	// resource's object
	return nil
}

// cleanup ensures that the supplied AWSResource's backing API resource is
// destroyed along with all child dependent resources
func (r *reconciler) cleanup(
	ctx context.Context,
	rm acktypes.AWSResourceManager,
	current acktypes.AWSResource,
) error {
	// TODO(jaypipes): Handle all dependent resources. The AWSResource
	// interface needs to get some methods that return schema relationships,
	// first though
	observed, err := rm.ReadOne(ctx, current)
	if err != nil {
		if err == ackerr.NotFound {
			return nil
		}
		return err
	}
	return rm.Delete(ctx, observed)
}

// getAWSResource returns an AWSResource representing the requested Kubernetes
// namespaced object
func (r *reconciler) getAWSResource(
	ctx context.Context,
	req ctrlrt.Request,
) (acktypes.AWSResource, error) {
	ko := r.rd.EmptyObject()
	if err := r.kc.Get(ctx, req.NamespacedName, ko); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return r.rd.ResourceFromObject(ko), nil
}

// handleReconcileError will handle errors from reconcile handlers, which
// respects runtime errors.
func (r *reconciler) handleReconcileError(err error) (ctrlrt.Result, error) {
	if err == nil {
		return ctrlrt.Result{}, nil
	}

	var requeueNeededAfter *requeue.RequeueNeededAfter
	if errors.As(err, &requeueNeededAfter) {
		after := requeueNeededAfter.Duration()
		r.log.V(1).Info(
			"requeue needed after error",
			"error", requeueNeededAfter.Unwrap(),
			"after", after,
		)
		return ctrlrt.Result{RequeueAfter: after}, nil
	}

	var requeueNeeded *requeue.RequeueNeeded
	if errors.As(err, &requeueNeeded) {
		r.log.V(1).Info(
			"requeue needed after error",
			"error", requeueNeeded.Unwrap(),
		)
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
		rd:  rmf.ResourceDescriptor(),
		log: log,
	}
}
