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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
	"github.com/aws/aws-controllers-k8s/pkg/requeue"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
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
	cfg Config
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
		rd.EmptyRuntimeObject(),
	).Complete(r)
}

// SecretValueFromReference fetches the value of a Secret given a
// SecretReference
func (r *reconciler) SecretValueFromReference(
	ref *corev1.SecretReference,
) (string, error) {
	// TODO(alina-kim): Implement this method :)
	return "", ackerr.NotImplemented
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

	acctID := r.getOwnerAccountID(res)

	r.log.WithValues(
		"account_id", acctID,
		"kind", r.rd.GroupKind().String(),
	)

	rm, err := r.rmf.ManagerFor(r, acctID)
	if err != nil {
		return err
	}

	if res.IsBeingDeleted() {
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

	isAdopted := IsAdopted(desired)

	// TODO(jaypipes): Validate all dependent resources. The AWSResource
	// interface needs to get some methods that return schema relationships,
	// first though

	latest, err := rm.ReadOne(ctx, desired)
	if err != nil {
		if err != ackerr.NotFound {
			return err
		}
		if isAdopted {
			return ackerr.AdoptedResourceNotFound
		}
		// Before we create the backend AWS service resources, let's first mark
		// the CR as being managed by ACK. Internally, this means adding a
		// finalizer to the CR; a finalizer that is removed once ACK no longer
		// manages the resource OR if the backend AWS service resource is
		// properly deleted.
		if err = r.setResourceManaged(ctx, desired); err != nil {
			return err
		}

		latest, err = rm.Create(ctx, desired)
		if err != nil {
			return err
		}
		r.log.V(1).Info(
			"reconciler.sync created new resource",
			"arn", latest.Identifiers().ARN(),
		)
	} else {
		// Check to see if the latest observed state already matches the
		// desired state and if so, simply return since there's nothing to do
		if r.rd.Equal(desired, latest) {
			return nil
		}
		diff := r.rd.Diff(desired, latest)
		r.log.V(2).Info(
			"desired resource state has changed",
			"diff", diff,
			"arn", latest.Identifiers().ARN(),
			"is_adopted", isAdopted,
		)
		latest, err = rm.Update(ctx, desired)
		if err != nil {
			return err
		}
		r.log.V(1).Info("reconciler.sync updated resource")
	}
	changedStatus, err := r.rd.UpdateCRStatus(latest)
	if err != nil {
		return err
	}
	if !changedStatus {
		return nil
	}
	err = r.kc.Status().Patch(
		ctx,
		latest.RuntimeObject(),
		client.MergeFrom(desired.RuntimeObject()),
	)
	if err != nil {
		return err
	}
	r.log.V(2).Info("patched CR status")
	return err
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
	if err = rm.Delete(ctx, observed); err != nil {
		return err
	}

	// Now that external AWS service resources have been appropriately cleaned
	// up, we remove the finalizer representing the CR is managed by ACK,
	// allowing the CR to be deleted by the Kubernetes API server
	return r.setResourceUnmanaged(ctx, observed)
}

// setResourceManaged marks the underlying CR in the supplied AWSResource with
// a finalizer that indicates the object is under ACK management and will not
// be deleted until that finalizer is removed (in setResourceUnmanaged())
func (r *reconciler) setResourceManaged(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	if r.rd.IsManaged(res) {
		return nil
	}
	orig := res.RuntimeObject().DeepCopyObject()
	r.rd.MarkManaged(res)
	err := r.kc.Patch(
		ctx,
		res.RuntimeObject(),
		client.MergeFrom(orig),
	)
	if err != nil {
		return err
	}
	r.log.V(2).Info("reconciler marked resource as managed")
	return nil
}

// setResourceUnmanaged removes a finalizer from the underlying CR in the
// supplied AWSResource that indicates the object is under ACK management. This
// allows the CR to be deleted by the Kubernetes API server.
func (r *reconciler) setResourceUnmanaged(
	ctx context.Context,
	res acktypes.AWSResource,
) error {
	if !r.rd.IsManaged(res) {
		return nil
	}
	orig := res.RuntimeObject().DeepCopyObject()
	r.rd.MarkUnmanaged(res)
	err := r.kc.Patch(
		ctx,
		res.RuntimeObject(),
		client.MergeFrom(orig),
	)
	if err != nil {
		return err
	}
	r.log.V(2).Info("reconciler removed resource from management")
	return nil
}

// getAWSResource returns an AWSResource representing the requested Kubernetes
// namespaced object
func (r *reconciler) getAWSResource(
	ctx context.Context,
	req ctrlrt.Request,
) (acktypes.AWSResource, error) {
	ro := r.rd.EmptyRuntimeObject()
	if err := r.kc.Get(ctx, req.NamespacedName, ro); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return r.rd.ResourceFromRuntimeObject(ro), nil
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

// getOwnerAccountID returns the AWS account that owns the supplied resource.
// The function looks to the common `Status.ACKResourceState` object, followed
// by the ACK OwnerAccountAccountID annotation, followed by the default AWS
// account ID associated with the Kubernetes Namespace in which the CR was
// created, followed by the AWS Account in which the IAM Role that the service
// controller is in.
func (r *reconciler) getOwnerAccountID(
	res acktypes.AWSResource,
) ackv1alpha1.AWSAccountID {
	acctID := res.Identifiers().OwnerAccountID()
	if acctID != nil {
		return *acctID
	}
	// OK, it's a new resource. Look for an override account ID annotation,
	// which indicates a cross-account resource request
	// TODO(jaypipes)
	return ackv1alpha1.AWSAccountID(r.cfg.AccountID)
}

// NewReconciler returns a new reconciler object that
func NewReconciler(
	rmf acktypes.AWSResourceManagerFactory,
	log logr.Logger,
	cfg Config,
) acktypes.AWSResourceReconciler {
	return &reconciler{
		rmf: rmf,
		rd:  rmf.ResourceDescriptor(),
		log: log,
		cfg: cfg,
	}
}
