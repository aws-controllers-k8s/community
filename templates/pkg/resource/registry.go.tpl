{{ template "boilerplate" }}

package resource

import (
	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
)

var (
	reg = ackrt.NewRegistry()
)

// GetManagerFactories returns a slice of resource manager factories that are
// registered with this package
func GetManagerFactories() []acktypes.AWSResourceManagerFactory {
	return reg.GetResourceManagerFactories()
}

// RegisterManagerFactory registers a resource manager factory with the
// package's registry
func RegisterManagerFactory(f acktypes.AWSResourceManagerFactory) {
	reg.RegisterResourceManagerFactory(f)
}
