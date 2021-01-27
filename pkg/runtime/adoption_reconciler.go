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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubernetes "k8s.io/client-go/kubernetes"
	ctrlrt "sigs.k8s.io/controller-runtime"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackcfg "github.com/aws/aws-controllers-k8s/pkg/config"
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
	ackmetrics "github.com/aws/aws-controllers-k8s/pkg/metrics"
	"github.com/aws/aws-controllers-k8s/pkg/requeue"
	ackrtcache "github.com/aws/aws-controllers-k8s/pkg/runtime/cache"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// adoptionReconciler is responsible for reconciling the state of any adopted resources
// of that match any Kubernetes custom resources (CRs) that support by a gievn
// AWS service.
// It implements the upstream controller-runtime `Reconciler` interface.
type adoptionReconciler struct {
	reconciler

	// rmFactories is a map of resource manager factories, keyed by the
	// GroupKind of the resource managed by the resource manager produced by
	// that factory
	rmFactories *map[string]acktypes.AWSResourceManagerFactory
}

// BindControllerManager sets up the AWSResourceReconciler with an instance
// of an upstream controller-runtime.Manager
func (r *adoptionReconciler) BindControllerManager(mgr ctrlrt.Manager) error {
	clusterConfig := mgr.GetConfig()
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}
	r.kc = mgr.GetClient()
	r.cache = ackrtcache.New(clientset, r.log)
	r.cache.Run()
	return ctrlrt.NewControllerManagedBy(
		mgr,
	).For(
		// Read only adopted resource objects
		&ackv1alpha1.AdoptedResource{},
	).Complete(r)
}

// SecretValueFromReference fetches the value of a Secret given a
// SecretReference
func (r *adoptionReconciler) SecretValueFromReference(
	ref *corev1.SecretReference,
) (string, error) {
	// TODO(alina-kim): Implement this method :)
	return "", ackerr.NotImplemented
}

// Reconcile implements `controller-runtime.Reconciler` and handles reconciling
// a CR CRUD request
func (r *adoptionReconciler) Reconcile(req ctrlrt.Request) (ctrlrt.Result, error) {
	return r.handleReconcileError(r.reconcile(req))
}

func (r *adoptionReconciler) reconcile(req ctrlrt.Request) error {
	ctx := context.Background()
	res, err := r.getAdoptedResource(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// resource wasn't found. just ignore these.
			return nil
		}
		return err
	}

	gk := r.getTargetResourceGroupKind(res)
	// Look up the rmf for the given target resource GVK
	rmf, ok := (*r.rmFactories)[gk.String()]
	if !ok {
		return ackerr.ResourceManagerFactoryNotFound
	}

	targetRD := rmf.ResourceDescriptor()
	acctID := r.getOwnerAccountID(res)
	region := r.getRegion(res)
	roleARN := r.getRoleARN(acctID)

	sess, err := NewSession(region, roleARN, targetRD.EmptyRuntimeObject().GetObjectKind().GroupVersionKind())
	if err != nil {
		return err
	}

	// TODO(RedbackThomson): Better logging for adopted resources
	r.log.Info("starting adoption reconciliation")

	rm, err := rmf.ManagerFor(
		r.cfg, r.log, r.metrics, r, sess, acctID, region,
	)
	if err != nil {
		return err
	}

	return r.sync(ctx, rm, res)
}

func (r *adoptionReconciler) sync(
	ctx context.Context,
	rm acktypes.AWSResourceManager,
	desired *ackv1alpha1.AdoptedResource,
) error {
	fmt.Printf("RM: %v", rm)
	fmt.Printf("Res: %v", desired)
	return nil
}

// getAdoptedResource returns an AdoptedResource representing the requested Kubernetes
// namespaced object
func (r *adoptionReconciler) getAdoptedResource(
	ctx context.Context,
	req ctrlrt.Request,
) (*ackv1alpha1.AdoptedResource, error) {
	ro := &ackv1alpha1.AdoptedResource{}
	if err := r.kc.Get(ctx, req.NamespacedName, ro); err != nil {
		return nil, err
	}
	return ro, nil
}

// getTargetResourceGroupKind returns the GroupKind as specified in the spec of
// the AdoptedResource object.
func (r *adoptionReconciler) getTargetResourceGroupKind(
	res *ackv1alpha1.AdoptedResource,
) schema.GroupKind {
	return schema.GroupKind{
		Group: *res.Spec.Kubernetes.Group,
		Kind:  *res.Spec.Kubernetes.Kind,
	}
}

// handleReconcileError will handle errors from reconcile handlers, which
// respects runtime errors.
func (r *adoptionReconciler) handleReconcileError(err error) (ctrlrt.Result, error) {
	if err == nil || err == ackerr.Terminal {
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
			"requeue needed error",
			"error", requeueNeeded.Unwrap(),
		)
		return ctrlrt.Result{Requeue: true}, nil
	}

	return ctrlrt.Result{}, err
}

// getOwnerAccountID returns the AWS account that owns the supplied resource.
// The function looks to the common `Status.ACKResourceState` object, followed
// by the default AWS account ID associated with the Kubernetes Namespace in
// which the CR was created, followed by the AWS Account in which the IAM Role
// that the service controller is in.
func (r *adoptionReconciler) getOwnerAccountID(
	res *ackv1alpha1.AdoptedResource,
) ackv1alpha1.AWSAccountID {
	// look for owner account id in the namespace annotations
	namespace := res.GetNamespace()
	accID, ok := r.cache.Namespaces.GetOwnerAccountID(namespace)
	if ok {
		return ackv1alpha1.AWSAccountID(accID)
	}

	// use controller configuration
	return ackv1alpha1.AWSAccountID(r.cfg.AccountID)
}

// getRoleARN return the Role ARN that should be assumed in order to manage
// the resources.
func (r *adoptionReconciler) getRoleARN(
	acctID ackv1alpha1.AWSAccountID,
) ackv1alpha1.AWSResourceName {
	roleARN, _ := r.cache.Accounts.GetAccountRoleARN(string(acctID))
	return ackv1alpha1.AWSResourceName(roleARN)
}

// getRegion returns the AWS region that the given resource is in or should be
// created in. If the CR have a region associated with it, it is used. Otherwise
// we look for the namespace associated region, if that is set we use it. Finally
// if none of these annotations are set we use the use the region specified in the
// configuration is used
func (r *adoptionReconciler) getRegion(
	res *ackv1alpha1.AdoptedResource,
) ackv1alpha1.AWSRegion {
	// look for region in CR metadata annotations
	resAnnotations := res.GetAnnotations()
	region, ok := resAnnotations[ackv1alpha1.AnnotationRegion]
	if ok {
		return ackv1alpha1.AWSRegion(region)
	}

	// look for default region in namespace metadata annotations
	ns := res.GetNamespace()
	defaultRegion, ok := r.cache.Namespaces.GetDefaultRegion(ns)
	if ok {
		return ackv1alpha1.AWSRegion(defaultRegion)
	}

	// use controller configuration region
	return ackv1alpha1.AWSRegion(r.cfg.Region)
}

// NewAdoptionReconciler returns a new adoptionReconciler object
func NewAdoptionReconciler(
	rmFactories *map[string]acktypes.AWSResourceManagerFactory,
	log logr.Logger,
	cfg ackcfg.Config,
	metrics *ackmetrics.Metrics,
) acktypes.ACKAdoptionReconciler {
	return &adoptionReconciler{
		reconciler: reconciler{
			log:     log.WithName("ackrt"),
			cfg:     cfg,
			metrics: metrics,
		},
		rmFactories: rmFactories,
	}
}
