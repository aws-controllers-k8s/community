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

package types

import (
	ctrlrt "sigs.k8s.io/controller-runtime"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// AWSResourceReconciler is responsible for reconciling the state of a SINGLE
// KIND of Kubernetes custom resources (CRs) that represent AWS service API
// resources.  It implements the upstream controller-runtime `Reconciler`
// interface.
//
// The upstream controller-runtime.Manager object ends up managing MULTIPLE
// controller-runtime.Controller objects (each containing a single
// AWSResourceReconciler object)s and sharing watch and informer queues across
// those controllers.
type AWSResourceReconciler interface {
	ctrlreconcile.Reconciler
	// GroupKind returns the string containing the API group and kind
	// reconciled by this reconciler
	GroupKind() string
	// BindControllerManager sets up the AWSResourceReconciler with an instance
	// of an upstream controller-runtime.Manager
	BindControllerManager(ctrlrt.Manager) error
}
