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

	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestRDS_DBInstance(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "rds")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("DBInstance", crds)
	require.NotNil(crd)

	assert.Equal("DBInstance", crd.Names.Camel)
	assert.Equal("dbInstance", crd.Names.CamelLower)
	assert.Equal("db_instance", crd.Names.Snake)

	// The RDS DBInstance API has the following operations:
	// - CreateDBInstance
	// - DescribeDBInstances
	// - ModifyDBInstance
	// - DeleteDBInstance
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.Update)

	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)

	// The DescribeDBInstances operation has the following definition:
	//
	//    "DescribeDBInstances":{
	//      "name":"DescribeDBInstances",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"DescribeDBInstancesMessage"},
	//      "output":{
	//        "shape":"DBInstanceMessage",
	//        "resultWrapper":"DescribeDBInstancesResult"
	//      },
	//      "errors":[
	//        {"shape":"DBInstanceNotFoundFault"}
	//      ]
	//    },
	//
	// NOTE: This is UNUSUAL for List operation to return a 404 Not Found
	// instead of a 200 OK with an empty array of results.
	//
	// Where the DBInstanceNotFoundFault shape looks like this:
	//
	//    "DBInstanceNotFoundFault":{
	//      "type":"structure",
	//      "members":{
	//      },
	//      "error":{
	//        "code":"DBInstanceNotFound",
	//        "httpStatusCode":404,
	//        "senderFault":true
	//      },
	//      "exception":true
	//    },
	//
	// Which clearly indicates it is the 404 HTTP fault for this resource even
	// though the "code" is "DBInstanceNotFound"
	assert.Equal("DBInstanceNotFound", crd.ExceptionCode(404))

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"AllocatedStorage",
		"AutoMinorVersionUpgrade",
		"AvailabilityZone",
		"BackupRetentionPeriod",
		"CharacterSetName",
		"CopyTagsToSnapshot",
		"DBClusterIdentifier",
		"DBInstanceClass",
		"DBInstanceIdentifier",
		"DBName",
		"DBParameterGroupName",
		"DBSecurityGroups",
		"DBSubnetGroupName",
		"DeletionProtection",
		"Domain",
		"DomainIAMRoleName",
		"EnableCloudwatchLogsExports",
		"EnableIAMDatabaseAuthentication",
		"EnablePerformanceInsights",
		"Engine",
		"EngineVersion",
		"IOPS",
		"KMSKeyID",
		"LicenseModel",
		"MasterUserPassword",
		"MasterUsername",
		"MaxAllocatedStorage",
		"MonitoringInterval",
		"MonitoringRoleARN",
		"MultiAZ",
		"OptionGroupName",
		"PerformanceInsightsKMSKeyID",
		"PerformanceInsightsRetentionPeriod",
		"Port",
		"PreferredBackupWindow",
		"PreferredMaintenanceWindow",
		"ProcessorFeatures",
		"PromotionTier",
		"PubliclyAccessible",
		"StorageEncrypted",
		"StorageType",
		"TDECredentialARN",
		"TDECredentialPassword",
		"Tags",
		"Timezone",
		"VPCSecurityGroupIDs",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"AssociatedRoles",
		"CACertificateIdentifier",
		"DBIResourceID",
		"DBInstancePort",
		"DBInstanceStatus",
		"DBParameterGroups",
		"DBSubnetGroup",
		"DomainMemberships",
		"EnabledCloudwatchLogsExports",
		"Endpoint",
		"EnhancedMonitoringResourceARN",
		"IAMDatabaseAuthenticationEnabled",
		"InstanceCreateTime",
		"LatestRestorableTime",
		"ListenerEndpoint",
		"OptionGroupMemberships",
		"PendingModifiedValues",
		"PerformanceInsightsEnabled",
		"ReadReplicaDBClusterIdentifiers",
		"ReadReplicaDBInstanceIdentifiers",
		"ReadReplicaSourceDBInstanceIdentifier",
		"SecondaryAvailabilityZone",
		"StatusInfos",
		"VPCSecurityGroups",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))
}
