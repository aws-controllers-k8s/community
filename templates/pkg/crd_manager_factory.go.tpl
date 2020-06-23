{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"sync"

	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
	acktypes "github.com/aws/aws-service-operator-k8s/pkg/types"

	svcresource "github.com/aws/aws-service-operator-k8s/services/{{ .ServiceAlias }}/pkg/resource"
)

// {{ .CRD.Names.CamelLower }}ResourceManagerFactory produces {{ .CRD.Names.CamelLower }}ResourceManager objects. It
// implements the `types.AWSResourceManagerFactory` interface.
type {{ .CRD.Names.CamelLower }}ResourceManagerFactory struct {
	sync.RWMutex
	// rmCache contains resource managers for a particular AWS account ID
	rmCache map[ackv1alpha1.AWSAccountID]*{{ .CRD.Names.CamelLower }}ResourceManager
}

// ResourcePrototype returns an AWSResource that resource managers produced by
// this factory will handle
func (f *{{ .CRD.Names.CamelLower }}ResourceManagerFactory) ResourceDescriptor() acktypes.AWSResourceDescriptor {
	return &{{ .CRD.Names.CamelLower }}ResourceDescriptor{}
}

// ManagerFor returns a resource manager object that can manage resources for a
// supplied AWS account
func (f *{{ .CRD.Names.CamelLower }}ResourceManagerFactory) ManagerFor(
	id ackv1alpha1.AWSAccountID,
) (acktypes.AWSResourceManager, error) {
	f.RLock()
	rm, found := f.rmCache[id]
	f.RUnlock()

	if found {
		return rm, nil
	}

	f.Lock()
	defer f.Unlock()

	rm, err := new{{ .CRD.Kind }}ResourceManager(id)
	if err != nil {
		return nil, err
	}
	f.rmCache[id] = rm
	return rm, nil
}

func new{{ .CRD.Kind }}ResourceManagerFactory() *{{ .CRD.Names.CamelLower }}ResourceManagerFactory {
	return &{{ .CRD.Names.CamelLower }}ResourceManagerFactory{
		rmCache: map[ackv1alpha1.AWSAccountID]*{{ .CRD.Names.CamelLower }}ResourceManager{},
	}
}

func init() {
	svcresource.RegisterManagerFactory(new{{ .CRD.Kind }}ResourceManagerFactory())
}
