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

	regexp "github.com/dlclark/regexp2" // for negative lookahead support
	"github.com/iancoleman/strcase"

	"github.com/aws/aws-controllers-k8s/pkg/util"
)

type initialismTranslator struct {
	// CamelCased initialism, e.g. Tls
	camel string
	// Uppercase representation of the initialism
	upper string
	// Lowercase representation of the initialism
	lower string
	// Regular expression matching the initialism within a subject string.
	// Usually nil, unless the camel-cased initialism is a series of characters
	// that is commonly confused with a longer form of the initialism (e.g. for
	// "Id", we don't want to match "Identifier")
	re *regexp.Regexp
}

var (
	// NOTE(jaypipes): these are ordered. Some things need to be processed
	// before others. For example, we need to process "Dbi" before "Db"
	initialisms = []initialismTranslator{
		// Special... even though IDS is a valid initialism, in AWS APIs, the
		// camel-cased "Ids" refers to a set of Identifiers, so the correct
		// uppercase representation is "IDs"
		{"Ids", "IDs", "ids", nil},
		// Need to prevent "Identifier" from becoming "IDentifier",
		// and "Idle" from becoming "IDle"
		{"Id", "ID", "id", regexp.MustCompile("Id(?!entifier|le)", regexp.None)},
		// Need to prevent "DbInstance" from becoming "dbinstance" when lower
		// prefix-converted (should be dbInstance). Amazingly, even within just
		// the RDS API, there are fields named "DbiResourceId",
		// "DBInstanceIdentifier" and "DbInstanceIdentifier" (note the
		// capitalization differences). This transformer handles this
		// problematic scenario and matches only the "Dbi" case-sensitive
		// expression and converts it to "DBI" or "dbi" depending on whether
		// the initialism appears at the start of the name
		{"Dbi", "DBI", "dbi", regexp.MustCompile("Dbi", regexp.None)},
		{"Db", "DB", "db", regexp.MustCompile("Db(?!i)", regexp.None)},
		{"Db", "DB", "db", regexp.MustCompile("DB", regexp.None)},
		// Prevent "CACertificateIdentifier" from becoming
		// "cACertificateIdentifier when lower prefix-converted (should be
		// "caCertificateIdentifier")
		{"CACert", "CACert", "caCert", regexp.MustCompile("CACert", regexp.None)},
		// Prevent "MD5OfBody" from becoming "MD5OfBody" when lower
		// prefix-converted (should be "md5OfBody")
		{"MD5Of", "MD5Of", "md5Of", regexp.MustCompile("M[dD]5Of", regexp.None)},
		// Prevent "MultipartUpload" from becoming "MultIPartUpload"
		{"Ip", "IP", "ip", regexp.MustCompile("Ip(?!art)", regexp.None)},
		// Easy find-and-replacements...
		{"Acl", "ACL", "acl", nil},
		{"Acp", "ACP", "acp", nil},
		{"Api", "API", "api", nil},
		{"Arn", "ARN", "arn", nil},
		{"Asn", "ASN", "asn", nil},
		{"Aws", "AWS", "aws", nil},
		{"Az", "AZ", "az", nil},
		{"Bgp", "BGP", "bgp", nil},
		{"Cidr", "CIDR", "cidr", nil},
		{"Cpu", "CPU", "cpu", nil},
		{"Dhcp", "DHCP", "dhcp", nil},
		{"Dns", "DNS", "dns", nil},
		{"Ebs", "EBS", "ebs", nil},
		{"Ec2", "EC2", "ec2", nil},
		{"Ecr", "ECR", "ecr", nil},
		{"Efs", "EFS", "efs", nil},
		{"Eks", "EKS", "eks", nil},
		{"Fpga", "FPGA", "fpga", nil},
		{"Gpu", "GPU", "gpu", nil},
		{"Html", "HTML", "html", nil},
		{"Http", "HTTP", "http", nil},
		{"Https", "HTTPS", "https", nil},
		{"Iam", "IAM", "iam", nil},
		{"Icmp", "ICMP", "icmp", nil},
		{"Iops", "IOPS", "iops", nil},
		{"Json", "JSON", "json", nil},
		{"Jwt", "JWT", "jwt", nil},
		{"Kms", "KMS", "kms", nil},
		{"Mfa", "MFA", "mfa", nil},
		{"Sdk", "SDK", "sdk", nil},
		{"Sha256", "SHA256", "sha256", nil},
		{"Sqs", "SQS", "sns", nil},
		{"Sse", "SSE", "sse", nil},
		{"Ssl", "SSL", "ssl", nil},
		{"Tcp", "TCP", "tcp", nil},
		{"Tde", "TDE", "tde", nil},
		{"Tls", "TLS", "tls", nil},
		{"Udp", "UDP", "udp", nil},
		// Need to prevent "security" from becoming "SecURIty"
		{"Uri", "URI", "uri", regexp.MustCompile("(?!sec)uri(?!ty)|(Uri)", regexp.None)},
		{"Url", "URL", "url", nil},
		{"Vpc", "VPC", "vpc", nil},
		{"Vpn", "VPN", "vpn", nil},
		{"Vgw", "VGW", "vgw", nil},
		{"Waf", "WAF", "waf", nil},
		{"Xml", "XML", "xml", nil},
		{"Yaml", "YAML", "yaml", nil},
	}
)

var goKeywords = []string{
	"break",
	"case",
	"chan",
	"const",
	"continue",
	"default",
	"defer",
	"else",
	"fallthrough",
	"for",
	"func",
	"go",
	"goto",
	"if",
	"import",
	"interface",
	"map",
	"package",
	"range",
	"return",
	"select",
	"struct",
	"switch",
	"type",
	"var",
}

type Names struct {
	ModelOrginal string
	Original     string
	Camel        string
	CamelLower   string
	Lower        string
	Snake        string
}

func New(original string) Names {
	return Names{
		Original:   original,
		Camel:      goName(original, false, false),
		CamelLower: goName(original, true, false),
		Lower:      strings.ToLower(original),
		Snake:      goName(original, false, true),
	}
}

func goName(original string, lowerFirst bool, snake bool) (result string) {
	result = original
	if !lowerFirst {
		result = strcase.ToCamel(result)
	}
	result, err := normalizeInitialisms(result, lowerFirst, snake)
	if err != nil {
		panic(err)
	}
	if lowerFirst {
		result, err = normalizeInitialisms(strcase.ToLowerCamel(result), lowerFirst, snake)
		if err != nil {
			panic(err)
		}
	}
	if snake {
		result = strcase.ToSnake(result)
	}
	if util.InStrings(result, goKeywords) {
		result = result + "_"
	}
	return
}

// normalizeInitialisms takes a subject string and adapts the string according
// to the Go best practice naming convention for initialisms.
//
// Examples:
//
//  original   | lowerFirst | output
// ------------+ ---------- + -------------------------
// Identifier  | true       | Identifier
// Identifier  | false      | Identifier
// Id          | true       | id
// Id          | false      | ID
// SSEKMSKeyId | true       | sseKMSKeyID
// SSEKMSKeyId | false      | SSEKMSKeyID
// RoleArn     | true       | roleARN
// RoleArn     | false      | RoleARN
//
// See: https://github.com/golang/go/wiki/CodeReviewComments#initialisms
func normalizeInitialisms(original string, lowerFirst bool, snake bool) (result string, err error) {
	result = original
	for _, initTrx := range initialisms {
		if initTrx.re == nil {
			if snake {
				// If we need to snakecase, we need to look for the uppercase
				// or lowercase initialism and replace with the lowercase
				// initialism plus an underscore. For example, if original ==
				// SSEKMSId and we pass snake == true, we want to return
				// sse_kms_key_id
				toReplace := "_" + initTrx.lower + "_"
				result = strings.Replace(result, initTrx.lower, toReplace, -1)
				result = strings.Replace(result, initTrx.upper, toReplace, -1)
				continue
			}
			if lowerFirst && strings.Index(result, initTrx.upper) == 0 {
				// if we need to lowercase initialisms, check to see if the
				// initialism's capitalized form starts the string, and if so,
				// lowercase it. For example, if we get original == SSEKMSKeyId
				// and we pass lower == true, we want to return sseKMSKeyID
				result = strings.Replace(result, initTrx.upper, initTrx.lower, 1)
			}
			// Replace CamelCased initialisms with the uppercase representation
			// of the initialism EXCEPT when the CamelCased initialism appears
			// at the start of the original string and we've passed a true
			// lower parameter, in which case we lowercase just the first
			// occurrence of the CamelCased initialism
			pos := strings.Index(result, initTrx.camel)
			switch pos {
			case -1:
				continue
			case 0:
				if lowerFirst {
					toReplace := initTrx.lower
					result = strings.Replace(result, initTrx.camel, toReplace, 1)
				}
				toReplace := initTrx.upper
				if snake {
					toReplace = "_" + toReplace + "_"
				}
				result = strings.Replace(result, initTrx.camel, toReplace, -1)
			default:
				toReplace := initTrx.upper
				if snake {
					toReplace = "_" + toReplace + "_"
				}
				result = strings.Replace(result, initTrx.camel, toReplace, -1)
			}
		} else {
			match, err := initTrx.re.FindStringMatch(result)
			if err != nil {
				return "", err
			}
			if match == nil {
				continue
			}
			startFrom := match.Group.Capture.Index
			if lowerFirst {
				if startFrom == 0 {
					// The matched string appears at the start of the string --
					// e.g. IdFirstElementId. In this case, if we've asked to lower
					// the output, we need to lower only the first occurrence of
					// the matched expression, not all of it -- e.g.
					// idFirstElementID
					toReplace := initTrx.lower
					result, err = initTrx.re.Replace(result, toReplace, 0, 1)
					if err != nil {
						return "", err
					}
					match, err = initTrx.re.FindNextMatch(match)
					if err != nil {
						return "", nil
					}
					if match == nil {
						continue
					}
					startFrom = match.Group.Capture.Index
				}
			}
			toReplace := initTrx.upper
			if snake {
				toReplace = "_" + initTrx.lower + "_"
			}
			result, err = initTrx.re.Replace(result, toReplace, startFrom, -1)
			if err != nil {
				return "", err
			}
		}
	}
	if snake {
		result = strings.Replace(result, "__", "_", -1)
		result = strings.Trim(result, "_")
	}
	return result, nil
}
