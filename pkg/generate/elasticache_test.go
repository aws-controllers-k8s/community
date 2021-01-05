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
