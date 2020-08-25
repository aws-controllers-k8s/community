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

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func TestRDS_DBInstance(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "rds")

	crds, err := sh.GetCRDs()
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

	expCreateInput := `
	if r.ko.Spec.AllocatedStorage != nil {
		res.SetAllocatedStorage(*r.ko.Spec.AllocatedStorage)
	}
	if r.ko.Spec.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*r.ko.Spec.AutoMinorVersionUpgrade)
	}
	if r.ko.Spec.AvailabilityZone != nil {
		res.SetAvailabilityZone(*r.ko.Spec.AvailabilityZone)
	}
	if r.ko.Spec.BackupRetentionPeriod != nil {
		res.SetBackupRetentionPeriod(*r.ko.Spec.BackupRetentionPeriod)
	}
	if r.ko.Spec.CharacterSetName != nil {
		res.SetCharacterSetName(*r.ko.Spec.CharacterSetName)
	}
	if r.ko.Spec.CopyTagsToSnapshot != nil {
		res.SetCopyTagsToSnapshot(*r.ko.Spec.CopyTagsToSnapshot)
	}
	if r.ko.Spec.DBClusterIdentifier != nil {
		res.SetDBClusterIdentifier(*r.ko.Spec.DBClusterIdentifier)
	}
	if r.ko.Spec.DBInstanceClass != nil {
		res.SetDBInstanceClass(*r.ko.Spec.DBInstanceClass)
	}
	if r.ko.Spec.DBInstanceIdentifier != nil {
		res.SetDBInstanceIdentifier(*r.ko.Spec.DBInstanceIdentifier)
	}
	if r.ko.Spec.DBName != nil {
		res.SetDBName(*r.ko.Spec.DBName)
	}
	if r.ko.Spec.DBParameterGroupName != nil {
		res.SetDBParameterGroupName(*r.ko.Spec.DBParameterGroupName)
	}
	if r.ko.Spec.DBSecurityGroups != nil {
		f11 := []*string{}
		for _, f11iter := range r.ko.Spec.DBSecurityGroups {
			var f11elem string
			f11elem = *f11iter
			f11 = append(f11, &f11elem)
		}
		res.SetDBSecurityGroups(f11)
	}
	if r.ko.Spec.DBSubnetGroupName != nil {
		res.SetDBSubnetGroupName(*r.ko.Spec.DBSubnetGroupName)
	}
	if r.ko.Spec.DeletionProtection != nil {
		res.SetDeletionProtection(*r.ko.Spec.DeletionProtection)
	}
	if r.ko.Spec.Domain != nil {
		res.SetDomain(*r.ko.Spec.Domain)
	}
	if r.ko.Spec.DomainIAMRoleName != nil {
		res.SetDomainIAMRoleName(*r.ko.Spec.DomainIAMRoleName)
	}
	if r.ko.Spec.EnableCloudwatchLogsExports != nil {
		f16 := []*string{}
		for _, f16iter := range r.ko.Spec.EnableCloudwatchLogsExports {
			var f16elem string
			f16elem = *f16iter
			f16 = append(f16, &f16elem)
		}
		res.SetEnableCloudwatchLogsExports(f16)
	}
	if r.ko.Spec.EnableIAMDatabaseAuthentication != nil {
		res.SetEnableIAMDatabaseAuthentication(*r.ko.Spec.EnableIAMDatabaseAuthentication)
	}
	if r.ko.Spec.EnablePerformanceInsights != nil {
		res.SetEnablePerformanceInsights(*r.ko.Spec.EnablePerformanceInsights)
	}
	if r.ko.Spec.Engine != nil {
		res.SetEngine(*r.ko.Spec.Engine)
	}
	if r.ko.Spec.EngineVersion != nil {
		res.SetEngineVersion(*r.ko.Spec.EngineVersion)
	}
	if r.ko.Spec.IOPS != nil {
		res.SetIops(*r.ko.Spec.IOPS)
	}
	if r.ko.Spec.KMSKeyID != nil {
		res.SetKmsKeyId(*r.ko.Spec.KMSKeyID)
	}
	if r.ko.Spec.LicenseModel != nil {
		res.SetLicenseModel(*r.ko.Spec.LicenseModel)
	}
	if r.ko.Spec.MasterUserPassword != nil {
		res.SetMasterUserPassword(*r.ko.Spec.MasterUserPassword)
	}
	if r.ko.Spec.MasterUsername != nil {
		res.SetMasterUsername(*r.ko.Spec.MasterUsername)
	}
	if r.ko.Spec.MaxAllocatedStorage != nil {
		res.SetMaxAllocatedStorage(*r.ko.Spec.MaxAllocatedStorage)
	}
	if r.ko.Spec.MonitoringInterval != nil {
		res.SetMonitoringInterval(*r.ko.Spec.MonitoringInterval)
	}
	if r.ko.Spec.MonitoringRoleARN != nil {
		res.SetMonitoringRoleArn(*r.ko.Spec.MonitoringRoleARN)
	}
	if r.ko.Spec.MultiAZ != nil {
		res.SetMultiAZ(*r.ko.Spec.MultiAZ)
	}
	if r.ko.Spec.OptionGroupName != nil {
		res.SetOptionGroupName(*r.ko.Spec.OptionGroupName)
	}
	if r.ko.Spec.PerformanceInsightsKMSKeyID != nil {
		res.SetPerformanceInsightsKMSKeyId(*r.ko.Spec.PerformanceInsightsKMSKeyID)
	}
	if r.ko.Spec.PerformanceInsightsRetentionPeriod != nil {
		res.SetPerformanceInsightsRetentionPeriod(*r.ko.Spec.PerformanceInsightsRetentionPeriod)
	}
	if r.ko.Spec.Port != nil {
		res.SetPort(*r.ko.Spec.Port)
	}
	if r.ko.Spec.PreferredBackupWindow != nil {
		res.SetPreferredBackupWindow(*r.ko.Spec.PreferredBackupWindow)
	}
	if r.ko.Spec.PreferredMaintenanceWindow != nil {
		res.SetPreferredMaintenanceWindow(*r.ko.Spec.PreferredMaintenanceWindow)
	}
	if r.ko.Spec.ProcessorFeatures != nil {
		f36 := []*svcsdk.ProcessorFeature{}
		for _, f36iter := range r.ko.Spec.ProcessorFeatures {
			f36elem := &svcsdk.ProcessorFeature{}
			if f36iter.Name != nil {
				f36elem.SetName(*f36iter.Name)
			}
			if f36iter.Value != nil {
				f36elem.SetValue(*f36iter.Value)
			}
			f36 = append(f36, f36elem)
		}
		res.SetProcessorFeatures(f36)
	}
	if r.ko.Spec.PromotionTier != nil {
		res.SetPromotionTier(*r.ko.Spec.PromotionTier)
	}
	if r.ko.Spec.PubliclyAccessible != nil {
		res.SetPubliclyAccessible(*r.ko.Spec.PubliclyAccessible)
	}
	if r.ko.Spec.StorageEncrypted != nil {
		res.SetStorageEncrypted(*r.ko.Spec.StorageEncrypted)
	}
	if r.ko.Spec.StorageType != nil {
		res.SetStorageType(*r.ko.Spec.StorageType)
	}
	if r.ko.Spec.Tags != nil {
		f41 := []*svcsdk.Tag{}
		for _, f41iter := range r.ko.Spec.Tags {
			f41elem := &svcsdk.Tag{}
			if f41iter.Key != nil {
				f41elem.SetKey(*f41iter.Key)
			}
			if f41iter.Value != nil {
				f41elem.SetValue(*f41iter.Value)
			}
			f41 = append(f41, f41elem)
		}
		res.SetTags(f41)
	}
	if r.ko.Spec.TDECredentialARN != nil {
		res.SetTdeCredentialArn(*r.ko.Spec.TDECredentialARN)
	}
	if r.ko.Spec.TDECredentialPassword != nil {
		res.SetTdeCredentialPassword(*r.ko.Spec.TDECredentialPassword)
	}
	if r.ko.Spec.Timezone != nil {
		res.SetTimezone(*r.ko.Spec.Timezone)
	}
	if r.ko.Spec.VPCSecurityGroupIDs != nil {
		f45 := []*string{}
		for _, f45iter := range r.ko.Spec.VPCSecurityGroupIDs {
			var f45elem string
			f45elem = *f45iter
			f45 = append(f45, &f45elem)
		}
		res.SetVpcSecurityGroupIds(f45)
	}
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko", "res", 1))

	expCreateOutput := `
	if resp.DBInstance.AssociatedRoles != nil {
		f1 := []*svcapitypes.DBInstanceRole{}
		for _, f1iter := range resp.DBInstance.AssociatedRoles {
			f1elem := &svcapitypes.DBInstanceRole{}
			if f1iter.FeatureName != nil {
				f1elem.FeatureName = f1iter.FeatureName
			}
			if f1iter.RoleArn != nil {
				f1elem.RoleARN = f1iter.RoleArn
			}
			if f1iter.Status != nil {
				f1elem.Status = f1iter.Status
			}
			f1 = append(f1, f1elem)
		}
		ko.Status.AssociatedRoles = f1
	}
	if resp.DBInstance.CACertificateIdentifier != nil {
		ko.Status.CACertificateIdentifier = resp.DBInstance.CACertificateIdentifier
	}
	if resp.DBInstance.DBInstanceStatus != nil {
		ko.Status.DBInstanceStatus = resp.DBInstance.DBInstanceStatus
	}
	if resp.DBInstance.DBParameterGroups != nil {
		f14 := []*svcapitypes.DBParameterGroupStatus_SDK{}
		for _, f14iter := range resp.DBInstance.DBParameterGroups {
			f14elem := &svcapitypes.DBParameterGroupStatus_SDK{}
			if f14iter.DBParameterGroupName != nil {
				f14elem.DBParameterGroupName = f14iter.DBParameterGroupName
			}
			if f14iter.ParameterApplyStatus != nil {
				f14elem.ParameterApplyStatus = f14iter.ParameterApplyStatus
			}
			f14 = append(f14, f14elem)
		}
		ko.Status.DBParameterGroups = f14
	}
	if resp.DBInstance.DBSubnetGroup != nil {
		f16 := &svcapitypes.DBSubnetGroup_SDK{}
		if resp.DBInstance.DBSubnetGroup.DBSubnetGroupArn != nil {
			f16.DBSubnetGroupARN = resp.DBInstance.DBSubnetGroup.DBSubnetGroupArn
		}
		if resp.DBInstance.DBSubnetGroup.DBSubnetGroupDescription != nil {
			f16.DBSubnetGroupDescription = resp.DBInstance.DBSubnetGroup.DBSubnetGroupDescription
		}
		if resp.DBInstance.DBSubnetGroup.DBSubnetGroupName != nil {
			f16.DBSubnetGroupName = resp.DBInstance.DBSubnetGroup.DBSubnetGroupName
		}
		if resp.DBInstance.DBSubnetGroup.SubnetGroupStatus != nil {
			f16.SubnetGroupStatus = resp.DBInstance.DBSubnetGroup.SubnetGroupStatus
		}
		if resp.DBInstance.DBSubnetGroup.Subnets != nil {
			f16f4 := []*svcapitypes.Subnet{}
			for _, f16f4iter := range resp.DBInstance.DBSubnetGroup.Subnets {
				f16f4elem := &svcapitypes.Subnet{}
				if f16f4iter.SubnetAvailabilityZone != nil {
					f16f4elemf0 := &svcapitypes.AvailabilityZone{}
					if f16f4iter.SubnetAvailabilityZone.Name != nil {
						f16f4elemf0.Name = f16f4iter.SubnetAvailabilityZone.Name
					}
					f16f4elem.SubnetAvailabilityZone = f16f4elemf0
				}
				if f16f4iter.SubnetIdentifier != nil {
					f16f4elem.SubnetIdentifier = f16f4iter.SubnetIdentifier
				}
				if f16f4iter.SubnetOutpost != nil {
					f16f4elemf2 := &svcapitypes.Outpost{}
					if f16f4iter.SubnetOutpost.Arn != nil {
						f16f4elemf2.ARN = f16f4iter.SubnetOutpost.Arn
					}
					f16f4elem.SubnetOutpost = f16f4elemf2
				}
				if f16f4iter.SubnetStatus != nil {
					f16f4elem.SubnetStatus = f16f4iter.SubnetStatus
				}
				f16f4 = append(f16f4, f16f4elem)
			}
			f16.Subnets = f16f4
		}
		if resp.DBInstance.DBSubnetGroup.VpcId != nil {
			f16.VPCID = resp.DBInstance.DBSubnetGroup.VpcId
		}
		ko.Status.DBSubnetGroup = f16
	}
	if resp.DBInstance.DbInstancePort != nil {
		ko.Status.DBInstancePort = resp.DBInstance.DbInstancePort
	}
	if resp.DBInstance.DbiResourceId != nil {
		ko.Status.DBIResourceID = resp.DBInstance.DbiResourceId
	}
	if resp.DBInstance.DomainMemberships != nil {
		f20 := []*svcapitypes.DomainMembership{}
		for _, f20iter := range resp.DBInstance.DomainMemberships {
			f20elem := &svcapitypes.DomainMembership{}
			if f20iter.Domain != nil {
				f20elem.Domain = f20iter.Domain
			}
			if f20iter.FQDN != nil {
				f20elem.FQDN = f20iter.FQDN
			}
			if f20iter.IAMRoleName != nil {
				f20elem.IAMRoleName = f20iter.IAMRoleName
			}
			if f20iter.Status != nil {
				f20elem.Status = f20iter.Status
			}
			f20 = append(f20, f20elem)
		}
		ko.Status.DomainMemberships = f20
	}
	if resp.DBInstance.EnabledCloudwatchLogsExports != nil {
		f21 := []*string{}
		for _, f21iter := range resp.DBInstance.EnabledCloudwatchLogsExports {
			var f21elem string
			f21elem = *f21iter
			f21 = append(f21, &f21elem)
		}
		ko.Status.EnabledCloudwatchLogsExports = f21
	}
	if resp.DBInstance.Endpoint != nil {
		f22 := &svcapitypes.Endpoint{}
		if resp.DBInstance.Endpoint.Address != nil {
			f22.Address = resp.DBInstance.Endpoint.Address
		}
		if resp.DBInstance.Endpoint.HostedZoneId != nil {
			f22.HostedZoneID = resp.DBInstance.Endpoint.HostedZoneId
		}
		if resp.DBInstance.Endpoint.Port != nil {
			f22.Port = resp.DBInstance.Endpoint.Port
		}
		ko.Status.Endpoint = f22
	}
	if resp.DBInstance.EnhancedMonitoringResourceArn != nil {
		ko.Status.EnhancedMonitoringResourceARN = resp.DBInstance.EnhancedMonitoringResourceArn
	}
	if resp.DBInstance.IAMDatabaseAuthenticationEnabled != nil {
		ko.Status.IAMDatabaseAuthenticationEnabled = resp.DBInstance.IAMDatabaseAuthenticationEnabled
	}
	if resp.DBInstance.InstanceCreateTime != nil {
		ko.Status.InstanceCreateTime = &metav1.Time{*resp.DBInstance.InstanceCreateTime}
	}
	if resp.DBInstance.LatestRestorableTime != nil {
		ko.Status.LatestRestorableTime = &metav1.Time{*resp.DBInstance.LatestRestorableTime}
	}
	if resp.DBInstance.ListenerEndpoint != nil {
		f32 := &svcapitypes.Endpoint{}
		if resp.DBInstance.ListenerEndpoint.Address != nil {
			f32.Address = resp.DBInstance.ListenerEndpoint.Address
		}
		if resp.DBInstance.ListenerEndpoint.HostedZoneId != nil {
			f32.HostedZoneID = resp.DBInstance.ListenerEndpoint.HostedZoneId
		}
		if resp.DBInstance.ListenerEndpoint.Port != nil {
			f32.Port = resp.DBInstance.ListenerEndpoint.Port
		}
		ko.Status.ListenerEndpoint = f32
	}
	if resp.DBInstance.OptionGroupMemberships != nil {
		f38 := []*svcapitypes.OptionGroupMembership{}
		for _, f38iter := range resp.DBInstance.OptionGroupMemberships {
			f38elem := &svcapitypes.OptionGroupMembership{}
			if f38iter.OptionGroupName != nil {
				f38elem.OptionGroupName = f38iter.OptionGroupName
			}
			if f38iter.Status != nil {
				f38elem.Status = f38iter.Status
			}
			f38 = append(f38, f38elem)
		}
		ko.Status.OptionGroupMemberships = f38
	}
	if resp.DBInstance.PendingModifiedValues != nil {
		f39 := &svcapitypes.PendingModifiedValues{}
		if resp.DBInstance.PendingModifiedValues.AllocatedStorage != nil {
			f39.AllocatedStorage = resp.DBInstance.PendingModifiedValues.AllocatedStorage
		}
		if resp.DBInstance.PendingModifiedValues.BackupRetentionPeriod != nil {
			f39.BackupRetentionPeriod = resp.DBInstance.PendingModifiedValues.BackupRetentionPeriod
		}
		if resp.DBInstance.PendingModifiedValues.CACertificateIdentifier != nil {
			f39.CACertificateIdentifier = resp.DBInstance.PendingModifiedValues.CACertificateIdentifier
		}
		if resp.DBInstance.PendingModifiedValues.DBInstanceClass != nil {
			f39.DBInstanceClass = resp.DBInstance.PendingModifiedValues.DBInstanceClass
		}
		if resp.DBInstance.PendingModifiedValues.DBInstanceIdentifier != nil {
			f39.DBInstanceIdentifier = resp.DBInstance.PendingModifiedValues.DBInstanceIdentifier
		}
		if resp.DBInstance.PendingModifiedValues.DBSubnetGroupName != nil {
			f39.DBSubnetGroupName = resp.DBInstance.PendingModifiedValues.DBSubnetGroupName
		}
		if resp.DBInstance.PendingModifiedValues.EngineVersion != nil {
			f39.EngineVersion = resp.DBInstance.PendingModifiedValues.EngineVersion
		}
		if resp.DBInstance.PendingModifiedValues.Iops != nil {
			f39.IOPS = resp.DBInstance.PendingModifiedValues.Iops
		}
		if resp.DBInstance.PendingModifiedValues.LicenseModel != nil {
			f39.LicenseModel = resp.DBInstance.PendingModifiedValues.LicenseModel
		}
		if resp.DBInstance.PendingModifiedValues.MasterUserPassword != nil {
			f39.MasterUserPassword = resp.DBInstance.PendingModifiedValues.MasterUserPassword
		}
		if resp.DBInstance.PendingModifiedValues.MultiAZ != nil {
			f39.MultiAZ = resp.DBInstance.PendingModifiedValues.MultiAZ
		}
		if resp.DBInstance.PendingModifiedValues.PendingCloudwatchLogsExports != nil {
			f39f11 := &svcapitypes.PendingCloudwatchLogsExports{}
			if resp.DBInstance.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable != nil {
				f39f11f0 := []*string{}
				for _, f39f11f0iter := range resp.DBInstance.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable {
					var f39f11f0elem string
					f39f11f0elem = *f39f11f0iter
					f39f11f0 = append(f39f11f0, &f39f11f0elem)
				}
				f39f11.LogTypesToDisable = f39f11f0
			}
			if resp.DBInstance.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable != nil {
				f39f11f1 := []*string{}
				for _, f39f11f1iter := range resp.DBInstance.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable {
					var f39f11f1elem string
					f39f11f1elem = *f39f11f1iter
					f39f11f1 = append(f39f11f1, &f39f11f1elem)
				}
				f39f11.LogTypesToEnable = f39f11f1
			}
			f39.PendingCloudwatchLogsExports = f39f11
		}
		if resp.DBInstance.PendingModifiedValues.Port != nil {
			f39.Port = resp.DBInstance.PendingModifiedValues.Port
		}
		if resp.DBInstance.PendingModifiedValues.ProcessorFeatures != nil {
			f39f13 := []*svcapitypes.ProcessorFeature{}
			for _, f39f13iter := range resp.DBInstance.PendingModifiedValues.ProcessorFeatures {
				f39f13elem := &svcapitypes.ProcessorFeature{}
				if f39f13iter.Name != nil {
					f39f13elem.Name = f39f13iter.Name
				}
				if f39f13iter.Value != nil {
					f39f13elem.Value = f39f13iter.Value
				}
				f39f13 = append(f39f13, f39f13elem)
			}
			f39.ProcessorFeatures = f39f13
		}
		if resp.DBInstance.PendingModifiedValues.StorageType != nil {
			f39.StorageType = resp.DBInstance.PendingModifiedValues.StorageType
		}
		ko.Status.PendingModifiedValues = f39
	}
	if resp.DBInstance.PerformanceInsightsEnabled != nil {
		ko.Status.PerformanceInsightsEnabled = resp.DBInstance.PerformanceInsightsEnabled
	}
	if resp.DBInstance.ReadReplicaDBClusterIdentifiers != nil {
		f48 := []*string{}
		for _, f48iter := range resp.DBInstance.ReadReplicaDBClusterIdentifiers {
			var f48elem string
			f48elem = *f48iter
			f48 = append(f48, &f48elem)
		}
		ko.Status.ReadReplicaDBClusterIdentifiers = f48
	}
	if resp.DBInstance.ReadReplicaDBInstanceIdentifiers != nil {
		f49 := []*string{}
		for _, f49iter := range resp.DBInstance.ReadReplicaDBInstanceIdentifiers {
			var f49elem string
			f49elem = *f49iter
			f49 = append(f49, &f49elem)
		}
		ko.Status.ReadReplicaDBInstanceIdentifiers = f49
	}
	if resp.DBInstance.ReadReplicaSourceDBInstanceIdentifier != nil {
		ko.Status.ReadReplicaSourceDBInstanceIdentifier = resp.DBInstance.ReadReplicaSourceDBInstanceIdentifier
	}
	if resp.DBInstance.SecondaryAvailabilityZone != nil {
		ko.Status.SecondaryAvailabilityZone = resp.DBInstance.SecondaryAvailabilityZone
	}
	if resp.DBInstance.StatusInfos != nil {
		f52 := []*svcapitypes.DBInstanceStatusInfo{}
		for _, f52iter := range resp.DBInstance.StatusInfos {
			f52elem := &svcapitypes.DBInstanceStatusInfo{}
			if f52iter.Message != nil {
				f52elem.Message = f52iter.Message
			}
			if f52iter.Normal != nil {
				f52elem.Normal = f52iter.Normal
			}
			if f52iter.Status != nil {
				f52elem.Status = f52iter.Status
			}
			if f52iter.StatusType != nil {
				f52elem.StatusType = f52iter.StatusType
			}
			f52 = append(f52, f52elem)
		}
		ko.Status.StatusInfos = f52
	}
	if resp.DBInstance.VpcSecurityGroups != nil {
		f57 := []*svcapitypes.VPCSecurityGroupMembership{}
		for _, f57iter := range resp.DBInstance.VpcSecurityGroups {
			f57elem := &svcapitypes.VPCSecurityGroupMembership{}
			if f57iter.Status != nil {
				f57elem.Status = f57iter.Status
			}
			if f57iter.VpcSecurityGroupId != nil {
				f57elem.VPCSecurityGroupID = f57iter.VpcSecurityGroupId
			}
			f57 = append(f57, f57elem)
		}
		ko.Status.VPCSecurityGroups = f57
	}
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// This asserts that the fields of the Spec and Status structs of the
	// target variable are constructed with cleaned, renamed-friendly names
	// referring to the generated Kubernetes API type definitions
	expReadManyOutput := `
	if len(resp.DBInstances) == 0 {
		return nil, ackerr.NotFound
	}
	found := false
	for _, elem := range resp.DBInstances {
		if elem.AllocatedStorage != nil {
			ko.Spec.AllocatedStorage = elem.AllocatedStorage
		}
		if elem.AssociatedRoles != nil {
			f1 := []*svcapitypes.DBInstanceRole{}
			for _, f1iter := range elem.AssociatedRoles {
				f1elem := &svcapitypes.DBInstanceRole{}
				if f1iter.FeatureName != nil {
					f1elem.FeatureName = f1iter.FeatureName
				}
				if f1iter.RoleArn != nil {
					f1elem.RoleARN = f1iter.RoleArn
				}
				if f1iter.Status != nil {
					f1elem.Status = f1iter.Status
				}
				f1 = append(f1, f1elem)
			}
			ko.Status.AssociatedRoles = f1
		}
		if elem.AutoMinorVersionUpgrade != nil {
			ko.Spec.AutoMinorVersionUpgrade = elem.AutoMinorVersionUpgrade
		}
		if elem.AvailabilityZone != nil {
			ko.Spec.AvailabilityZone = elem.AvailabilityZone
		}
		if elem.BackupRetentionPeriod != nil {
			ko.Spec.BackupRetentionPeriod = elem.BackupRetentionPeriod
		}
		if elem.CACertificateIdentifier != nil {
			ko.Status.CACertificateIdentifier = elem.CACertificateIdentifier
		}
		if elem.CharacterSetName != nil {
			ko.Spec.CharacterSetName = elem.CharacterSetName
		}
		if elem.CopyTagsToSnapshot != nil {
			ko.Spec.CopyTagsToSnapshot = elem.CopyTagsToSnapshot
		}
		if elem.DBClusterIdentifier != nil {
			ko.Spec.DBClusterIdentifier = elem.DBClusterIdentifier
		}
		if elem.DBInstanceArn != nil {
			if ko.Status.ACKResourceMetadata == nil {
				ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			}
			tmpARN := ackv1alpha1.AWSResourceName(*elem.DBInstanceArn)
			ko.Status.ACKResourceMetadata.ARN = &tmpARN
		}
		if elem.DBInstanceClass != nil {
			ko.Spec.DBInstanceClass = elem.DBInstanceClass
		}
		if elem.DBInstanceIdentifier != nil {
			ko.Spec.DBInstanceIdentifier = elem.DBInstanceIdentifier
		}
		if elem.DBInstanceStatus != nil {
			ko.Status.DBInstanceStatus = elem.DBInstanceStatus
		}
		if elem.DBName != nil {
			ko.Spec.DBName = elem.DBName
		}
		if elem.DBParameterGroups != nil {
			f14 := []*svcapitypes.DBParameterGroupStatus_SDK{}
			for _, f14iter := range elem.DBParameterGroups {
				f14elem := &svcapitypes.DBParameterGroupStatus_SDK{}
				if f14iter.DBParameterGroupName != nil {
					f14elem.DBParameterGroupName = f14iter.DBParameterGroupName
				}
				if f14iter.ParameterApplyStatus != nil {
					f14elem.ParameterApplyStatus = f14iter.ParameterApplyStatus
				}
				f14 = append(f14, f14elem)
			}
			ko.Status.DBParameterGroups = f14
		}
		if elem.DBSecurityGroups != nil {
			f15 := []*svcapitypes.DBSecurityGroupMembership{}
			for _, f15iter := range elem.DBSecurityGroups {
				f15elem := &svcapitypes.DBSecurityGroupMembership{}
				if f15iter.DBSecurityGroupName != nil {
					f15elem.DBSecurityGroupName = f15iter.DBSecurityGroupName
				}
				if f15iter.Status != nil {
					f15elem.Status = f15iter.Status
				}
				f15 = append(f15, f15elem)
			}
			ko.Spec.DBSecurityGroups = f15
		}
		if elem.DBSubnetGroup != nil {
			f16 := &svcapitypes.DBSubnetGroup_SDK{}
			if elem.DBSubnetGroup.DBSubnetGroupArn != nil {
				f16.DBSubnetGroupARN = elem.DBSubnetGroup.DBSubnetGroupArn
			}
			if elem.DBSubnetGroup.DBSubnetGroupDescription != nil {
				f16.DBSubnetGroupDescription = elem.DBSubnetGroup.DBSubnetGroupDescription
			}
			if elem.DBSubnetGroup.DBSubnetGroupName != nil {
				f16.DBSubnetGroupName = elem.DBSubnetGroup.DBSubnetGroupName
			}
			if elem.DBSubnetGroup.SubnetGroupStatus != nil {
				f16.SubnetGroupStatus = elem.DBSubnetGroup.SubnetGroupStatus
			}
			if elem.DBSubnetGroup.Subnets != nil {
				f16f4 := []*svcapitypes.Subnet{}
				for _, f16f4iter := range elem.DBSubnetGroup.Subnets {
					f16f4elem := &svcapitypes.Subnet{}
					if f16f4iter.SubnetAvailabilityZone != nil {
						f16f4elemf0 := &svcapitypes.AvailabilityZone{}
						if f16f4iter.SubnetAvailabilityZone.Name != nil {
							f16f4elemf0.Name = f16f4iter.SubnetAvailabilityZone.Name
						}
						f16f4elem.SubnetAvailabilityZone = f16f4elemf0
					}
					if f16f4iter.SubnetIdentifier != nil {
						f16f4elem.SubnetIdentifier = f16f4iter.SubnetIdentifier
					}
					if f16f4iter.SubnetOutpost != nil {
						f16f4elemf2 := &svcapitypes.Outpost{}
						if f16f4iter.SubnetOutpost.Arn != nil {
							f16f4elemf2.ARN = f16f4iter.SubnetOutpost.Arn
						}
						f16f4elem.SubnetOutpost = f16f4elemf2
					}
					if f16f4iter.SubnetStatus != nil {
						f16f4elem.SubnetStatus = f16f4iter.SubnetStatus
					}
					f16f4 = append(f16f4, f16f4elem)
				}
				f16.Subnets = f16f4
			}
			if elem.DBSubnetGroup.VpcId != nil {
				f16.VPCID = elem.DBSubnetGroup.VpcId
			}
			ko.Status.DBSubnetGroup = f16
		}
		if elem.DbInstancePort != nil {
			ko.Status.DBInstancePort = elem.DbInstancePort
		}
		if elem.DbiResourceId != nil {
			ko.Status.DBIResourceID = elem.DbiResourceId
		}
		if elem.DeletionProtection != nil {
			ko.Spec.DeletionProtection = elem.DeletionProtection
		}
		if elem.DomainMemberships != nil {
			f20 := []*svcapitypes.DomainMembership{}
			for _, f20iter := range elem.DomainMemberships {
				f20elem := &svcapitypes.DomainMembership{}
				if f20iter.Domain != nil {
					f20elem.Domain = f20iter.Domain
				}
				if f20iter.FQDN != nil {
					f20elem.FQDN = f20iter.FQDN
				}
				if f20iter.IAMRoleName != nil {
					f20elem.IAMRoleName = f20iter.IAMRoleName
				}
				if f20iter.Status != nil {
					f20elem.Status = f20iter.Status
				}
				f20 = append(f20, f20elem)
			}
			ko.Status.DomainMemberships = f20
		}
		if elem.EnabledCloudwatchLogsExports != nil {
			f21 := []*string{}
			for _, f21iter := range elem.EnabledCloudwatchLogsExports {
				var f21elem string
				f21elem = *f21iter
				f21 = append(f21, &f21elem)
			}
			ko.Status.EnabledCloudwatchLogsExports = f21
		}
		if elem.Endpoint != nil {
			f22 := &svcapitypes.Endpoint{}
			if elem.Endpoint.Address != nil {
				f22.Address = elem.Endpoint.Address
			}
			if elem.Endpoint.HostedZoneId != nil {
				f22.HostedZoneID = elem.Endpoint.HostedZoneId
			}
			if elem.Endpoint.Port != nil {
				f22.Port = elem.Endpoint.Port
			}
			ko.Status.Endpoint = f22
		}
		if elem.Engine != nil {
			ko.Spec.Engine = elem.Engine
		}
		if elem.EngineVersion != nil {
			ko.Spec.EngineVersion = elem.EngineVersion
		}
		if elem.EnhancedMonitoringResourceArn != nil {
			ko.Status.EnhancedMonitoringResourceARN = elem.EnhancedMonitoringResourceArn
		}
		if elem.IAMDatabaseAuthenticationEnabled != nil {
			ko.Status.IAMDatabaseAuthenticationEnabled = elem.IAMDatabaseAuthenticationEnabled
		}
		if elem.InstanceCreateTime != nil {
			ko.Status.InstanceCreateTime = &metav1.Time{*elem.InstanceCreateTime}
		}
		if elem.Iops != nil {
			ko.Spec.IOPS = elem.Iops
		}
		if elem.KmsKeyId != nil {
			ko.Spec.KMSKeyID = elem.KmsKeyId
		}
		if elem.LatestRestorableTime != nil {
			ko.Status.LatestRestorableTime = &metav1.Time{*elem.LatestRestorableTime}
		}
		if elem.LicenseModel != nil {
			ko.Spec.LicenseModel = elem.LicenseModel
		}
		if elem.ListenerEndpoint != nil {
			f32 := &svcapitypes.Endpoint{}
			if elem.ListenerEndpoint.Address != nil {
				f32.Address = elem.ListenerEndpoint.Address
			}
			if elem.ListenerEndpoint.HostedZoneId != nil {
				f32.HostedZoneID = elem.ListenerEndpoint.HostedZoneId
			}
			if elem.ListenerEndpoint.Port != nil {
				f32.Port = elem.ListenerEndpoint.Port
			}
			ko.Status.ListenerEndpoint = f32
		}
		if elem.MasterUsername != nil {
			ko.Spec.MasterUsername = elem.MasterUsername
		}
		if elem.MaxAllocatedStorage != nil {
			ko.Spec.MaxAllocatedStorage = elem.MaxAllocatedStorage
		}
		if elem.MonitoringInterval != nil {
			ko.Spec.MonitoringInterval = elem.MonitoringInterval
		}
		if elem.MonitoringRoleArn != nil {
			ko.Spec.MonitoringRoleARN = elem.MonitoringRoleArn
		}
		if elem.MultiAZ != nil {
			ko.Spec.MultiAZ = elem.MultiAZ
		}
		if elem.OptionGroupMemberships != nil {
			f38 := []*svcapitypes.OptionGroupMembership{}
			for _, f38iter := range elem.OptionGroupMemberships {
				f38elem := &svcapitypes.OptionGroupMembership{}
				if f38iter.OptionGroupName != nil {
					f38elem.OptionGroupName = f38iter.OptionGroupName
				}
				if f38iter.Status != nil {
					f38elem.Status = f38iter.Status
				}
				f38 = append(f38, f38elem)
			}
			ko.Status.OptionGroupMemberships = f38
		}
		if elem.PendingModifiedValues != nil {
			f39 := &svcapitypes.PendingModifiedValues{}
			if elem.PendingModifiedValues.AllocatedStorage != nil {
				f39.AllocatedStorage = elem.PendingModifiedValues.AllocatedStorage
			}
			if elem.PendingModifiedValues.BackupRetentionPeriod != nil {
				f39.BackupRetentionPeriod = elem.PendingModifiedValues.BackupRetentionPeriod
			}
			if elem.PendingModifiedValues.CACertificateIdentifier != nil {
				f39.CACertificateIdentifier = elem.PendingModifiedValues.CACertificateIdentifier
			}
			if elem.PendingModifiedValues.DBInstanceClass != nil {
				f39.DBInstanceClass = elem.PendingModifiedValues.DBInstanceClass
			}
			if elem.PendingModifiedValues.DBInstanceIdentifier != nil {
				f39.DBInstanceIdentifier = elem.PendingModifiedValues.DBInstanceIdentifier
			}
			if elem.PendingModifiedValues.DBSubnetGroupName != nil {
				f39.DBSubnetGroupName = elem.PendingModifiedValues.DBSubnetGroupName
			}
			if elem.PendingModifiedValues.EngineVersion != nil {
				f39.EngineVersion = elem.PendingModifiedValues.EngineVersion
			}
			if elem.PendingModifiedValues.Iops != nil {
				f39.IOPS = elem.PendingModifiedValues.Iops
			}
			if elem.PendingModifiedValues.LicenseModel != nil {
				f39.LicenseModel = elem.PendingModifiedValues.LicenseModel
			}
			if elem.PendingModifiedValues.MasterUserPassword != nil {
				f39.MasterUserPassword = elem.PendingModifiedValues.MasterUserPassword
			}
			if elem.PendingModifiedValues.MultiAZ != nil {
				f39.MultiAZ = elem.PendingModifiedValues.MultiAZ
			}
			if elem.PendingModifiedValues.PendingCloudwatchLogsExports != nil {
				f39f11 := &svcapitypes.PendingCloudwatchLogsExports{}
				if elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable != nil {
					f39f11f0 := []*string{}
					for _, f39f11f0iter := range elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable {
						var f39f11f0elem string
						f39f11f0elem = *f39f11f0iter
						f39f11f0 = append(f39f11f0, &f39f11f0elem)
					}
					f39f11.LogTypesToDisable = f39f11f0
				}
				if elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable != nil {
					f39f11f1 := []*string{}
					for _, f39f11f1iter := range elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable {
						var f39f11f1elem string
						f39f11f1elem = *f39f11f1iter
						f39f11f1 = append(f39f11f1, &f39f11f1elem)
					}
					f39f11.LogTypesToEnable = f39f11f1
				}
				f39.PendingCloudwatchLogsExports = f39f11
			}
			if elem.PendingModifiedValues.Port != nil {
				f39.Port = elem.PendingModifiedValues.Port
			}
			if elem.PendingModifiedValues.ProcessorFeatures != nil {
				f39f13 := []*svcapitypes.ProcessorFeature{}
				for _, f39f13iter := range elem.PendingModifiedValues.ProcessorFeatures {
					f39f13elem := &svcapitypes.ProcessorFeature{}
					if f39f13iter.Name != nil {
						f39f13elem.Name = f39f13iter.Name
					}
					if f39f13iter.Value != nil {
						f39f13elem.Value = f39f13iter.Value
					}
					f39f13 = append(f39f13, f39f13elem)
				}
				f39.ProcessorFeatures = f39f13
			}
			if elem.PendingModifiedValues.StorageType != nil {
				f39.StorageType = elem.PendingModifiedValues.StorageType
			}
			ko.Status.PendingModifiedValues = f39
		}
		if elem.PerformanceInsightsEnabled != nil {
			ko.Status.PerformanceInsightsEnabled = elem.PerformanceInsightsEnabled
		}
		if elem.PerformanceInsightsKMSKeyId != nil {
			ko.Spec.PerformanceInsightsKMSKeyID = elem.PerformanceInsightsKMSKeyId
		}
		if elem.PerformanceInsightsRetentionPeriod != nil {
			ko.Spec.PerformanceInsightsRetentionPeriod = elem.PerformanceInsightsRetentionPeriod
		}
		if elem.PreferredBackupWindow != nil {
			ko.Spec.PreferredBackupWindow = elem.PreferredBackupWindow
		}
		if elem.PreferredMaintenanceWindow != nil {
			ko.Spec.PreferredMaintenanceWindow = elem.PreferredMaintenanceWindow
		}
		if elem.ProcessorFeatures != nil {
			f45 := []*svcapitypes.ProcessorFeature{}
			for _, f45iter := range elem.ProcessorFeatures {
				f45elem := &svcapitypes.ProcessorFeature{}
				if f45iter.Name != nil {
					f45elem.Name = f45iter.Name
				}
				if f45iter.Value != nil {
					f45elem.Value = f45iter.Value
				}
				f45 = append(f45, f45elem)
			}
			ko.Spec.ProcessorFeatures = f45
		}
		if elem.PromotionTier != nil {
			ko.Spec.PromotionTier = elem.PromotionTier
		}
		if elem.PubliclyAccessible != nil {
			ko.Spec.PubliclyAccessible = elem.PubliclyAccessible
		}
		if elem.ReadReplicaDBClusterIdentifiers != nil {
			f48 := []*string{}
			for _, f48iter := range elem.ReadReplicaDBClusterIdentifiers {
				var f48elem string
				f48elem = *f48iter
				f48 = append(f48, &f48elem)
			}
			ko.Status.ReadReplicaDBClusterIdentifiers = f48
		}
		if elem.ReadReplicaDBInstanceIdentifiers != nil {
			f49 := []*string{}
			for _, f49iter := range elem.ReadReplicaDBInstanceIdentifiers {
				var f49elem string
				f49elem = *f49iter
				f49 = append(f49, &f49elem)
			}
			ko.Status.ReadReplicaDBInstanceIdentifiers = f49
		}
		if elem.ReadReplicaSourceDBInstanceIdentifier != nil {
			ko.Status.ReadReplicaSourceDBInstanceIdentifier = elem.ReadReplicaSourceDBInstanceIdentifier
		}
		if elem.SecondaryAvailabilityZone != nil {
			ko.Status.SecondaryAvailabilityZone = elem.SecondaryAvailabilityZone
		}
		if elem.StatusInfos != nil {
			f52 := []*svcapitypes.DBInstanceStatusInfo{}
			for _, f52iter := range elem.StatusInfos {
				f52elem := &svcapitypes.DBInstanceStatusInfo{}
				if f52iter.Message != nil {
					f52elem.Message = f52iter.Message
				}
				if f52iter.Normal != nil {
					f52elem.Normal = f52iter.Normal
				}
				if f52iter.Status != nil {
					f52elem.Status = f52iter.Status
				}
				if f52iter.StatusType != nil {
					f52elem.StatusType = f52iter.StatusType
				}
				f52 = append(f52, f52elem)
			}
			ko.Status.StatusInfos = f52
		}
		if elem.StorageEncrypted != nil {
			ko.Spec.StorageEncrypted = elem.StorageEncrypted
		}
		if elem.StorageType != nil {
			ko.Spec.StorageType = elem.StorageType
		}
		if elem.TdeCredentialArn != nil {
			ko.Spec.TDECredentialARN = elem.TdeCredentialArn
		}
		if elem.Timezone != nil {
			ko.Spec.Timezone = elem.Timezone
		}
		if elem.VpcSecurityGroups != nil {
			f57 := []*svcapitypes.VPCSecurityGroupMembership{}
			for _, f57iter := range elem.VpcSecurityGroups {
				f57elem := &svcapitypes.VPCSecurityGroupMembership{}
				if f57iter.Status != nil {
					f57elem.Status = f57iter.Status
				}
				if f57iter.VpcSecurityGroupId != nil {
					f57elem.VPCSecurityGroupID = f57iter.VpcSecurityGroupId
				}
				f57 = append(f57, f57elem)
			}
			ko.Status.VPCSecurityGroups = f57
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}
`
	assert.Equal(expReadManyOutput, crd.GoCodeSetOutput(model.OpTypeList, "resp", "ko", 1))
}
