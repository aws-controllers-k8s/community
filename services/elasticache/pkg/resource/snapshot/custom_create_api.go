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

package snapshot

import (
	"context"
	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	svcapitypes "github.com/aws/aws-controllers-k8s/services/elasticache/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (rm *resourceManager) CustomCreateSnapshot(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	if r.ko.Spec.SourceSnapshotName != nil {
		if r.ko.Spec.CacheClusterID != nil || r.ko.Spec.ReplicationGroupID != nil {
			return nil, awserr.New("InvalidParameterCombination", "Cannot specify CacheClusteId or "+
				"ReplicationGroupId while SourceSnapshotName is specified", nil)
		}

		input, err := rm.newCopySnapshotPayload(r)
		if err != nil {
			return nil, err
		}

		resp, respErr := rm.sdkapi.CopySnapshot(input)

		rm.metrics.RecordAPICall("CREATE", "CopySnapshot", respErr)
		if respErr != nil {
			return nil, respErr
		}
		// Merge in the information we read from the API call above to the copy of
		// the original Kubernetes object we passed to the function
		ko := r.ko.DeepCopy()

		if ko.Status.ACKResourceMetadata == nil {
			ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
		}
		if resp.Snapshot.ARN != nil {
			arn := ackv1alpha1.AWSResourceName(*resp.Snapshot.ARN)
			ko.Status.ACKResourceMetadata.ARN = &arn
		}
		if resp.Snapshot.AutoMinorVersionUpgrade != nil {
			ko.Status.AutoMinorVersionUpgrade = resp.Snapshot.AutoMinorVersionUpgrade
		}
		if resp.Snapshot.AutomaticFailover != nil {
			ko.Status.AutomaticFailover = resp.Snapshot.AutomaticFailover
		}
		if resp.Snapshot.CacheClusterCreateTime != nil {
			ko.Status.CacheClusterCreateTime = &metav1.Time{*resp.Snapshot.CacheClusterCreateTime}
		}
		if resp.Snapshot.CacheNodeType != nil {
			ko.Status.CacheNodeType = resp.Snapshot.CacheNodeType
		}
		if resp.Snapshot.CacheParameterGroupName != nil {
			ko.Status.CacheParameterGroupName = resp.Snapshot.CacheParameterGroupName
		}
		if resp.Snapshot.CacheSubnetGroupName != nil {
			ko.Status.CacheSubnetGroupName = resp.Snapshot.CacheSubnetGroupName
		}
		if resp.Snapshot.Engine != nil {
			ko.Status.Engine = resp.Snapshot.Engine
		}
		if resp.Snapshot.EngineVersion != nil {
			ko.Status.EngineVersion = resp.Snapshot.EngineVersion
		}
		if resp.Snapshot.NodeSnapshots != nil {
			f11 := []*svcapitypes.NodeSnapshot{}
			for _, f11iter := range resp.Snapshot.NodeSnapshots {
				f11elem := &svcapitypes.NodeSnapshot{}
				if f11iter.CacheClusterId != nil {
					f11elem.CacheClusterID = f11iter.CacheClusterId
				}
				if f11iter.CacheNodeCreateTime != nil {
					f11elem.CacheNodeCreateTime = &metav1.Time{*f11iter.CacheNodeCreateTime}
				}
				if f11iter.CacheNodeId != nil {
					f11elem.CacheNodeID = f11iter.CacheNodeId
				}
				if f11iter.CacheSize != nil {
					f11elem.CacheSize = f11iter.CacheSize
				}
				if f11iter.NodeGroupConfiguration != nil {
					f11elemf4 := &svcapitypes.NodeGroupConfiguration{}
					if f11iter.NodeGroupConfiguration.NodeGroupId != nil {
						f11elemf4.NodeGroupID = f11iter.NodeGroupConfiguration.NodeGroupId
					}
					if f11iter.NodeGroupConfiguration.PrimaryAvailabilityZone != nil {
						f11elemf4.PrimaryAvailabilityZone = f11iter.NodeGroupConfiguration.PrimaryAvailabilityZone
					}
					if f11iter.NodeGroupConfiguration.ReplicaAvailabilityZones != nil {
						f11elemf4f2 := []*string{}
						for _, f11elemf4f2iter := range f11iter.NodeGroupConfiguration.ReplicaAvailabilityZones {
							var f11elemf4f2elem string
							f11elemf4f2elem = *f11elemf4f2iter
							f11elemf4f2 = append(f11elemf4f2, &f11elemf4f2elem)
						}
						f11elemf4.ReplicaAvailabilityZones = f11elemf4f2
					}
					if f11iter.NodeGroupConfiguration.ReplicaCount != nil {
						f11elemf4.ReplicaCount = f11iter.NodeGroupConfiguration.ReplicaCount
					}
					if f11iter.NodeGroupConfiguration.Slots != nil {
						f11elemf4.Slots = f11iter.NodeGroupConfiguration.Slots
					}
					f11elem.NodeGroupConfiguration = f11elemf4
				}
				if f11iter.NodeGroupId != nil {
					f11elem.NodeGroupID = f11iter.NodeGroupId
				}
				if f11iter.SnapshotCreateTime != nil {
					f11elem.SnapshotCreateTime = &metav1.Time{*f11iter.SnapshotCreateTime}
				}
				f11 = append(f11, f11elem)
			}
			ko.Status.NodeSnapshots = f11
		}
		if resp.Snapshot.NumCacheNodes != nil {
			ko.Status.NumCacheNodes = resp.Snapshot.NumCacheNodes
		}
		if resp.Snapshot.NumNodeGroups != nil {
			ko.Status.NumNodeGroups = resp.Snapshot.NumNodeGroups
		}
		if resp.Snapshot.Port != nil {
			ko.Status.Port = resp.Snapshot.Port
		}
		if resp.Snapshot.PreferredAvailabilityZone != nil {
			ko.Status.PreferredAvailabilityZone = resp.Snapshot.PreferredAvailabilityZone
		}
		if resp.Snapshot.PreferredMaintenanceWindow != nil {
			ko.Status.PreferredMaintenanceWindow = resp.Snapshot.PreferredMaintenanceWindow
		}
		if resp.Snapshot.ReplicationGroupDescription != nil {
			ko.Status.ReplicationGroupDescription = resp.Snapshot.ReplicationGroupDescription
		}

		if resp.Snapshot.SnapshotRetentionLimit != nil {
			ko.Status.SnapshotRetentionLimit = resp.Snapshot.SnapshotRetentionLimit
		}
		if resp.Snapshot.SnapshotSource != nil {
			ko.Status.SnapshotSource = resp.Snapshot.SnapshotSource
		}
		if resp.Snapshot.SnapshotStatus != nil {
			ko.Status.SnapshotStatus = resp.Snapshot.SnapshotStatus
		}
		if resp.Snapshot.SnapshotWindow != nil {
			ko.Status.SnapshotWindow = resp.Snapshot.SnapshotWindow
		}
		if resp.Snapshot.TopicArn != nil {
			ko.Status.TopicARN = resp.Snapshot.TopicArn
		}
		if resp.Snapshot.VpcId != nil {
			ko.Status.VPCID = resp.Snapshot.VpcId
		}

		rm.setStatusDefaults(ko)
		// custom set output from response
		rm.CustomCopySnapshotSetOutput(r, resp, ko)
		return &resource{ko}, nil
	}

	return nil, nil
}

// newCopySnapshotPayload returns an SDK-specific struct for the HTTP request
// payload of the CopySnapshot API call
func (rm *resourceManager) newCopySnapshotPayload(
	r *resource,
) (*svcsdk.CopySnapshotInput, error) {
	res := &svcsdk.CopySnapshotInput{}

	if r.ko.Spec.SourceSnapshotName != nil {
		res.SetSourceSnapshotName(*r.ko.Spec.SourceSnapshotName)
	}
	if r.ko.Spec.KMSKeyID != nil {
		res.SetKmsKeyId(*r.ko.Spec.KMSKeyID)
	}

	if r.ko.Spec.SnapshotName != nil {
		res.SetTargetSnapshotName(*r.ko.Spec.SnapshotName)
	}

	return res, nil
}
