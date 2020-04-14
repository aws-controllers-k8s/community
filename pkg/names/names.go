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

package names

import (
	"strings"

	"github.com/iancoleman/strcase"
)

var (
	initialisms = map[string]string{
		"Url":   "URL",
		"Uri":   "URI",
		"Id":    "ID",
		"Arn":   "ARN",
		"Acl":   "ACL",
		"Acp":   "ACP",
		"Mfa":   "MFA",
		"Api":   "API",
		"Sse":   "SSE",
		"Http":  "HTTP",
		"Https": "HTTPS",
		"Tls":   "TLS",
		"Ssl":   "SSL",
		"Tcp":   "TCP",
		"Udp":   "UDP",
		"Dns":   "DNS",
		"Json":  "JSON",
		"Yaml":  "YAML",
		"Xml":   "XML",
		"Html":  "HTML",
		"Aws":   "AWS",
		"Jwt":   "JWT",
		"Vpc":   "VPC",
		"Ec2":   "EC2",
		"Kms":   "KMS",
		"Sqs":   "SQS",
		"Sdk":   "SDK",
		"Ecr":   "ECR",
		"Eks":   "EKS",
		"Ebs":   "EBS",
		"Efs":   "EFS",
		"Waf":   "WAF",
		"Db":    "DB",
	}
)

type Names struct {
	Original     string
	GoUnexported string
	GoExported   string
	JSON         string
}

func New(original string) Names {
	return Names{
		Original:     original,
		GoUnexported: goName(original, true),
		GoExported:   goName(original, false),
		JSON:         jsonName(original),
	}
}

func goName(original string, lower bool) string {
	return normalizeInitialisms(strcase.ToCamel(original), lower)
}

func jsonName(original string) string {
	return strcase.ToLowerCamel(normalizeInitialisms(original, true))
}

// normalizeInitialisms takes a subject string and adapts the string according
// to the Go best practice naming convention for initialisms.
//
// See: https://github.com/golang/go/wiki/CodeReviewComments#initialisms
func normalizeInitialisms(original string, lower bool) string {
	for from, to := range initialisms {
		x := strings.Index(original, from)
		switch x {
		case -1:
			// if we need to lowercase initialisms, check to see if the
			// initialism's capitalized form starts the string, and if so,
			// lowercase it. For example, if we get original == SSEKMSKeyId and
			// we pass lower == true, we want to return sseKMSKeyID
			if lower && strings.Index(original, to) == 0 {
				original = strings.Replace(original, to, strings.ToLower(to), 1)
			}
			continue
		case 0:
			if !lower {
				original = strings.Replace(original, from, to, -1)
			} else {
				original = strings.Replace(original, from, strings.ToLower(to), -1)
			}
		default:
			original = strings.Replace(original, from, to, -1)
		}
	}
	return original
}
