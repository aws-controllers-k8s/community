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

package log

import (
	"github.com/go-logr/logr"

	acktypes "github.com/aws/aws-controllers-k8s/pkg/types"
)

// AdaptResource returns a logger with log values set for the resource's kind,
// namespace, name, etc
func AdaptResource(
	log logr.Logger,
	res acktypes.AWSResource,
	additionalValues ...interface{},
) logr.Logger {
	metaObj := res.MetaObject()
	ns := metaObj.GetNamespace()
	resName := metaObj.GetName()
	generation := metaObj.GetGeneration()
	rtObj := res.RuntimeObject()
	kind := rtObj.GetObjectKind().GroupVersionKind().Kind
	vals := []interface{}{
		"kind", kind,
		"namespace", ns,
		"name", resName,
		"generation", generation,
	}
	if len(additionalValues) > 0 {
		for _, v := range additionalValues {
			vals = append(vals, v)
		}
	}
	return log.WithValues(vals...)
}

// DebugResource writes a supplied log message about a resource that includes a
// set of standard log values for the resource's kind, namespace, name, etc
func DebugResource(
	log logr.Logger,
	res acktypes.AWSResource,
	msg string,
	additionalValues ...interface{},
) {
	AdaptResource(log, res, additionalValues...).V(1).Info(msg)
}

// InfoResource writes a supplied log message about a resource that includes a
// set of standard log values for the resource's kind, namespace, name, etc
func InfoResource(
	log logr.Logger,
	res acktypes.AWSResource,
	msg string,
	additionalValues ...interface{},
) {
	AdaptResource(log, res, additionalValues...).V(0).Info(msg)
}
