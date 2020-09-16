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
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

func NewSession(region ackv1alpha1.AWSRegion) (*session.Session, error) {
	awsCfg := aws.Config{
		Region:              aws.String(string(region)),
		STSRegionalEndpoint: endpoints.RegionalSTSEndpoint,
	}
	sess, err := session.NewSession(&awsCfg)
	if err != nil {
		return nil, err
	}
	// TODO(jaypipes): Handle throttling
	return sess, nil
}
