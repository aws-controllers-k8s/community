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
	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func (rm *resourceManager) UpdateShardConfiguration(
	ctx context.Context,
	r *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	input, err := rm.newUpdateShardConfigurationRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, respErr := rm.sdkapi.ModifyReplicationGroupShardConfigurationWithContext(ctx, input)
	if respErr != nil {
		fmt.Printf("Failed to ModifyReplicationGroupShardConfigurationWithContext, Error: %v\n", respErr)
		return nil, respErr
	}
	return provideUpdatedResource(r, resp.ReplicationGroup)
}

func (rm *resourceManager) UpdateReplicaCount(
	ctx context.Context,
	r *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {

	for _, diff := range diffReporter.Differences {
		if diff.Path == "Spec.ReplicasPerNodeGroup" {
			desired, err1 := strconv.Atoi(diff.ValueA)
			latest, err2 := strconv.Atoi(diff.ValueB)

			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("UpdateReplicaCount failed: invalid values")
			}

			if latest < desired { // increase
				fmt.Printf("Requesting Increase Replica Count. Old value: %v, New Value: %v\n", latest, desired)
				input, err := rm.newIncreaseReplicaCountRequestPayload(r)
				if err != nil {
					fmt.Printf("Error occurred: %v\n", err)
					return nil, err
				}
				resp, respErr := rm.sdkapi.IncreaseReplicaCountWithContext(ctx, input)
				if respErr != nil {
					fmt.Printf("Failed to IncreaseReplicaCountWithContext. Error: %v\n", respErr)
					return nil, respErr
				}
				return provideUpdatedResource(r, resp.ReplicationGroup)
			} else { // decrease
				input, err := rm.newDecreaseReplicaCountRequestPayload(r)
				fmt.Printf("Requesting Decrease Replica Count. Old value: %v, New Value: %v\n", latest, desired)
				if err != nil {
					fmt.Printf("Error occurred: %v\n", err)
					return nil, err
				}
				resp, respErr := rm.sdkapi.DecreaseReplicaCountWithContext(ctx, input)
				if respErr != nil {
					fmt.Printf("Failed to DecreaseReplicaCountWithContext. Error: %v\n", respErr)
					return nil, respErr
				}
				return provideUpdatedResource(r, resp.ReplicationGroup)
			}

			break
		}
	}
	return nil, fmt.Errorf("UpdateReplicaCount failed")
}

// newUpdate(ShardConfiguration)RequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateShardConfigurationRequestPayload(
	r *resource,
) (*svcsdk.ModifyReplicationGroupShardConfigurationInput, error) {
	res := &svcsdk.ModifyReplicationGroupShardConfigurationInput{}

	res.SetApplyImmediately(true)
	if r.ko.Spec.ReplicationGroupID != nil {
		res.SetReplicationGroupId(*r.ko.Spec.ReplicationGroupID)
	}
	if r.ko.Spec.NumNodeGroups != nil {
		res.SetNodeGroupCount(*r.ko.Spec.NumNodeGroups)
	}

	nodegroupsToRetain := []*string{}

	// TODO: optional -only if- NumNodeGroups increases shards
	if r.ko.Spec.NodeGroupConfiguration != nil {
		f13 := []*svcsdk.ReshardingConfiguration{}
		for _, f13iter := range r.ko.Spec.NodeGroupConfiguration {
			f13elem := &svcsdk.ReshardingConfiguration{}
			if f13iter.NodeGroupID != nil {
				f13elem.SetNodeGroupId(*f13iter.NodeGroupID)
				nodegroupsToRetain = append(nodegroupsToRetain, &(*f13iter.NodeGroupID))
			}
			f13elemf2 := []*string{}
			if f13iter.PrimaryAvailabilityZone != nil {
				f13elemf2 = append(f13elemf2, &(*f13iter.PrimaryAvailabilityZone))
			}
			if f13iter.ReplicaAvailabilityZones != nil {
				for _, f13elemf2iter := range f13iter.ReplicaAvailabilityZones {
					var f13elemf2elem string
					f13elemf2elem = *f13elemf2iter
					f13elemf2 = append(f13elemf2, &f13elemf2elem)
				}
				f13elem.SetPreferredAvailabilityZones(f13elemf2)
			}
			f13 = append(f13, f13elem)
		}
		res.SetReshardingConfiguration(f13)
	}

	// TODO: optional - only if -  NumNodeGroups decreases shards
	// res.SetNodeGroupsToRemove() or res.SetNodeGroupsToRetain()
	res.SetNodeGroupsToRetain(nodegroupsToRetain)

	return res, nil
}

// new(IncreaseReplicaCount)RequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newIncreaseReplicaCountRequestPayload(
	r *resource,
) (*svcsdk.IncreaseReplicaCountInput, error) {
	res := &svcsdk.IncreaseReplicaCountInput{}

	res.SetApplyImmediately(true)
	if r.ko.Spec.ReplicationGroupID != nil {
		res.SetReplicationGroupId(*r.ko.Spec.ReplicationGroupID)
	}
	if r.ko.Spec.NumNodeGroups != nil {
		res.SetNewReplicaCount(*r.ko.Spec.ReplicasPerNodeGroup)
	}

	if r.ko.Spec.NodeGroupConfiguration != nil {
		f13 := []*svcsdk.ConfigureShard{}
		for _, f13iter := range r.ko.Spec.NodeGroupConfiguration {
			f13elem := &svcsdk.ConfigureShard{}
			if f13iter.NodeGroupID != nil {
				f13elem.SetNodeGroupId(*f13iter.NodeGroupID)
			}
			if f13iter.ReplicaCount != nil {
				f13elem.SetNewReplicaCount(*f13iter.ReplicaCount)
			}
			f13elemf2 := []*string{}
			if f13iter.PrimaryAvailabilityZone != nil {
				f13elemf2 = append(f13elemf2, &(*f13iter.PrimaryAvailabilityZone))
			}
			if f13iter.ReplicaAvailabilityZones != nil {
				for _, f13elemf2iter := range f13iter.ReplicaAvailabilityZones {
					var f13elemf2elem string
					f13elemf2elem = *f13elemf2iter
					f13elemf2 = append(f13elemf2, &f13elemf2elem)
				}
				f13elem.SetPreferredAvailabilityZones(f13elemf2)
			}
			f13 = append(f13, f13elem)
		}
		res.SetReplicaConfiguration(f13)
	}

	return res, nil
}

// new(DecreaseReplicaCount)RequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newDecreaseReplicaCountRequestPayload(
	r *resource,
) (*svcsdk.DecreaseReplicaCountInput, error) {
	res := &svcsdk.DecreaseReplicaCountInput{}

	res.SetApplyImmediately(true)
	if r.ko.Spec.ReplicationGroupID != nil {
		res.SetReplicationGroupId(*r.ko.Spec.ReplicationGroupID)
	}
	if r.ko.Spec.NumNodeGroups != nil {
		res.SetNewReplicaCount(*r.ko.Spec.ReplicasPerNodeGroup)
	}

	if r.ko.Spec.NodeGroupConfiguration != nil {
		f13 := []*svcsdk.ConfigureShard{}
		for _, f13iter := range r.ko.Spec.NodeGroupConfiguration {
			f13elem := &svcsdk.ConfigureShard{}
			if f13iter.NodeGroupID != nil {
				f13elem.SetNodeGroupId(*f13iter.NodeGroupID)
			}
			if f13iter.ReplicaCount != nil {
				f13elem.SetNewReplicaCount(*f13iter.ReplicaCount)
			}
			f13elemf2 := []*string{}
			if f13iter.PrimaryAvailabilityZone != nil {
				f13elemf2 = append(f13elemf2, &(*f13iter.PrimaryAvailabilityZone))
			}
			if f13iter.ReplicaAvailabilityZones != nil {
				for _, f13elemf2iter := range f13iter.ReplicaAvailabilityZones {
					var f13elemf2elem string
					f13elemf2elem = *f13elemf2iter
					f13elemf2 = append(f13elemf2, &f13elemf2elem)
				}
				f13elem.SetPreferredAvailabilityZones(f13elemf2)
			}
			f13 = append(f13, f13elem)
		}
		res.SetReplicaConfiguration(f13)
	}

	return res, nil
}

// This method copies the data from given replicationGroup by populating it into copy of supplied resource
// and returns it.
func provideUpdatedResource(
	r *resource,
	replicationGroup *svcsdk.ReplicationGroup,
) (*resource, error) {
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	if replicationGroup.AuthTokenEnabled != nil {
		ko.Status.AuthTokenEnabled = replicationGroup.AuthTokenEnabled
	}
	if replicationGroup.AuthTokenLastModifiedDate != nil {
		ko.Status.AuthTokenLastModifiedDate = &metav1.Time{*replicationGroup.AuthTokenLastModifiedDate}
	}
	if replicationGroup.AutomaticFailover != nil {
		ko.Status.AutomaticFailover = replicationGroup.AutomaticFailover
	}
	if replicationGroup.ClusterEnabled != nil {
		ko.Status.ClusterEnabled = replicationGroup.ClusterEnabled
	}
	if replicationGroup.ConfigurationEndpoint != nil {
		f7 := &svcapitypes.Endpoint{}
		if replicationGroup.ConfigurationEndpoint.Address != nil {
			f7.Address = replicationGroup.ConfigurationEndpoint.Address
		}
		if replicationGroup.ConfigurationEndpoint.Port != nil {
			f7.Port = replicationGroup.ConfigurationEndpoint.Port
		}
		ko.Status.ConfigurationEndpoint = f7
	}
	if replicationGroup.Description != nil {
		ko.Status.Description = replicationGroup.Description
	}
	if replicationGroup.GlobalReplicationGroupInfo != nil {
		f9 := &svcapitypes.GlobalReplicationGroupInfo{}
		if replicationGroup.GlobalReplicationGroupInfo.GlobalReplicationGroupId != nil {
			f9.GlobalReplicationGroupID = replicationGroup.GlobalReplicationGroupInfo.GlobalReplicationGroupId
		}
		if replicationGroup.GlobalReplicationGroupInfo.GlobalReplicationGroupMemberRole != nil {
			f9.GlobalReplicationGroupMemberRole = replicationGroup.GlobalReplicationGroupInfo.GlobalReplicationGroupMemberRole
		}
		ko.Status.GlobalReplicationGroupInfo = f9
	}
	if replicationGroup.MemberClusters != nil {
		f11 := []*string{}
		for _, f11iter := range replicationGroup.MemberClusters {
			var f11elem string
			f11elem = *f11iter
			f11 = append(f11, &f11elem)
		}
		ko.Status.MemberClusters = f11
	}
	if replicationGroup.MultiAZ != nil {
		ko.Status.MultiAZ = replicationGroup.MultiAZ
	}
	if replicationGroup.NodeGroups != nil {
		f13 := []*svcapitypes.NodeGroup{}
		for _, f13iter := range replicationGroup.NodeGroups {
			f13elem := &svcapitypes.NodeGroup{}
			if f13iter.NodeGroupId != nil {
				f13elem.NodeGroupID = f13iter.NodeGroupId
			}
			if f13iter.NodeGroupMembers != nil {
				f13elemf1 := []*svcapitypes.NodeGroupMember{}
				for _, f13elemf1iter := range f13iter.NodeGroupMembers {
					f13elemf1elem := &svcapitypes.NodeGroupMember{}
					if f13elemf1iter.CacheClusterId != nil {
						f13elemf1elem.CacheClusterID = f13elemf1iter.CacheClusterId
					}
					if f13elemf1iter.CacheNodeId != nil {
						f13elemf1elem.CacheNodeID = f13elemf1iter.CacheNodeId
					}
					if f13elemf1iter.CurrentRole != nil {
						f13elemf1elem.CurrentRole = f13elemf1iter.CurrentRole
					}
					if f13elemf1iter.PreferredAvailabilityZone != nil {
						f13elemf1elem.PreferredAvailabilityZone = f13elemf1iter.PreferredAvailabilityZone
					}
					if f13elemf1iter.ReadEndpoint != nil {
						f13elemf1elemf4 := &svcapitypes.Endpoint{}
						if f13elemf1iter.ReadEndpoint.Address != nil {
							f13elemf1elemf4.Address = f13elemf1iter.ReadEndpoint.Address
						}
						if f13elemf1iter.ReadEndpoint.Port != nil {
							f13elemf1elemf4.Port = f13elemf1iter.ReadEndpoint.Port
						}
						f13elemf1elem.ReadEndpoint = f13elemf1elemf4
					}
					f13elemf1 = append(f13elemf1, f13elemf1elem)
				}
				f13elem.NodeGroupMembers = f13elemf1
			}
			if f13iter.PrimaryEndpoint != nil {
				f13elemf2 := &svcapitypes.Endpoint{}
				if f13iter.PrimaryEndpoint.Address != nil {
					f13elemf2.Address = f13iter.PrimaryEndpoint.Address
				}
				if f13iter.PrimaryEndpoint.Port != nil {
					f13elemf2.Port = f13iter.PrimaryEndpoint.Port
				}
				f13elem.PrimaryEndpoint = f13elemf2
			}
			if f13iter.ReaderEndpoint != nil {
				f13elemf3 := &svcapitypes.Endpoint{}
				if f13iter.ReaderEndpoint.Address != nil {
					f13elemf3.Address = f13iter.ReaderEndpoint.Address
				}
				if f13iter.ReaderEndpoint.Port != nil {
					f13elemf3.Port = f13iter.ReaderEndpoint.Port
				}
				f13elem.ReaderEndpoint = f13elemf3
			}
			if f13iter.Slots != nil {
				f13elem.Slots = f13iter.Slots
			}
			if f13iter.Status != nil {
				f13elem.Status = f13iter.Status
			}
			f13 = append(f13, f13elem)
		}
		ko.Status.NodeGroups = f13
	}
	if replicationGroup.PendingModifiedValues != nil {
		f14 := &svcapitypes.ReplicationGroupPendingModifiedValues{}
		if replicationGroup.PendingModifiedValues.AuthTokenStatus != nil {
			f14.AuthTokenStatus = replicationGroup.PendingModifiedValues.AuthTokenStatus
		}
		if replicationGroup.PendingModifiedValues.AutomaticFailoverStatus != nil {
			f14.AutomaticFailoverStatus = replicationGroup.PendingModifiedValues.AutomaticFailoverStatus
		}
		if replicationGroup.PendingModifiedValues.PrimaryClusterId != nil {
			f14.PrimaryClusterID = replicationGroup.PendingModifiedValues.PrimaryClusterId
		}
		if replicationGroup.PendingModifiedValues.Resharding != nil {
			f14f3 := &svcapitypes.ReshardingStatus{}
			if replicationGroup.PendingModifiedValues.Resharding.SlotMigration != nil {
				f14f3f0 := &svcapitypes.SlotMigration{}
				if replicationGroup.PendingModifiedValues.Resharding.SlotMigration.ProgressPercentage != nil {
					f14f3f0.ProgressPercentage = replicationGroup.PendingModifiedValues.Resharding.SlotMigration.ProgressPercentage
				}
				f14f3.SlotMigration = f14f3f0
			}
			f14.Resharding = f14f3
		}
		ko.Status.PendingModifiedValues = f14
	}
	if replicationGroup.SnapshottingClusterId != nil {
		ko.Status.SnapshottingClusterID = replicationGroup.SnapshottingClusterId
	}
	if replicationGroup.Status != nil {
		ko.Status.Status = replicationGroup.Status
	}

	return &resource{ko}, nil
}
