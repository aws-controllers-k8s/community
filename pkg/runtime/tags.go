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
	"strings"
	"time"

	ackconfig "github.com/aws/aws-controllers-k8s/pkg/config"
	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
)

// GetDefaultTags provides Default tags (key value pairs) for given resource
func GetDefaultTags(
	config *ackconfig.Config,
	metadata *acktypes.RuntimeMetaObject,
) map[string]string {
	if metadata == nil || config == nil || len(config.ResourceTags) == 0 {
		return nil
	}
	var populatedTags = make(map[string]string)
	for _, tagKeyVal := range config.ResourceTags {
		keyVal := strings.Split(tagKeyVal, "=")
		if keyVal == nil && len(keyVal) != 2 {
			continue
		}
		key := strings.TrimSpace(keyVal[0])
		val := strings.TrimSpace(keyVal[1])
		if key == "" || val == "" {
			continue
		}
		populatedValue := expandTagValue(&val, metadata)
		populatedTags[key] = *populatedValue
	}
	if len(populatedTags) == 0 {
		return nil
	}
	return populatedTags
}

func expandTagValue(
	value *string,
	metadata *acktypes.RuntimeMetaObject,
) *string {
	if value == nil || metadata == nil {
		return nil
	}
	var expandedValue string = ""
	switch *value {
	case "%UTCNOW%":
		expandedValue = time.Now().UTC().String()
	case "%KUBERNETES_NAMESPACE%":
		expandedValue = (*metadata).GetNamespace()
	default:
		expandedValue = *value
	}
	return &expandedValue
}
