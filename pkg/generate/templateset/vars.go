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

package templateset

// MetaVars contains template variables that most templates need access to
// that describe the service alias, its package name, etc
type MetaVars struct {
	// ServiceAlias contains the exact string used to identify the AWS service
	// API in the aws-sdk-go's models/apis/ directory. Note that some APIs this
	// alias does not match the ServiceID. e.g. The AWS Step Functions API has
	// a ServiceID of "SFN" and a service alias of "states"...
	ServiceAlias string
	// ServiceID is the exact string that appears in the AWS service API's
	// api-2.json descriptor file under `metadata.serviceId`
	ServiceID string
	// ServiceIDClean is the ServiceID lowercased and stripped of any
	// non-alphanumeric characters
	ServiceIDClean string
	// APIVersion contains the version of the Kubernetes API resources, e.g.
	// "v1alpha1"
	APIVersion string
	// APIGroup contains the normalized name of the Kubernetes APIGroup used
	// for custom resources, e.g. "sns.services.k8s.aws" or
	// "sfn.services.k8s.aws"
	APIGroup string
	// SDKAPIInterfaceTypeName is the name of the interface type used by the
	// aws-sdk-go services/$SERVICE/api.go file
	SDKAPIInterfaceTypeName string
	//CRDNames contains all crds names lowercased and in plural
	CRDNames []string
}
