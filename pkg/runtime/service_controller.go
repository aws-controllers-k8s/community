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
	"sync"

	"github.com/go-logr/logr"
	ctrlrt "sigs.k8s.io/controller-runtime"

	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
)

// ServiceController wraps a number of `controller-runtime.Reconciler` that are
// related to a specific AWS service API.
type ServiceController struct {
	metaLock sync.RWMutex
	// ServiceAlias is a string with the alias of the service API, e.g. "s3"
	ServiceAlias string
	// ServiceAPIGroup is a string with the full DNS-correct API group that
	// this service controller manages, e.g. "s3.services.k8s.aws"
	ServiceAPIGroup string
	// rmFactories is a map of resource manager factories, keyed by the
	// GroupKind of the resource managed by the resource manager produced by
	// that factory
	rmFactories map[string]acktypes.AWSResourceManagerFactory
	// reconcilers is a map containing AWSResourceReconciler objects that are
	// bound to the `controller-runtime.Manager` in `BindControllerManager`
	reconcilers []acktypes.AWSResourceReconciler
	// log refers to the logr.Logger object handling logging for the service
	// controller
	log logr.Logger
}

// GetReconcilers returns a slice of types.AWSResourceReconcilers associated
// with this service controller
func (c *ServiceController) GetReconcilers() []acktypes.AWSResourceReconciler {
	c.metaLock.RLock()
	defer c.metaLock.RUnlock()
	return c.reconcilers
}

// WithLogger sets up the service controller with the supplied logger
func (c *ServiceController) WithLogger(log logr.Logger) *ServiceController {
	c.log = log
	return c
}

// WithResourceManagerFactories sets the controller up to manage resources with
// a set of supplied factories
func (c *ServiceController) WithResourceManagerFactories(
	rmfs []acktypes.AWSResourceManagerFactory,
) *ServiceController {
	c.metaLock.Lock()
	defer c.metaLock.Unlock()

	if c.rmFactories == nil {
		c.rmFactories = make(
			map[string]acktypes.AWSResourceManagerFactory,
			len(rmfs),
		)
	}

	for _, rmf := range rmfs {
		c.rmFactories[rmf.ResourceDescriptor().GroupKind().String()] = rmf
	}
	return c
}

// BindControllerManager takes a `controller-runtime.Manager`, creates all the
// AWSResourceReconcilers needed for the service and binds all of the
// reconcilers within the service controller with that manager
func (c *ServiceController) BindControllerManager(mgr ctrlrt.Manager, cfg Config) error {
	c.metaLock.Lock()
	defer c.metaLock.Unlock()
	for _, rmf := range c.rmFactories {
		rec := NewReconciler(rmf, c.log, cfg)
		if err := rec.BindControllerManager(mgr); err != nil {
			return err
		}
		c.reconcilers = append(c.reconcilers, rec)
	}
	return nil
}

// NewServiceController returns a new ServiceController instance
func NewServiceController(
	svcAlias string,
	svcAPIGroup string,
) *ServiceController {
	return &ServiceController{
		ServiceAlias:    svcAlias,
		ServiceAPIGroup: svcAPIGroup,
	}
}
