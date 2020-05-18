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

import "github.com/aws/aws-sdk-go/aws/session"

func NewSession() (*session.Session, error) {
	// NOTE(jaypipes): session.NewSession() is needed for the STS::AssumeRole
	// stuff we will need to do...
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	// TODO(jaypipes): Handling all common region endpoint, throttling
	// configuration, TLS, etc
	return sess, nil
}
