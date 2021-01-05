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

func TestElasticache_CacheCluster(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("CacheCluster", crds)
	require.NotNil(crd)

	assert.Equal("CacheCluster", crd.Names.Camel)
	assert.Equal("cacheCluster", crd.Names.CamelLower)
	assert.Equal("cache_cluster", crd.Names.Snake)

	// The DescribeCacheClusters operation has the following definition:
	//
	//    "DescribeCacheClusters":{
	//      "name":"DescribeCacheClusters",
	//      "http":{
	//        "method":"POST",
	//        "requestUri":"/"
	//      },
	//      "input":{"shape":"DescribeCacheClustersMessage"},
	//      "output":{
	//        "shape":"CacheClusterMessage",
	//        "resultWrapper":"DescribeCacheClustersResult"
	//      },
	//      "errors":[
	//        {"shape":"CacheClusterNotFoundFault"},
	//        {"shape":"InvalidParameterValueException"},
	//        {"shape":"InvalidParameterCombinationException"}
	//      ]
	//    },
	//
	// Where the CacheClusterNotFoundFault shape looks like this:
	//
	//    "CacheClusterNotFoundFault":{
	//      "type":"structure",
	//      "members":{
	//      },
	//      "error":{
	//        "code":"CacheClusterNotFound",
	//        "httpStatusCode":404,
	//        "senderFault":true
	//      },
	//      "exception":true
	//    },
	//
	// Which indicates that the error is a 404 and is our NotFoundException
	// error with a "code" value of CacheClusterNotFound
	assert.Equal("CacheClusterNotFound", crd.ExceptionCode(404))

	// The Elasticache CacheCluster API has CUD+L operations:
	//
	// * CreateCacheCluster
	// * DeleteCacheCluster
	// * UpdateCacheCluster
	// * GetCacheClusters
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.Update)
	assert.NotNil(crd.Ops.ReadMany)

	// But no ReadOne operation...
	assert.Nil(crd.Ops.ReadOne)

	// And no separate get/set attributes calls.
	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"AZMode",
		"AuthToken",
		"AutoMinorVersionUpgrade",
		"CacheClusterID",
		"CacheNodeType",
		"CacheParameterGroupName",
		"CacheSecurityGroupNames",
		"CacheSubnetGroupName",
		"Engine",
		"EngineVersion",
		"NotificationTopicARN",
		"NumCacheNodes",
		"Port",
		"PreferredAvailabilityZone",
		"PreferredAvailabilityZones",
		"PreferredMaintenanceWindow",
		"ReplicationGroupID",
		"SecurityGroupIDs",
		"SnapshotARNs",
		"SnapshotName",
		"SnapshotRetentionLimit",
		"SnapshotWindow",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"AtRestEncryptionEnabled",
		"AuthTokenEnabled",
		"AuthTokenLastModifiedDate",
		"CacheClusterCreateTime",
		"CacheClusterStatus",
		"CacheNodes",
		"CacheParameterGroup",
		"CacheSecurityGroups",
		"ClientDownloadLandingPage",
		"ConfigurationEndpoint",
		"NotificationConfiguration",
		"PendingModifiedValues",
		"SecurityGroups",
		"TransitEncryptionEnabled",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	expCreateOutput := `
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.CacheCluster.ARN != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.CacheCluster.ARN)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.CacheCluster.AtRestEncryptionEnabled != nil {
		ko.Status.AtRestEncryptionEnabled = resp.CacheCluster.AtRestEncryptionEnabled
	}
	if resp.CacheCluster.AuthTokenEnabled != nil {
		ko.Status.AuthTokenEnabled = resp.CacheCluster.AuthTokenEnabled
	}
	if resp.CacheCluster.AuthTokenLastModifiedDate != nil {
		ko.Status.AuthTokenLastModifiedDate = &metav1.Time{*resp.CacheCluster.AuthTokenLastModifiedDate}
	}
	if resp.CacheCluster.CacheClusterCreateTime != nil {
		ko.Status.CacheClusterCreateTime = &metav1.Time{*resp.CacheCluster.CacheClusterCreateTime}
	}
	if resp.CacheCluster.CacheClusterStatus != nil {
		ko.Status.CacheClusterStatus = resp.CacheCluster.CacheClusterStatus
	}
	if resp.CacheCluster.CacheNodes != nil {
		f9 := []*svcapitypes.CacheNode{}
		for _, f9iter := range resp.CacheCluster.CacheNodes {
			f9elem := &svcapitypes.CacheNode{}
			if f9iter.CacheNodeCreateTime != nil {
				f9elem.CacheNodeCreateTime = &metav1.Time{*f9iter.CacheNodeCreateTime}
			}
			if f9iter.CacheNodeId != nil {
				f9elem.CacheNodeID = f9iter.CacheNodeId
			}
			if f9iter.CacheNodeStatus != nil {
				f9elem.CacheNodeStatus = f9iter.CacheNodeStatus
			}
			if f9iter.CustomerAvailabilityZone != nil {
				f9elem.CustomerAvailabilityZone = f9iter.CustomerAvailabilityZone
			}
			if f9iter.Endpoint != nil {
				f9elemf4 := &svcapitypes.Endpoint{}
				if f9iter.Endpoint.Address != nil {
					f9elemf4.Address = f9iter.Endpoint.Address
				}
				if f9iter.Endpoint.Port != nil {
					f9elemf4.Port = f9iter.Endpoint.Port
				}
				f9elem.Endpoint = f9elemf4
			}
			if f9iter.ParameterGroupStatus != nil {
				f9elem.ParameterGroupStatus = f9iter.ParameterGroupStatus
			}
			if f9iter.SourceCacheNodeId != nil {
				f9elem.SourceCacheNodeID = f9iter.SourceCacheNodeId
			}
			f9 = append(f9, f9elem)
		}
		ko.Status.CacheNodes = f9
	}
	if resp.CacheCluster.CacheParameterGroup != nil {
		f10 := &svcapitypes.CacheParameterGroupStatus_SDK{}
		if resp.CacheCluster.CacheParameterGroup.CacheNodeIdsToReboot != nil {
			f10f0 := []*string{}
			for _, f10f0iter := range resp.CacheCluster.CacheParameterGroup.CacheNodeIdsToReboot {
				var f10f0elem string
				f10f0elem = *f10f0iter
				f10f0 = append(f10f0, &f10f0elem)
			}
			f10.CacheNodeIDsToReboot = f10f0
		}
		if resp.CacheCluster.CacheParameterGroup.CacheParameterGroupName != nil {
			f10.CacheParameterGroupName = resp.CacheCluster.CacheParameterGroup.CacheParameterGroupName
		}
		if resp.CacheCluster.CacheParameterGroup.ParameterApplyStatus != nil {
			f10.ParameterApplyStatus = resp.CacheCluster.CacheParameterGroup.ParameterApplyStatus
		}
		ko.Status.CacheParameterGroup = f10
	}
	if resp.CacheCluster.CacheSecurityGroups != nil {
		f11 := []*svcapitypes.CacheSecurityGroupMembership{}
		for _, f11iter := range resp.CacheCluster.CacheSecurityGroups {
			f11elem := &svcapitypes.CacheSecurityGroupMembership{}
			if f11iter.CacheSecurityGroupName != nil {
				f11elem.CacheSecurityGroupName = f11iter.CacheSecurityGroupName
			}
			if f11iter.Status != nil {
				f11elem.Status = f11iter.Status
			}
			f11 = append(f11, f11elem)
		}
		ko.Status.CacheSecurityGroups = f11
	}
	if resp.CacheCluster.ClientDownloadLandingPage != nil {
		ko.Status.ClientDownloadLandingPage = resp.CacheCluster.ClientDownloadLandingPage
	}
	if resp.CacheCluster.ConfigurationEndpoint != nil {
		f14 := &svcapitypes.Endpoint{}
		if resp.CacheCluster.ConfigurationEndpoint.Address != nil {
			f14.Address = resp.CacheCluster.ConfigurationEndpoint.Address
		}
		if resp.CacheCluster.ConfigurationEndpoint.Port != nil {
			f14.Port = resp.CacheCluster.ConfigurationEndpoint.Port
		}
		ko.Status.ConfigurationEndpoint = f14
	}
	if resp.CacheCluster.NotificationConfiguration != nil {
		f17 := &svcapitypes.NotificationConfiguration{}
		if resp.CacheCluster.NotificationConfiguration.TopicArn != nil {
			f17.TopicARN = resp.CacheCluster.NotificationConfiguration.TopicArn
		}
		if resp.CacheCluster.NotificationConfiguration.TopicStatus != nil {
			f17.TopicStatus = resp.CacheCluster.NotificationConfiguration.TopicStatus
		}
		ko.Status.NotificationConfiguration = f17
	}
	if resp.CacheCluster.PendingModifiedValues != nil {
		f19 := &svcapitypes.PendingModifiedValues{}
		if resp.CacheCluster.PendingModifiedValues.AuthTokenStatus != nil {
			f19.AuthTokenStatus = resp.CacheCluster.PendingModifiedValues.AuthTokenStatus
		}
		if resp.CacheCluster.PendingModifiedValues.CacheNodeIdsToRemove != nil {
			f19f1 := []*string{}
			for _, f19f1iter := range resp.CacheCluster.PendingModifiedValues.CacheNodeIdsToRemove {
				var f19f1elem string
				f19f1elem = *f19f1iter
				f19f1 = append(f19f1, &f19f1elem)
			}
			f19.CacheNodeIDsToRemove = f19f1
		}
		if resp.CacheCluster.PendingModifiedValues.CacheNodeType != nil {
			f19.CacheNodeType = resp.CacheCluster.PendingModifiedValues.CacheNodeType
		}
		if resp.CacheCluster.PendingModifiedValues.EngineVersion != nil {
			f19.EngineVersion = resp.CacheCluster.PendingModifiedValues.EngineVersion
		}
		if resp.CacheCluster.PendingModifiedValues.NumCacheNodes != nil {
			f19.NumCacheNodes = resp.CacheCluster.PendingModifiedValues.NumCacheNodes
		}
		ko.Status.PendingModifiedValues = f19
	}
	if resp.CacheCluster.SecurityGroups != nil {
		f23 := []*svcapitypes.SecurityGroupMembership{}
		for _, f23iter := range resp.CacheCluster.SecurityGroups {
			f23elem := &svcapitypes.SecurityGroupMembership{}
			if f23iter.SecurityGroupId != nil {
				f23elem.SecurityGroupID = f23iter.SecurityGroupId
			}
			if f23iter.Status != nil {
				f23elem.Status = f23iter.Status
			}
			f23 = append(f23, f23elem)
		}
		ko.Status.SecurityGroups = f23
	}
	if resp.CacheCluster.TransitEncryptionEnabled != nil {
		ko.Status.TransitEncryptionEnabled = resp.CacheCluster.TransitEncryptionEnabled
	}
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko", 1, false))

	expReadManyOutput := `
	found := false
	for _, elem := range resp.CacheClusters {
		if elem.ARN != nil {
			if ko.Status.ACKResourceMetadata == nil {
				ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
			}
			tmpARN := ackv1alpha1.AWSResourceName(*elem.ARN)
			ko.Status.ACKResourceMetadata.ARN = &tmpARN
		}
		if elem.AtRestEncryptionEnabled != nil {
			ko.Status.AtRestEncryptionEnabled = elem.AtRestEncryptionEnabled
		}
		if elem.AuthTokenEnabled != nil {
			ko.Status.AuthTokenEnabled = elem.AuthTokenEnabled
		}
		if elem.AuthTokenLastModifiedDate != nil {
			ko.Status.AuthTokenLastModifiedDate = &metav1.Time{*elem.AuthTokenLastModifiedDate}
		}
		if elem.AutoMinorVersionUpgrade != nil {
			ko.Spec.AutoMinorVersionUpgrade = elem.AutoMinorVersionUpgrade
		}
		if elem.CacheClusterCreateTime != nil {
			ko.Status.CacheClusterCreateTime = &metav1.Time{*elem.CacheClusterCreateTime}
		}
		if elem.CacheClusterId != nil {
			ko.Spec.CacheClusterID = elem.CacheClusterId
		}
		if elem.CacheClusterStatus != nil {
			ko.Status.CacheClusterStatus = elem.CacheClusterStatus
		}
		if elem.CacheNodeType != nil {
			ko.Spec.CacheNodeType = elem.CacheNodeType
		}
		if elem.CacheNodes != nil {
			f9 := []*svcapitypes.CacheNode{}
			for _, f9iter := range elem.CacheNodes {
				f9elem := &svcapitypes.CacheNode{}
				if f9iter.CacheNodeCreateTime != nil {
					f9elem.CacheNodeCreateTime = &metav1.Time{*f9iter.CacheNodeCreateTime}
				}
				if f9iter.CacheNodeId != nil {
					f9elem.CacheNodeID = f9iter.CacheNodeId
				}
				if f9iter.CacheNodeStatus != nil {
					f9elem.CacheNodeStatus = f9iter.CacheNodeStatus
				}
				if f9iter.CustomerAvailabilityZone != nil {
					f9elem.CustomerAvailabilityZone = f9iter.CustomerAvailabilityZone
				}
				if f9iter.Endpoint != nil {
					f9elemf4 := &svcapitypes.Endpoint{}
					if f9iter.Endpoint.Address != nil {
						f9elemf4.Address = f9iter.Endpoint.Address
					}
					if f9iter.Endpoint.Port != nil {
						f9elemf4.Port = f9iter.Endpoint.Port
					}
					f9elem.Endpoint = f9elemf4
				}
				if f9iter.ParameterGroupStatus != nil {
					f9elem.ParameterGroupStatus = f9iter.ParameterGroupStatus
				}
				if f9iter.SourceCacheNodeId != nil {
					f9elem.SourceCacheNodeID = f9iter.SourceCacheNodeId
				}
				f9 = append(f9, f9elem)
			}
			ko.Status.CacheNodes = f9
		}
		if elem.CacheParameterGroup != nil {
			f10 := &svcapitypes.CacheParameterGroupStatus_SDK{}
			if elem.CacheParameterGroup.CacheNodeIdsToReboot != nil {
				f10f0 := []*string{}
				for _, f10f0iter := range elem.CacheParameterGroup.CacheNodeIdsToReboot {
					var f10f0elem string
					f10f0elem = *f10f0iter
					f10f0 = append(f10f0, &f10f0elem)
				}
				f10.CacheNodeIDsToReboot = f10f0
			}
			if elem.CacheParameterGroup.CacheParameterGroupName != nil {
				f10.CacheParameterGroupName = elem.CacheParameterGroup.CacheParameterGroupName
			}
			if elem.CacheParameterGroup.ParameterApplyStatus != nil {
				f10.ParameterApplyStatus = elem.CacheParameterGroup.ParameterApplyStatus
			}
			ko.Status.CacheParameterGroup = f10
		}
		if elem.CacheSecurityGroups != nil {
			f11 := []*svcapitypes.CacheSecurityGroupMembership{}
			for _, f11iter := range elem.CacheSecurityGroups {
				f11elem := &svcapitypes.CacheSecurityGroupMembership{}
				if f11iter.CacheSecurityGroupName != nil {
					f11elem.CacheSecurityGroupName = f11iter.CacheSecurityGroupName
				}
				if f11iter.Status != nil {
					f11elem.Status = f11iter.Status
				}
				f11 = append(f11, f11elem)
			}
			ko.Status.CacheSecurityGroups = f11
		}
		if elem.CacheSubnetGroupName != nil {
			ko.Spec.CacheSubnetGroupName = elem.CacheSubnetGroupName
		}
		if elem.ClientDownloadLandingPage != nil {
			ko.Status.ClientDownloadLandingPage = elem.ClientDownloadLandingPage
		}
		if elem.ConfigurationEndpoint != nil {
			f14 := &svcapitypes.Endpoint{}
			if elem.ConfigurationEndpoint.Address != nil {
				f14.Address = elem.ConfigurationEndpoint.Address
			}
			if elem.ConfigurationEndpoint.Port != nil {
				f14.Port = elem.ConfigurationEndpoint.Port
			}
			ko.Status.ConfigurationEndpoint = f14
		}
		if elem.Engine != nil {
			ko.Spec.Engine = elem.Engine
		}
		if elem.EngineVersion != nil {
			ko.Spec.EngineVersion = elem.EngineVersion
		}
		if elem.NotificationConfiguration != nil {
			f17 := &svcapitypes.NotificationConfiguration{}
			if elem.NotificationConfiguration.TopicArn != nil {
				f17.TopicARN = elem.NotificationConfiguration.TopicArn
			}
			if elem.NotificationConfiguration.TopicStatus != nil {
				f17.TopicStatus = elem.NotificationConfiguration.TopicStatus
			}
			ko.Status.NotificationConfiguration = f17
		}
		if elem.NumCacheNodes != nil {
			ko.Spec.NumCacheNodes = elem.NumCacheNodes
		}
		if elem.PendingModifiedValues != nil {
			f19 := &svcapitypes.PendingModifiedValues{}
			if elem.PendingModifiedValues.AuthTokenStatus != nil {
				f19.AuthTokenStatus = elem.PendingModifiedValues.AuthTokenStatus
			}
			if elem.PendingModifiedValues.CacheNodeIdsToRemove != nil {
				f19f1 := []*string{}
				for _, f19f1iter := range elem.PendingModifiedValues.CacheNodeIdsToRemove {
					var f19f1elem string
					f19f1elem = *f19f1iter
					f19f1 = append(f19f1, &f19f1elem)
				}
				f19.CacheNodeIDsToRemove = f19f1
			}
			if elem.PendingModifiedValues.CacheNodeType != nil {
				f19.CacheNodeType = elem.PendingModifiedValues.CacheNodeType
			}
			if elem.PendingModifiedValues.EngineVersion != nil {
				f19.EngineVersion = elem.PendingModifiedValues.EngineVersion
			}
			if elem.PendingModifiedValues.NumCacheNodes != nil {
				f19.NumCacheNodes = elem.PendingModifiedValues.NumCacheNodes
			}
			ko.Status.PendingModifiedValues = f19
		}
		if elem.PreferredAvailabilityZone != nil {
			ko.Spec.PreferredAvailabilityZone = elem.PreferredAvailabilityZone
		}
		if elem.PreferredMaintenanceWindow != nil {
			ko.Spec.PreferredMaintenanceWindow = elem.PreferredMaintenanceWindow
		}
		if elem.ReplicationGroupId != nil {
			ko.Spec.ReplicationGroupID = elem.ReplicationGroupId
		}
		if elem.SecurityGroups != nil {
			f23 := []*svcapitypes.SecurityGroupMembership{}
			for _, f23iter := range elem.SecurityGroups {
				f23elem := &svcapitypes.SecurityGroupMembership{}
				if f23iter.SecurityGroupId != nil {
					f23elem.SecurityGroupID = f23iter.SecurityGroupId
				}
				if f23iter.Status != nil {
					f23elem.Status = f23iter.Status
				}
				f23 = append(f23, f23elem)
			}
			ko.Status.SecurityGroups = f23
		}
		if elem.SnapshotRetentionLimit != nil {
			ko.Spec.SnapshotRetentionLimit = elem.SnapshotRetentionLimit
		}
		if elem.SnapshotWindow != nil {
			ko.Spec.SnapshotWindow = elem.SnapshotWindow
		}
		if elem.TransitEncryptionEnabled != nil {
			ko.Status.TransitEncryptionEnabled = elem.TransitEncryptionEnabled
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}
`
	assert.Equal(expReadManyOutput, crd.GoCodeSetOutput(model.OpTypeList, "resp", "ko", 1, true))
}

func TestElasticache_Ignored_Operations(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Snapshot", crds)
	require.NotNil(crd)
	require.NotNil(crd.Ops.Create)
	require.Nil(crd.Ops.Delete)
}

func TestElasticache_Ignored_Resources(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")

	crds, err := g.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("GlobalReplicationGroup", crds)
	require.Nil(crd)
}

func TestElasticache_Additional_Snapshot_Spec(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("Snapshot", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.SpecFields, "SourceSnapshotName")
}

func TestElasticache_Additional_CacheParameterGroup_Spec(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("CacheParameterGroup", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.SpecFields, "ParameterNameValues")
}

func TestElasticache_Additional_CacheParameterGroup_Status(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("CacheParameterGroup", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.StatusFields, "Parameters")
	assert.Contains(crd.StatusFields, "Events")
}

func TestElasticache_Additional_ReplicationGroup_Status(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("ReplicationGroup", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.StatusFields, "Events")
}

func TestElasticache_Additional_CacheSubnetGroup_Status(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("CacheSubnetGroup", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.StatusFields, "Events")
}

func TestElasticache_Additional_ReplicationGroup_Status_RenameField(t *testing.T) {
	require := require.New(t)

	g := testutil.NewGeneratorForService(t, "elasticache")
	crds, err := g.GetCRDs()

	require.Nil(err)

	crd := getCRDByName("ReplicationGroup", crds)
	require.NotNil(crd)

	assert := assert.New(t)
	assert.Contains(crd.StatusFields, "AllowedScaleUpModifications")
	assert.Contains(crd.StatusFields, "AllowedScaleDownModifications")
}
