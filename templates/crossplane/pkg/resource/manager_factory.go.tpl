{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"sync"

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackmetrics "github.com/aws/aws-controllers-k8s/pkg/metrics"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"

	svcresource "github.com/aws/aws-controllers-k8s/services/{{ .ServiceIDClean }}/pkg/resource"
)

// resourceManagerFactory produces resourceManager objects. It implements the
// `types.AWSResourceManagerFactory` interface.
type resourceManagerFactory struct {
	sync.RWMutex
	// rmCache contains resource managers for a particular AWS account ID
	rmCache map[ackv1alpha1.AWSAccountID]*resourceManager
}

// ResourcePrototype returns an AWSResource that resource managers produced by
// this factory will handle
func (f *resourceManagerFactory) ResourceDescriptor() acktypes.AWSResourceDescriptor {
	return &resourceDescriptor{}
}

// ManagerFor returns a resource manager object that can manage resources for a
// supplied AWS account
func (f *resourceManagerFactory) ManagerFor(
	log logr.Logger,
	metrics *ackmetrics.Metrics,
	rr acktypes.AWSResourceReconciler,
	sess *session.Session,
	id ackv1alpha1.AWSAccountID,
	region ackv1alpha1.AWSRegion,
) (acktypes.AWSResourceManager, error) {
	f.RLock()
	rm, found := f.rmCache[id]
	f.RUnlock()

	if found {
		return rm, nil
	}

	f.Lock()
	defer f.Unlock()

	rm, err := newResourceManager(log, metrics, rr, sess, id, region)
	if err != nil {
		return nil, err
	}
	f.rmCache[id] = rm
	return rm, nil
}

func newResourceManagerFactory() *resourceManagerFactory {
	return &resourceManagerFactory{
		rmCache: map[ackv1alpha1.AWSAccountID]*resourceManager{},
	}
}

func init() {
	svcresource.RegisterManagerFactory(newResourceManagerFactory())
}
