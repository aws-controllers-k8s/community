{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	ackv1alpha1 "github.com/aws/aws-service-operator-k8s/apis/core/v1alpha1"
)

// {{ .CRD.Names.CamelLower }}ResourceIdentifiers implements the
// `aws-service-operator-k8s/pkg/types.AWSResourceIdentifiers` interface
type {{ .CRD.Names.CamelLower }}ResourceIdentifiers struct {
	meta *ackv1alpha1.ResourceMetadata
}

// ARN returns the AWS Resource Name for the backend AWS resource. If nil,
// this means the resource has not yet been created in the backend AWS
// service.
func (ri *{{ .CRD.Names.CamelLower }}ResourceIdentifiers) ARN() *ackv1alpha1.AWSResourceName {
	if ri.meta != nil {
		return ri.meta.ARN
	}
	return nil
}

// OwnerAccountID returns the AWS account identifier in which the
// backend AWS resource resides, or nil if this information is not known
// for the resource
func (ri *{{ .CRD.Names.CamelLower }}ResourceIdentifiers) OwnerAccountID() *ackv1alpha1.AWSAccountID {
	if ri.meta != nil {
		return ri.meta.OwnerAccountID
	}
	return nil
}
