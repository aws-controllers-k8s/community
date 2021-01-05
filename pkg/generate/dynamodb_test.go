// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	 http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestDynamoDB_Table(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "dynamodb")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Table", crds)
	require.NotNil(crd)

	// The DynamoDB Table API has these operations:
	//
	// * CreateTable
	// * DeleteTable
	// * DescribeTable
	// * ListTables
	// * UpdateTable
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.Update)

	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)

	// The DescribeTable operation has the following definition:
	//
	//    "DescribeTable":{
	//      "name":"DescribeTable",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"DescribeTableInput"},
	//      "output":{"shape":"DescribeTableOutput"},
	//      "errors":[
	//        {"shape":"ResourceNotFoundException"},
	//        {"shape":"InternalServerError"}
	//      ],
	//      "endpointdiscovery":{
	//      }
	//    },
	//
	// Where the ResourceNotFoundException shape looks like this:
	//
	//    "ResourceNotFoundException":{
	//      "type":"structure",
	//      "members":{
	//        "message":{"shape":"ErrorMessage"}
	//      },
	//      "exception":true
	//    },
	//
	//
	// Which does not indicate that the error is a 404 :( So, the logic in the
	// CRD.ExceptionCode(404) method needs to get its override from the
	// generate.yaml configuration file.
	assert.Equal("ResourceNotFoundException", crd.ExceptionCode(404))

	specFields := crd.SpecFields

	expSpecFieldCamel := []string{
		"AttributeDefinitions",
		"BillingMode",
		"GlobalSecondaryIndexes",
		"KeySchema",
		"LocalSecondaryIndexes",
		"ProvisionedThroughput",
		"SSESpecification",
		"StreamSpecification",
		"TableName",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// The DynamoDB API uses an API that uses "wrapper" single-member objects
	// in the JSON response for the create/describe calls. In other words, the
	// returned result from the DescribeTable API looks like this:
	//
	// {
	//   "table": {
	//	 .. bunch of fields for the table ..
	//   }
	// }
	//
	// However, the *ShapeName* of the "table" field is actually
	// TableDescription. This tests that we're properly outputting the
	// memberName (which is "Table" and not "TableDescription") when we build
	// the Table CRD's Status field from the DescribeTableOutput shape.
	expReadOneOutput := `
	if resp.Table.ArchivalSummary != nil {
		f0 := &svcapitypes.ArchivalSummary{}
		if resp.Table.ArchivalSummary.ArchivalBackupArn != nil {
			f0.ArchivalBackupARN = resp.Table.ArchivalSummary.ArchivalBackupArn
		}
		if resp.Table.ArchivalSummary.ArchivalDateTime != nil {
			f0.ArchivalDateTime = &metav1.Time{*resp.Table.ArchivalSummary.ArchivalDateTime}
		}
		if resp.Table.ArchivalSummary.ArchivalReason != nil {
			f0.ArchivalReason = resp.Table.ArchivalSummary.ArchivalReason
		}
		ko.Status.ArchivalSummary = f0
	}
	if resp.Table.AttributeDefinitions != nil {
		f1 := []*svcapitypes.AttributeDefinition{}
		for _, f1iter := range resp.Table.AttributeDefinitions {
			f1elem := &svcapitypes.AttributeDefinition{}
			if f1iter.AttributeName != nil {
				f1elem.AttributeName = f1iter.AttributeName
			}
			if f1iter.AttributeType != nil {
				f1elem.AttributeType = f1iter.AttributeType
			}
			f1 = append(f1, f1elem)
		}
		ko.Spec.AttributeDefinitions = f1
	}
	if resp.Table.BillingModeSummary != nil {
		f2 := &svcapitypes.BillingModeSummary{}
		if resp.Table.BillingModeSummary.BillingMode != nil {
			f2.BillingMode = resp.Table.BillingModeSummary.BillingMode
		}
		if resp.Table.BillingModeSummary.LastUpdateToPayPerRequestDateTime != nil {
			f2.LastUpdateToPayPerRequestDateTime = &metav1.Time{*resp.Table.BillingModeSummary.LastUpdateToPayPerRequestDateTime}
		}
		ko.Status.BillingModeSummary = f2
	}
	if resp.Table.CreationDateTime != nil {
		ko.Status.CreationDateTime = &metav1.Time{*resp.Table.CreationDateTime}
	}
	if resp.Table.GlobalSecondaryIndexes != nil {
		f4 := []*svcapitypes.GlobalSecondaryIndex{}
		for _, f4iter := range resp.Table.GlobalSecondaryIndexes {
			f4elem := &svcapitypes.GlobalSecondaryIndex{}
			if f4iter.IndexName != nil {
				f4elem.IndexName = f4iter.IndexName
			}
			if f4iter.KeySchema != nil {
				f4elemf6 := []*svcapitypes.KeySchemaElement{}
				for _, f4elemf6iter := range f4iter.KeySchema {
					f4elemf6elem := &svcapitypes.KeySchemaElement{}
					if f4elemf6iter.AttributeName != nil {
						f4elemf6elem.AttributeName = f4elemf6iter.AttributeName
					}
					if f4elemf6iter.KeyType != nil {
						f4elemf6elem.KeyType = f4elemf6iter.KeyType
					}
					f4elemf6 = append(f4elemf6, f4elemf6elem)
				}
				f4elem.KeySchema = f4elemf6
			}
			if f4iter.Projection != nil {
				f4elemf7 := &svcapitypes.Projection{}
				if f4iter.Projection.NonKeyAttributes != nil {
					f4elemf7f0 := []*string{}
					for _, f4elemf7f0iter := range f4iter.Projection.NonKeyAttributes {
						var f4elemf7f0elem string
						f4elemf7f0elem = *f4elemf7f0iter
						f4elemf7f0 = append(f4elemf7f0, &f4elemf7f0elem)
					}
					f4elemf7.NonKeyAttributes = f4elemf7f0
				}
				if f4iter.Projection.ProjectionType != nil {
					f4elemf7.ProjectionType = f4iter.Projection.ProjectionType
				}
				f4elem.Projection = f4elemf7
			}
			if f4iter.ProvisionedThroughput != nil {
				f4elemf8 := &svcapitypes.ProvisionedThroughput{}
				if f4iter.ProvisionedThroughput.ReadCapacityUnits != nil {
					f4elemf8.ReadCapacityUnits = f4iter.ProvisionedThroughput.ReadCapacityUnits
				}
				if f4iter.ProvisionedThroughput.WriteCapacityUnits != nil {
					f4elemf8.WriteCapacityUnits = f4iter.ProvisionedThroughput.WriteCapacityUnits
				}
				f4elem.ProvisionedThroughput = f4elemf8
			}
			f4 = append(f4, f4elem)
		}
		ko.Spec.GlobalSecondaryIndexes = f4
	}
	if resp.Table.GlobalTableVersion != nil {
		ko.Status.GlobalTableVersion = resp.Table.GlobalTableVersion
	}
	if resp.Table.ItemCount != nil {
		ko.Status.ItemCount = resp.Table.ItemCount
	}
	if resp.Table.KeySchema != nil {
		f7 := []*svcapitypes.KeySchemaElement{}
		for _, f7iter := range resp.Table.KeySchema {
			f7elem := &svcapitypes.KeySchemaElement{}
			if f7iter.AttributeName != nil {
				f7elem.AttributeName = f7iter.AttributeName
			}
			if f7iter.KeyType != nil {
				f7elem.KeyType = f7iter.KeyType
			}
			f7 = append(f7, f7elem)
		}
		ko.Spec.KeySchema = f7
	}
	if resp.Table.LatestStreamArn != nil {
		ko.Status.LatestStreamARN = resp.Table.LatestStreamArn
	}
	if resp.Table.LatestStreamLabel != nil {
		ko.Status.LatestStreamLabel = resp.Table.LatestStreamLabel
	}
	if resp.Table.LocalSecondaryIndexes != nil {
		f10 := []*svcapitypes.LocalSecondaryIndex{}
		for _, f10iter := range resp.Table.LocalSecondaryIndexes {
			f10elem := &svcapitypes.LocalSecondaryIndex{}
			if f10iter.IndexName != nil {
				f10elem.IndexName = f10iter.IndexName
			}
			if f10iter.KeySchema != nil {
				f10elemf4 := []*svcapitypes.KeySchemaElement{}
				for _, f10elemf4iter := range f10iter.KeySchema {
					f10elemf4elem := &svcapitypes.KeySchemaElement{}
					if f10elemf4iter.AttributeName != nil {
						f10elemf4elem.AttributeName = f10elemf4iter.AttributeName
					}
					if f10elemf4iter.KeyType != nil {
						f10elemf4elem.KeyType = f10elemf4iter.KeyType
					}
					f10elemf4 = append(f10elemf4, f10elemf4elem)
				}
				f10elem.KeySchema = f10elemf4
			}
			if f10iter.Projection != nil {
				f10elemf5 := &svcapitypes.Projection{}
				if f10iter.Projection.NonKeyAttributes != nil {
					f10elemf5f0 := []*string{}
					for _, f10elemf5f0iter := range f10iter.Projection.NonKeyAttributes {
						var f10elemf5f0elem string
						f10elemf5f0elem = *f10elemf5f0iter
						f10elemf5f0 = append(f10elemf5f0, &f10elemf5f0elem)
					}
					f10elemf5.NonKeyAttributes = f10elemf5f0
				}
				if f10iter.Projection.ProjectionType != nil {
					f10elemf5.ProjectionType = f10iter.Projection.ProjectionType
				}
				f10elem.Projection = f10elemf5
			}
			f10 = append(f10, f10elem)
		}
		ko.Spec.LocalSecondaryIndexes = f10
	}
	if resp.Table.ProvisionedThroughput != nil {
		f11 := &svcapitypes.ProvisionedThroughput{}
		if resp.Table.ProvisionedThroughput.ReadCapacityUnits != nil {
			f11.ReadCapacityUnits = resp.Table.ProvisionedThroughput.ReadCapacityUnits
		}
		if resp.Table.ProvisionedThroughput.WriteCapacityUnits != nil {
			f11.WriteCapacityUnits = resp.Table.ProvisionedThroughput.WriteCapacityUnits
		}
		ko.Spec.ProvisionedThroughput = f11
	}
	if resp.Table.Replicas != nil {
		f12 := []*svcapitypes.ReplicaDescription{}
		for _, f12iter := range resp.Table.Replicas {
			f12elem := &svcapitypes.ReplicaDescription{}
			if f12iter.GlobalSecondaryIndexes != nil {
				f12elemf0 := []*svcapitypes.ReplicaGlobalSecondaryIndexDescription{}
				for _, f12elemf0iter := range f12iter.GlobalSecondaryIndexes {
					f12elemf0elem := &svcapitypes.ReplicaGlobalSecondaryIndexDescription{}
					if f12elemf0iter.IndexName != nil {
						f12elemf0elem.IndexName = f12elemf0iter.IndexName
					}
					if f12elemf0iter.ProvisionedThroughputOverride != nil {
						f12elemf0elemf1 := &svcapitypes.ProvisionedThroughputOverride{}
						if f12elemf0iter.ProvisionedThroughputOverride.ReadCapacityUnits != nil {
							f12elemf0elemf1.ReadCapacityUnits = f12elemf0iter.ProvisionedThroughputOverride.ReadCapacityUnits
						}
						f12elemf0elem.ProvisionedThroughputOverride = f12elemf0elemf1
					}
					f12elemf0 = append(f12elemf0, f12elemf0elem)
				}
				f12elem.GlobalSecondaryIndexes = f12elemf0
			}
			if f12iter.KMSMasterKeyId != nil {
				f12elem.KMSMasterKeyID = f12iter.KMSMasterKeyId
			}
			if f12iter.ProvisionedThroughputOverride != nil {
				f12elemf2 := &svcapitypes.ProvisionedThroughputOverride{}
				if f12iter.ProvisionedThroughputOverride.ReadCapacityUnits != nil {
					f12elemf2.ReadCapacityUnits = f12iter.ProvisionedThroughputOverride.ReadCapacityUnits
				}
				f12elem.ProvisionedThroughputOverride = f12elemf2
			}
			if f12iter.RegionName != nil {
				f12elem.RegionName = f12iter.RegionName
			}
			if f12iter.ReplicaStatus != nil {
				f12elem.ReplicaStatus = f12iter.ReplicaStatus
			}
			if f12iter.ReplicaStatusDescription != nil {
				f12elem.ReplicaStatusDescription = f12iter.ReplicaStatusDescription
			}
			if f12iter.ReplicaStatusPercentProgress != nil {
				f12elem.ReplicaStatusPercentProgress = f12iter.ReplicaStatusPercentProgress
			}
			f12 = append(f12, f12elem)
		}
		ko.Status.Replicas = f12
	}
	if resp.Table.RestoreSummary != nil {
		f13 := &svcapitypes.RestoreSummary{}
		if resp.Table.RestoreSummary.RestoreDateTime != nil {
			f13.RestoreDateTime = &metav1.Time{*resp.Table.RestoreSummary.RestoreDateTime}
		}
		if resp.Table.RestoreSummary.RestoreInProgress != nil {
			f13.RestoreInProgress = resp.Table.RestoreSummary.RestoreInProgress
		}
		if resp.Table.RestoreSummary.SourceBackupArn != nil {
			f13.SourceBackupARN = resp.Table.RestoreSummary.SourceBackupArn
		}
		if resp.Table.RestoreSummary.SourceTableArn != nil {
			f13.SourceTableARN = resp.Table.RestoreSummary.SourceTableArn
		}
		ko.Status.RestoreSummary = f13
	}
	if resp.Table.SSEDescription != nil {
		f14 := &svcapitypes.SSEDescription{}
		if resp.Table.SSEDescription.InaccessibleEncryptionDateTime != nil {
			f14.InaccessibleEncryptionDateTime = &metav1.Time{*resp.Table.SSEDescription.InaccessibleEncryptionDateTime}
		}
		if resp.Table.SSEDescription.KMSMasterKeyArn != nil {
			f14.KMSMasterKeyARN = resp.Table.SSEDescription.KMSMasterKeyArn
		}
		if resp.Table.SSEDescription.SSEType != nil {
			f14.SSEType = resp.Table.SSEDescription.SSEType
		}
		if resp.Table.SSEDescription.Status != nil {
			f14.Status = resp.Table.SSEDescription.Status
		}
		ko.Status.SSEDescription = f14
	}
	if resp.Table.StreamSpecification != nil {
		f15 := &svcapitypes.StreamSpecification{}
		if resp.Table.StreamSpecification.StreamEnabled != nil {
			f15.StreamEnabled = resp.Table.StreamSpecification.StreamEnabled
		}
		if resp.Table.StreamSpecification.StreamViewType != nil {
			f15.StreamViewType = resp.Table.StreamSpecification.StreamViewType
		}
		ko.Spec.StreamSpecification = f15
	}
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.Table.TableArn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.Table.TableArn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Table.TableId != nil {
		ko.Status.TableID = resp.Table.TableId
	}
	if resp.Table.TableName != nil {
		ko.Spec.TableName = resp.Table.TableName
	}
	if resp.Table.TableSizeBytes != nil {
		ko.Status.TableSizeBytes = resp.Table.TableSizeBytes
	}
	if resp.Table.TableStatus != nil {
		ko.Status.TableStatus = resp.Table.TableStatus
	}
`
	assert.Equal(expReadOneOutput, crd.GoCodeSetOutput(model.OpTypeGet, "resp", "ko", 1, true))
}
