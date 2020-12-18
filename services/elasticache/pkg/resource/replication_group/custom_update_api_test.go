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

package replication_group

import (
	"context"
	"fmt"
	"github.com/aws/aws-controllers-k8s/pkg/requeue"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	ctrlrtzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
)

// Helper methods to setup tests
// provideResourceManager returns pointer to resourceManager
func provideResourceManager() *resourceManager {
	zapOptions := ctrlrtzap.Options{
		Development: true,
		Level:       zapcore.InfoLevel,
	}
	fakeLogger := ctrlrtzap.New(ctrlrtzap.UseFlagOptions(&zapOptions))
	return &resourceManager{
		rr:           nil,
		awsAccountID: "",
		awsRegion:    "",
		sess:         nil,
		sdkapi:       nil,
		log:          fakeLogger,
	}
}

// provideResource returns pointer to resource
func provideResource() *resource {
	return provideResourceWithStatus("available")
}

// provideCacheCluster returns pointer to CacheCluster
func provideCacheCluster() *svcsdk.CacheCluster {
	return &svcsdk.CacheCluster{}
}

// provideResource returns pointer to resource
func provideResourceWithStatus(rgStatus string) *resource {
	return &resource{
		ko: &svcapitypes.ReplicationGroup{
			Status: svcapitypes.ReplicationGroupStatus{
				Status: &rgStatus,
			},
		},
	}
}

// provideNodeGroups provides NodeGroups array for given node IDs
func provideNodeGroups(IDs ...string) []*svcapitypes.NodeGroup {
	return provideNodeGroupsWithReplicas(3, IDs...)
}

// provideMemberClusters returns the member cluster Ids from given node groups
func provideMemberClusters(nodeGroups []*svcapitypes.NodeGroup) []*string {
	if nodeGroups == nil {
		return nil
	}
	memberClusters := []*string{}
	for _, nodeGroup := range nodeGroups {
		for _, member := range nodeGroup.NodeGroupMembers {
			cacheClusterId := *member.CacheClusterID
			memberClusters = append(memberClusters, &cacheClusterId)
		}
	}
	return memberClusters
}

// provideNodeGroupsWithReplicas provides NodeGroups array for given node IDs
// each node group is populated with supplied numbers of replica nodes and a primary node.
func provideNodeGroupsWithReplicas(replicasCount int, IDs ...string) []*svcapitypes.NodeGroup {
	nodeGroups := []*svcapitypes.NodeGroup{}
	for _, ID := range IDs {
		nodeId := ID
		nodeGroups = append(nodeGroups, &svcapitypes.NodeGroup{
			NodeGroupID:      &nodeId,
			NodeGroupMembers: provideNodeGroupMembers(&nodeId, replicasCount+1), // primary node + replicas
			PrimaryEndpoint:  nil,
			ReaderEndpoint:   nil,
			Slots:            nil,
			Status:           nil,
		})
	}
	return nodeGroups
}

// provideNodeGroupMembers returns array of NodeGroupMember (replicas and a primary node) for given shard id
func provideNodeGroupMembers(nodeID *string, membersCount int) []*svcapitypes.NodeGroupMember {
	if membersCount <= 0 {
		return nil
	}
	rolePrimary := "primary"
	roleReplica := "replica"
	availabilityZones := provideAvailabilityZones(*nodeID, membersCount)

	members := []*svcapitypes.NodeGroupMember{}
	// primary
	primary := &svcapitypes.NodeGroupMember{}
	cacheClusterId := fmt.Sprintf("RG-%s-00%d", *nodeID, 1)
	primary.CacheClusterID = &cacheClusterId
	primary.CurrentRole = &rolePrimary
	primary.PreferredAvailabilityZone = availabilityZones[0]
	members = append(members, primary)
	// replicas
	for i := 1; i <= membersCount-1; i++ {
		replica := &svcapitypes.NodeGroupMember{}
		cacheClusterId := fmt.Sprintf("RG-%s-00%d", *nodeID, i+1)
		replica.CacheClusterID = &cacheClusterId
		replica.CacheNodeID = nodeID
		replica.CurrentRole = &roleReplica
		replica.PreferredAvailabilityZone = availabilityZones[i]
		members = append(members, replica)
	}
	return members
}

func provideNodeGroupConfiguration(IDs ...string) []*svcapitypes.NodeGroupConfiguration {
	replicasCount := 3
	return provideNodeGroupConfigurationWithReplicas(replicasCount, IDs...)
}

// provideNodeGroupConfiguration provides NodeGroupConfiguration array for given node IDs and replica count
func provideNodeGroupConfigurationWithReplicas(
	replicaCount int, IDs ...string,
) []*svcapitypes.NodeGroupConfiguration {
	nodeGroupConfig := []*svcapitypes.NodeGroupConfiguration{}
	for _, ID := range IDs {
		nodeId := ID
		azCount := replicaCount + 1 // replicas + a primary node
		numberOfReplicas := int64(replicaCount)
		availabilityZones := provideAvailabilityZones(nodeId, azCount)
		nodeGroupConfig = append(nodeGroupConfig, &svcapitypes.NodeGroupConfiguration{
			NodeGroupID:              &nodeId,
			PrimaryAvailabilityZone:  availabilityZones[0],
			ReplicaAvailabilityZones: availabilityZones[1:],
			ReplicaCount:             &numberOfReplicas,
			Slots:                    nil,
		})
	}

	return nodeGroupConfig
}

// provideAvailabilityZones returns availability zones array for given nodeId
func provideAvailabilityZones(nodeId string, count int) []*string {
	if count <= 0 {
		return nil
	}

	availabilityZones := []*string{}
	for i := 1; i <= count; i++ {
		az := fmt.Sprintf("%s_%s%d", nodeId, "az", i)
		availabilityZones = append(availabilityZones, &az)
	}
	return availabilityZones
}

// validatePayloadReshardingConfig validates given payloadReshardingConfigs against given desiredNodeGroupConfigs
// this is used for tests that are related to shard configuration (scale in/out)
func validatePayloadReshardingConfig(
	desiredNodeGroupConfigs []*svcapitypes.NodeGroupConfiguration,
	payloadReshardingConfigs []*svcsdk.ReshardingConfiguration,
	assert *assert.Assertions,
	require *require.Assertions,
) {
	assert.NotNil(desiredNodeGroupConfigs)
	require.NotNil(payloadReshardingConfigs) // built as provided in desired object NodeGroupConfiguration
	for _, desiredNodeGroup := range desiredNodeGroupConfigs {
		found := false
		for _, payloadReshardConfig := range payloadReshardingConfigs {
			require.NotNil(payloadReshardConfig.PreferredAvailabilityZones)
			if *desiredNodeGroup.NodeGroupID == *payloadReshardConfig.NodeGroupId {
				found = true
				expectedShardAZs := []*string{desiredNodeGroup.PrimaryAvailabilityZone}
				for _, expectedAZ := range desiredNodeGroup.ReplicaAvailabilityZones {
					expectedShardAZs = append(expectedShardAZs, expectedAZ)
				}
				assert.Equal(len(expectedShardAZs), len(payloadReshardConfig.PreferredAvailabilityZones),
					"Node group id %s", *desiredNodeGroup.NodeGroupID)
				for i := 0; i < len(expectedShardAZs); i++ {
					assert.Equal(*expectedShardAZs[i], *payloadReshardConfig.PreferredAvailabilityZones[i],
						"Node group id %s", *desiredNodeGroup.NodeGroupID)
				}
				break
			}
		}
		assert.True(found, "Expected node group id %s not found in payload", *desiredNodeGroup.NodeGroupID)
	}
	assert.Equal(len(desiredNodeGroupConfigs), len(payloadReshardingConfigs))
}

// validatePayloadReplicaConfig validates given payloadReplicaConfigs against given desiredNodeGroupConfigs
// this is used for tests that are related to increase/decrease replica count.
func validatePayloadReplicaConfig(
	desiredNodeGroupConfigs []*svcapitypes.NodeGroupConfiguration,
	payloadReplicaConfigs []*svcsdk.ConfigureShard,
	assert *assert.Assertions,
	require *require.Assertions,
) {
	assert.NotNil(desiredNodeGroupConfigs)
	require.NotNil(payloadReplicaConfigs) // built as provided in desired object NodeGroupConfiguration
	for _, desiredNodeGroup := range desiredNodeGroupConfigs {
		found := false
		for _, payloadShard := range payloadReplicaConfigs {
			require.NotNil(payloadShard.PreferredAvailabilityZones)
			if *desiredNodeGroup.NodeGroupID == *payloadShard.NodeGroupId {
				found = true
				// validate replica count
				assert.Equal(*desiredNodeGroup.ReplicaCount, *payloadShard.NewReplicaCount)

				// validate AZs
				expectedShardAZs := []*string{desiredNodeGroup.PrimaryAvailabilityZone}
				for _, expectedAZ := range desiredNodeGroup.ReplicaAvailabilityZones {
					expectedShardAZs = append(expectedShardAZs, expectedAZ)
				}
				assert.Equal(len(expectedShardAZs), len(payloadShard.PreferredAvailabilityZones),
					"Node group id %s", *desiredNodeGroup.NodeGroupID)
				for i := 0; i < len(expectedShardAZs); i++ {
					assert.Equal(*expectedShardAZs[i], *payloadShard.PreferredAvailabilityZones[i],
						"Node group id %s", *desiredNodeGroup.NodeGroupID)
				}
				break
			}
		}
		assert.True(found, "Expected node group id %s not found in payload", *desiredNodeGroup.NodeGroupID)
	}
	assert.Equal(len(desiredNodeGroupConfigs), len(payloadReplicaConfigs))
}

func TestCustomModifyReplicationGroup(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager()
	// Tests
	t.Run("NoAction=NoDiff", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		var diffReporter ackcompare.Reporter
		var ctx context.Context
		res, err := rm.CustomModifyReplicationGroup(ctx, desired, latest, &diffReporter)
		assert.Nil(res)
		assert.Nil(err)
	})
}

func TestCustomModifyReplicationGroup_Unavailable(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager()
	// Tests
	t.Run("UnavailableRG=Requeue", func(t *testing.T) {
		desired := provideResource()
		latest := provideResourceWithStatus("modifying")
		var diffReporter ackcompare.Reporter
		var ctx context.Context
		res, err := rm.CustomModifyReplicationGroup(ctx, desired, latest, &diffReporter)
		assert.Nil(res)
		assert.NotNil(err)
		var requeueNeededAfter *requeue.RequeueNeededAfter
		assert.True(errors.As(err, &requeueNeededAfter))
	})
}

func TestCustomModifyReplicationGroup_NodeGroup_Unvailable(t *testing.T) {
	assert := assert.New(t)
	// Setup
	rm := provideResourceManager()
	// Tests
	t.Run("UnavailableNodeGroup=Requeue", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		unavailableStatus := "modifying"
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			nodeGroup.Status = &unavailableStatus
		}
		var diffReporter ackcompare.Reporter
		var ctx context.Context
		res, err := rm.CustomModifyReplicationGroup(ctx, desired, latest, &diffReporter)
		assert.Nil(res)
		assert.NotNil(err)
		var requeueNeededAfter *requeue.RequeueNeededAfter
		assert.True(errors.As(err, &requeueNeededAfter))
	})
}

func TestCustomModifyReplicationGroup_NodeGroup_MemberClusters_mismatch(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// Setup
	rm := provideResourceManager()
	// Tests
	t.Run("nodeGroup_memberClustersMismatch=Diff", func(t *testing.T) {
		desired := provideResource()
		desired.ko.Status.NodeGroups = provideNodeGroups("1001", "1002")
		latest := provideResource()
		latest.ko.Status.NodeGroups = provideNodeGroups("1001", "1002")
		latest.ko.Status.MemberClusters = provideMemberClusters(latest.ko.Status.NodeGroups)
		surplusMemberCluster := "RG-Surplus-Member-Cluster"
		latest.ko.Status.MemberClusters = append(latest.ko.Status.MemberClusters, &surplusMemberCluster)
		availableStatus := "available"
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			nodeGroup.Status = &availableStatus
		}
		var diffReporter ackcompare.Reporter
		var ctx context.Context
		require.NotNil(latest.ko.Status.MemberClusters)
		res, err := rm.CustomModifyReplicationGroup(ctx, desired, latest, &diffReporter)
		assert.Nil(res)
		assert.NotNil(err) // due to surplus member cluster
		var requeueNeededAfter *requeue.RequeueNeededAfter
		assert.True(errors.As(err, &requeueNeededAfter))
	})
}

func TestCustomModifyReplicationGroup_NodeGroup_available(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// Setup
	rm := provideResourceManager()
	// Tests
	t.Run("availableNodeGroup=NoDiff", func(t *testing.T) {
		desired := provideResource()
		desired.ko.Status.NodeGroups = provideNodeGroups("1001")
		latest := provideResource()
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		latest.ko.Status.MemberClusters = provideMemberClusters(latest.ko.Status.NodeGroups)
		availableStatus := "available"
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			nodeGroup.Status = &availableStatus
		}
		var diffReporter ackcompare.Reporter
		var ctx context.Context
		require.NotNil(latest.ko.Status.MemberClusters)
		res, err := rm.CustomModifyReplicationGroup(ctx, desired, latest, &diffReporter)
		assert.Nil(res)
		assert.Nil(err)
	})
}

// TestReplicaCountDifference tests scenarios to check if desired, latest replica count
// configurations differ
func TestReplicaCountDifference(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("NoDiff=NoSpec_NoStatus", func(t *testing.T) {
		// no replica configuration in spec as well as status
		desired := provideResource()
		latest := provideResource()
		diff := rm.replicaCountDifference(desired, latest)
		assert.Nil(desired.ko.Spec.ReplicasPerNodeGroup)
		assert.Nil(desired.ko.Spec.NodeGroupConfiguration)
		assert.Nil(latest.ko.Status.NodeGroups)
		assert.Equal(0, diff)
	})
	t.Run("NoDiff=NoSpec_Status.NodeGroups", func(t *testing.T) {
		// no replica configuration in spec but status has nodes as replicas
		desired := provideResource()
		latest := provideResource()
		replicasCount := 2
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(replicasCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.Nil(desired.ko.Spec.ReplicasPerNodeGroup)
		assert.Nil(desired.ko.Spec.NodeGroupConfiguration)
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(replicasCount+1, len(nodeGroup.NodeGroupMembers)) // replica + primary node
		}
		assert.Equal(0, diff)
	})
	t.Run("NoDiff=Spec.ReplicasPerNodeGroup_Status.NodeGroups", func(t *testing.T) {
		// replica configuration in spec as 'ReplicasPerNodeGroup' and status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		replicaCount := int64(2)
		desired.ko.Spec.ReplicasPerNodeGroup = &replicaCount
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(int(replicaCount), "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.Nil(desired.ko.Spec.NodeGroupConfiguration)
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(int(replicaCount)+1, len(nodeGroup.NodeGroupMembers)) // replica + primary node
		}
		assert.Equal(0, diff)
	})
	t.Run("NoDiff=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		// no 'ReplicasPerNodeGroup' in spec but spec has 'NodeGroupConfiguration' with replicas details
		// status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		replicaCount := 2
		desired.ko.Spec.ReplicasPerNodeGroup = nil
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(replicaCount, "1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(replicaCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.NotNil(desired.ko.Spec.NodeGroupConfiguration)
		for _, nodeGroupConfig := range desired.ko.Spec.NodeGroupConfiguration {
			require.NotNil(nodeGroupConfig.ReplicaCount)
			assert.Equal(replicaCount, int(*nodeGroupConfig.ReplicaCount))
		}
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(replicaCount+1, len(nodeGroup.NodeGroupMembers)) // replica + primary node
		}
		assert.Equal(0, diff)
	})
	t.Run("NoDiff=Prefer_Spec.ReplicasPerNodeGroup", func(t *testing.T) {
		// prefer 'ReplicasPerNodeGroup over 'NodeGroupConfiguration' in desired configuration:
		// 'ReplicasPerNodeGroup' in desired spec as well as 'NodeGroupConfiguration' with different desired replicas details.
		// latest status has matching number of replicas with desired 'ReplicasPerNodeGroup'
		desired := provideResource()
		latest := provideResource()
		replicaCount := int64(2)
		desired.ko.Spec.ReplicasPerNodeGroup = &replicaCount
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(int(replicaCount)+1, "1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(int(replicaCount), "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.NotNil(desired.ko.Spec.NodeGroupConfiguration)
		for _, nodeGroupConfig := range desired.ko.Spec.NodeGroupConfiguration {
			require.NotNil(nodeGroupConfig.ReplicaCount)
			assert.Equal(int(replicaCount)+1, int(*nodeGroupConfig.ReplicaCount))
		}
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(int(replicaCount)+1, len(nodeGroup.NodeGroupMembers)) // replica + primary node
		}
		assert.Equal(0, diff)
	})
	t.Run("DiffIncreaseReplica=Spec.ReplicasPerNodeGroup_Status.NodeGroups", func(t *testing.T) {
		// replica configuration in spec as 'ReplicasPerNodeGroup' and status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		desiredReplicaCount := int64(2)
		latestReplicaCount := 1
		desired.ko.Spec.ReplicasPerNodeGroup = &desiredReplicaCount
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(latestReplicaCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.Nil(desired.ko.Spec.NodeGroupConfiguration)
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(latestReplicaCount+1, len(nodeGroup.NodeGroupMembers)) // replicas + 1 primary node
		}
		assert.True(diff > 0) // desired replicas > latest replicas
	})
	t.Run("DiffIncreaseReplica=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		// no 'ReplicasPerNodeGroup' in spec but spec has 'NodeGroupConfiguration' with replicas details
		// status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		desiredReplicaCount := 2
		latestReplicaCount := 1
		desired.ko.Spec.ReplicasPerNodeGroup = nil
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(desiredReplicaCount, "1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(latestReplicaCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.NotNil(desired.ko.Spec.NodeGroupConfiguration)
		for _, nodeGroupConfig := range desired.ko.Spec.NodeGroupConfiguration {
			require.NotNil(nodeGroupConfig.ReplicaCount)
			assert.Equal(desiredReplicaCount, int(*nodeGroupConfig.ReplicaCount))
		}
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(latestReplicaCount+1, len(nodeGroup.NodeGroupMembers)) // replicas + primary node
		}
		assert.True(diff > 0) // desired replicas > latest replicas
	})
	t.Run("DiffDecreaseReplica=Spec.ReplicasPerNodeGroup_Status.NodeGroups", func(t *testing.T) {
		// replica configuration in spec as 'ReplicasPerNodeGroup' and status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		desiredReplicaCount := int64(2)
		latestReplicaCount := 3
		desired.ko.Spec.ReplicasPerNodeGroup = &desiredReplicaCount
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(latestReplicaCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.Nil(desired.ko.Spec.NodeGroupConfiguration)
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(latestReplicaCount+1, len(nodeGroup.NodeGroupMembers)) // replicas + 1 primary node
		}
		assert.True(diff < 0) // desired replicas < latest replicas
	})
	t.Run("DiffDecreaseReplica=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		// no 'ReplicasPerNodeGroup' in spec but spec has 'NodeGroupConfiguration' with replicas details
		// status has matching number of replicas
		desired := provideResource()
		latest := provideResource()
		desiredReplicaCount := 2
		latestReplicaCount := 3
		desired.ko.Spec.ReplicasPerNodeGroup = nil
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(desiredReplicaCount, "1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(latestReplicaCount, "1001")
		diff := rm.replicaCountDifference(desired, latest)
		assert.NotNil(desired.ko.Spec.NodeGroupConfiguration)
		for _, nodeGroupConfig := range desired.ko.Spec.NodeGroupConfiguration {
			require.NotNil(nodeGroupConfig.ReplicaCount)
			assert.Equal(desiredReplicaCount, int(*nodeGroupConfig.ReplicaCount))
		}
		assert.NotNil(latest.ko.Status.NodeGroups)
		for _, nodeGroup := range latest.ko.Status.NodeGroups {
			require.NotNil(nodeGroup.NodeGroupMembers)
			assert.Equal(latestReplicaCount+1, len(nodeGroup.NodeGroupMembers)) // replicas + primary node
		}
		assert.True(diff < 0) // desired replicas < latest replicas
	})
}

// TestNewIncreaseReplicaCountRequestPayload tests scenarios to
// check request payload by providing desired spec details  for increase replica count.
func TestNewIncreaseReplicaCountRequestPayload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("EmptyPayload=NoSpec", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		payload, err := rm.newIncreaseReplicaCountRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Nil(payload.ReplicationGroupId)
		assert.Nil(payload.NewReplicaCount)
		assert.Nil(payload.ReplicaConfiguration)
		assert.Nil(err)
	})
	t.Run("Payload=Spec", func(t *testing.T) {
		desired := provideResource()
		replicationGroupID := "test-rg"
		desired.ko.Spec.ReplicationGroupID = &replicationGroupID
		desiredReplicaCount := int64(2)
		desired.ko.Spec.ReplicasPerNodeGroup = &desiredReplicaCount
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(
			int(desiredReplicaCount), "1001", "1002")
		// expected: only node groups that are present on server are included in payload.
		expectedPayloadReplicaConfiguration := provideNodeGroupConfigurationWithReplicas(
			int(desiredReplicaCount), "1001")
		latest := provideResource()
		latestReplicaCount := int64(3)
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(
			int(latestReplicaCount), "1001", "1003")
		payload, err := rm.newIncreaseReplicaCountRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Equal(replicationGroupID, *payload.ReplicationGroupId)
		assert.Equal(desiredReplicaCount, *payload.NewReplicaCount)
		assert.NotNil(payload.ReplicaConfiguration)
		validatePayloadReplicaConfig(expectedPayloadReplicaConfiguration, payload.ReplicaConfiguration, assert, require)
		assert.Nil(err)
	})
}

// TestNewDecreaseReplicaCountRequestPayload tests scenarios to
// check request payload by providing desired spec details for decrease replica count.
func TestNewDecreaseReplicaCountRequestPayload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("EmptyPayload=NoSpec", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		payload, err := rm.newDecreaseReplicaCountRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Nil(payload.ReplicationGroupId)
		assert.Nil(payload.NewReplicaCount)
		assert.Nil(payload.ReplicaConfiguration)
		assert.Nil(err)
	})
	t.Run("Payload=Spec_Server", func(t *testing.T) {
		desired := provideResource()
		replicationGroupID := "test-rg"
		desired.ko.Spec.ReplicationGroupID = &replicationGroupID
		desiredReplicaCount := int64(2)
		desired.ko.Spec.ReplicasPerNodeGroup = &desiredReplicaCount
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfigurationWithReplicas(
			int(desiredReplicaCount), "1001", "1002")
		// expected: only node groups that are present on server are included in payload.
		expectedPayloadReplicaConfiguration := provideNodeGroupConfigurationWithReplicas(
			int(desiredReplicaCount), "1001")
		latest := provideResource()
		latestReplicaCount := int64(1)
		latest.ko.Status.NodeGroups = provideNodeGroupsWithReplicas(
			int(latestReplicaCount), "1001")
		payload, err := rm.newDecreaseReplicaCountRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Equal(replicationGroupID, *payload.ReplicationGroupId)
		assert.Equal(desiredReplicaCount, *payload.NewReplicaCount)
		assert.NotNil(payload.ReplicaConfiguration)
		validatePayloadReplicaConfig(expectedPayloadReplicaConfiguration, payload.ReplicaConfiguration, assert, require)
		assert.Nil(err)
	})
}

// TestShardConfigurationsDiffer tests scenarios to check if desired, latest shards
// configurations differ.
func TestShardConfigurationsDiffer(t *testing.T) {
	assert := assert.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("NoDiff=NoSpec_NoStatus", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.False(differ)
	})
	t.Run("NoDiff=NoSpec_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.False(differ)
	})
	t.Run("Diff=Spec.NumNodeGroups_NoStatus", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desiredShards := int64(2)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.True(differ)
	})
	t.Run("Diff=Spec.NodeGroupConfiguration_NoStatus", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.True(differ)
	})
	t.Run("NoDiff=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.False(differ)
	})
	t.Run("Diff=ScaleIn_Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001", "1002", "1003")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.True(differ)
	})
	t.Run("Diff=ScaleOut_Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.True(differ)
	})
	t.Run("NoDiff=Spec.NumNodeGroups_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desiredShards := int64(1)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.False(differ)
	})
	t.Run("Diff=Spec.NumNodeGroups_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desiredShards := int64(2)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.True(differ)
	})

	t.Run("NoDiff=Prefer_Spec.NumNodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desiredShards := int64(2)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002", "1003")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001", "1002")
		differ := rm.shardConfigurationsDiffer(desired, latest)
		assert.False(differ)
	})
}

// TestNewUpdateShardConfigurationRequestPayload tests scenarios to
// check request payload by providing desired, latest details
func TestNewUpdateShardConfigurationRequestPayload(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("EmptyPayload=NoSpec_NoStatus", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Nil(payload.ReplicationGroupId)
		assert.Nil(payload.NodeGroupCount)
		assert.Nil(payload.ReshardingConfiguration)
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("EmptyPayload=NoSpec_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Nil(payload.ReplicationGroupId)
		assert.Nil(payload.NodeGroupCount)
		assert.Nil(payload.ReshardingConfiguration)
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("ScaleOutPayload=Prefer_Spec.NumNodeGroups_NoStatus", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		desiredShards := int64(2)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002", "1003")
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		require.NotNil(payload.NodeGroupCount)
		assert.Equal(*desired.ko.Spec.NumNodeGroups, *payload.NodeGroupCount) // preferred NumNodeGroups over len(NodeGroupConfiguration)
		require.NotNil(payload.ReshardingConfiguration)                       // built as provided in desired object NodeGroupConfiguration
		assert.Equal(len(desired.ko.Spec.NodeGroupConfiguration), len(payload.ReshardingConfiguration))
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("ScaleOutPayload=Computed_Spec.NodeGroupConfiguration_NoStatus", func(t *testing.T) {
		desired := provideResource()
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002", "1003")
		latest := provideResource()
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		require.NotNil(payload.NodeGroupCount)
		assert.Equal(int64(len(desired.ko.Spec.NodeGroupConfiguration)), *payload.NodeGroupCount)
		require.NotNil(payload.ReshardingConfiguration) // increase scenario as no-status
		assert.Equal(len(desired.ko.Spec.NodeGroupConfiguration), len(payload.ReshardingConfiguration))
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("ScaleOutPayload=Prefer_Spec.NumNodeGroups_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		replicationGroupID := "test-rg"
		desired.ko.Spec.ReplicationGroupID = &replicationGroupID
		desiredShards := int64(2)
		desired.ko.Spec.NumNodeGroups = &desiredShards
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002", "1003")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Equal(*desired.ko.Spec.ReplicationGroupID, *payload.ReplicationGroupId)
		require.NotNil(payload.NodeGroupCount)
		assert.Equal(*desired.ko.Spec.NumNodeGroups, *payload.NodeGroupCount)
		validatePayloadReshardingConfig(desired.ko.Spec.NodeGroupConfiguration, payload.ReshardingConfiguration, assert, require)
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("ScaleOutPayload=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		replicationGroupID := "test-rg"
		desired.ko.Spec.ReplicationGroupID = &replicationGroupID
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001", "1002", "1003")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001")
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Equal(*desired.ko.Spec.ReplicationGroupID, *payload.ReplicationGroupId)
		require.NotNil(payload.NodeGroupCount)
		assert.Equal(int64(len(desired.ko.Spec.NodeGroupConfiguration)), *payload.NodeGroupCount)
		require.NotNil(payload.ReshardingConfiguration)
		validatePayloadReshardingConfig(desired.ko.Spec.NodeGroupConfiguration, payload.ReshardingConfiguration, assert, require)
		assert.Nil(payload.NodeGroupsToRetain)
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
	t.Run("ScaleInPayload=Spec.NodeGroupConfiguration_Status.NodeGroups", func(t *testing.T) {
		desired := provideResource()
		latest := provideResource()
		replicationGroupID := "test-rg"
		desired.ko.Spec.ReplicationGroupID = &replicationGroupID
		desired.ko.Spec.NodeGroupConfiguration = provideNodeGroupConfiguration("1001")
		latest.ko.Status.NodeGroups = provideNodeGroups("1001", "1002", "1003")
		payload, err := rm.newUpdateShardConfigurationRequestPayload(desired, latest)
		require.NotNil(payload)
		require.NotNil(payload.ApplyImmediately)
		assert.True(*payload.ApplyImmediately)
		assert.Equal(*desired.ko.Spec.ReplicationGroupID, *payload.ReplicationGroupId)
		require.NotNil(payload.NodeGroupCount)
		assert.Equal(int64(len(desired.ko.Spec.NodeGroupConfiguration)), *payload.NodeGroupCount)
		assert.Nil(payload.ReshardingConfiguration)
		require.NotNil(payload.NodeGroupsToRetain)
		assert.Equal(len(desired.ko.Spec.NodeGroupConfiguration), len(payload.NodeGroupsToRetain))
		for _, desiredNodeGroup := range desired.ko.Spec.NodeGroupConfiguration {
			found := false
			for _, nodeGroupId := range payload.NodeGroupsToRetain {
				if *desiredNodeGroup.NodeGroupID == *nodeGroupId {
					found = true
					break
				}
			}
			assert.True(found, "Expected node group id %s not found in payload", desiredNodeGroup.NodeGroupID)
		}
		assert.Nil(payload.NodeGroupsToRemove)
		assert.Nil(err)
	})
}

// TestSecurityGroupIdsDiffer tests scenarios to check if desired, latest (from cache cluster)
// security group ids configuration differs.
func TestSecurityGroupIdsDiffer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("NoDiff=NoSpec_NoStatus", func(t *testing.T) {
		desiredRG := provideResource()
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		require.Nil(desiredRG.ko.Spec.SecurityGroupIDs)
		require.Nil(latestCacheCluster.SecurityGroups)
		differ := rm.securityGroupIdsDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("NoDiff=NoSpec_HasStatus", func(t *testing.T) {
		desiredRG := provideResource()
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestCacheCluster.SecurityGroups = provideCacheClusterSecurityGroups("sg-001, sg-002")
		require.Nil(desiredRG.ko.Spec.SecurityGroupIDs)
		require.NotNil(latestCacheCluster.SecurityGroups)
		differ := rm.securityGroupIdsDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("NoDiff=Spec_Status_Match", func(t *testing.T) {
		desiredRG := provideResource()
		sg1 := "sg-001"
		sg2 := "sg-002"
		desiredRG.ko.Spec.SecurityGroupIDs = []*string{&sg1, &sg2}
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestCacheCluster.SecurityGroups = provideCacheClusterSecurityGroups(sg2, sg1) // same but out of order
		require.NotNil(desiredRG.ko.Spec.SecurityGroupIDs)
		require.NotNil(latestCacheCluster.SecurityGroups)
		differ := rm.securityGroupIdsDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("Diff=Spec_Status_MisMatch", func(t *testing.T) {
		desiredRG := provideResource()
		sg1 := "sg-001"
		sg2 := "sg-002"
		desiredRG.ko.Spec.SecurityGroupIDs = []*string{&sg1}
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestCacheCluster.SecurityGroups = provideCacheClusterSecurityGroups(sg2, sg1) // sg2 is additional
		require.NotNil(desiredRG.ko.Spec.SecurityGroupIDs)
		require.NotNil(latestCacheCluster.SecurityGroups)
		differ := rm.securityGroupIdsDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.True(differ)
	})
}

// provideNodeGroupsWithReplicas provides NodeGroups array for given node IDs
// each node group is populated with supplied numbers of replica nodes and a primary node.
func provideCacheClusterSecurityGroups(IDs ...string) []*svcsdk.SecurityGroupMembership {
	securityGroups := []*svcsdk.SecurityGroupMembership{}
	for _, ID := range IDs {
		securityGroupId := ID
		status := "available"
		securityGroups = append(securityGroups, &svcsdk.SecurityGroupMembership{
			SecurityGroupId: &securityGroupId,
			Status:          &status,
		})
	}
	return securityGroups
}

// TestEngineVersionDiffer tests scenarios to check if desired, latest (from cache cluster)
// Engine Version configuration differs.
func TestEngineVersionDiffer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// setup
	rm := provideResourceManager()
	// Tests
	t.Run("NoDiff=NoSpec_NoStatus", func(t *testing.T) {
		desiredRG := provideResource()
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		require.Nil(desiredRG.ko.Spec.SecurityGroupIDs)
		require.Nil(latestCacheCluster.SecurityGroups)
		differ := rm.engineVersionDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("NoDiff=NoSpec_HasStatus", func(t *testing.T) {
		desiredRG := provideResource()
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestEV := "test-engine-version"
		latestCacheCluster.EngineVersion = &latestEV
		require.Nil(desiredRG.ko.Spec.EngineVersion)
		require.NotNil(latestCacheCluster.EngineVersion)
		differ := rm.engineVersionDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("NoDiff=Spec_Status_Match", func(t *testing.T) {
		desiredRG := provideResource()
		latestEV := "test-engine-version"
		desiredRG.ko.Spec.EngineVersion = &latestEV
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestCacheCluster.EngineVersion = &latestEV
		require.NotNil(desiredRG.ko.Spec.EngineVersion)
		require.NotNil(latestCacheCluster.EngineVersion)
		differ := rm.engineVersionDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.False(differ)
	})
	t.Run("Diff=Spec_Status_MisMatch", func(t *testing.T) {
		desiredRG := provideResource()
		desiredEV := "desired-test-engine-version"
		desiredRG.ko.Spec.EngineVersion = &desiredEV
		latestRG := provideResource()
		latestCacheCluster := provideCacheCluster()
		latestEV := "latest-test-engine-version"
		latestCacheCluster.EngineVersion = &latestEV

		require.NotNil(desiredRG.ko.Spec.EngineVersion)
		require.NotNil(latestCacheCluster.EngineVersion)
		require.NotEqual(*desiredRG.ko.Spec.EngineVersion, *latestCacheCluster.EngineVersion)
		differ := rm.engineVersionDiffer(desiredRG, latestRG, latestCacheCluster)
		assert.True(differ)
	})
}
