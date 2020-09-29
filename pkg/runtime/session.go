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
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession returns a new session object. Buy default the returned session is
// created using pod IRSA environment variables. If assumeRoleARN is not empty,
// NewSession will call STS::AssumeRole and use the returned credentials to create
// the session.
func NewSession(
	region ackv1alpha1.AWSRegion,
	assumeRoleARN ackv1alpha1.AWSResourceName,
) (*session.Session, error) {
	awsCfg := aws.Config{
		Region:              aws.String(string(region)),
		STSRegionalEndpoint: endpoints.RegionalSTSEndpoint,
	}
	sess, err := session.NewSession(&awsCfg)
	if err != nil {
		return nil, err
	}

	if assumeRoleARN != "" {
		// call STS::AssumeRole
		creds := stscreds.NewCredentials(sess, string(assumeRoleARN))
		// recreate session with the new credentials
		awsCfg.Credentials = creds
		sess, err = session.NewSession(&awsCfg)
		if err != nil {
			return nil, err
		}
	}

	// TODO(jaypipes): Handle throttling
	return sess, nil
}
