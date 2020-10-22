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

package cache

import (
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	k8scache "k8s.io/client-go/tools/cache"
)

const (
	// ACKRoleAccountMap is the name of the configmap map object storing
	// all the AWS Account IDs associated with their AWS Role ARNs.
	ACKRoleAccountMap = "ack-role-account-map"
)

// AccountCache is responsible for caching the CARM configmap
// data. It is listening to all the events related to the CARM map and
// make the changes accordingly.
type AccountCache struct {
	sync.RWMutex

	log logr.Logger

	// ConfigMap informer
	informer k8scache.SharedInformer
	roleARNs map[string]string
}

// NewAccountCache makes a new AccountCache from a client.Interface
// and a logr.Logger
func NewAccountCache(clientset kubernetes.Interface, log logr.Logger) *AccountCache {
	sharedInformer := informersv1.NewConfigMapInformer(
		clientset,
		currentNamespace,
		informerResyncPeriod,
		k8scache.Indexers{},
	)
	return &AccountCache{
		informer: sharedInformer,
		log:      log.WithName("cache.account"),
		roleARNs: make(map[string]string),
	}
}

// resourceMatchACKRoleAccountConfigMap verifies if a resource is
// the CARM configmap. It verifies the name, namespace and object type.
func resourceMatchACKRoleAccountsConfigMap(raw interface{}) bool {
	object, ok := raw.(*corev1.ConfigMap)
	return ok && object.ObjectMeta.Name == ACKRoleAccountMap
}

// Run adds the default event handler functions to the SharedInformer and
// runs the informer to begin processing items.
func (c *AccountCache) Run(stopCh <-chan struct{}) {
	c.informer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if resourceMatchACKRoleAccountsConfigMap(obj) {
				cm := obj.(*corev1.ConfigMap)
				object := cm.DeepCopy()
				c.updateAccountRoleData(object.Data)
				c.log.V(1).Info("created account config map", "name", cm.ObjectMeta.Name)
			}
		},
		UpdateFunc: func(orig, desired interface{}) {
			if resourceMatchACKRoleAccountsConfigMap(desired) {
				cm := desired.(*corev1.ConfigMap)
				object := cm.DeepCopy()
				//TODO(a-hilaly): compare data checksum before updating the cache
				c.updateAccountRoleData(object.Data)
				c.log.V(1).Info("updated account config map", "name", cm.ObjectMeta.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if resourceMatchACKRoleAccountsConfigMap(obj) {
				cm := obj.(*corev1.ConfigMap)
				newMap := make(map[string]string)
				c.updateAccountRoleData(newMap)
				c.log.V(1).Info("deleted account config map", "name", cm.ObjectMeta.Name)
			}
		},
	})
	go c.informer.Run(stopCh)
}

// GetAccountRoleARN queries the AWS accountID associated Role ARN
// from the cached CARM configmap. This function is thread safe.
func (c *AccountCache) GetAccountRoleARN(accountID string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	roleARN, ok := c.roleARNs[accountID]
	return roleARN, ok && roleARN != ""
}

// updateAccountRoleData updates the CARM map. This function is thread safe.
func (c *AccountCache) updateAccountRoleData(data map[string]string) {
	c.Lock()
	defer c.Unlock()
	c.roleARNs = data
}
