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

package model_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-controllers-k8s/pkg/model"
	"github.com/aws/aws-controllers-k8s/pkg/testutil"
)

func attrCamelNames(fields map[string]*model.CRDField) []string {
	res := []string{}
	for _, attr := range fields {
		res = append(res, attr.Names.Camel)
	}
	sort.Strings(res)
	return res
}

func getCRDByName(name string, crds []*model.CRD) *model.CRD {
	for _, c := range crds {
		if c.Names.Original == name {
			return c
		}
	}
	return nil
}

func TestSNSTopic(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "sns")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Topic", crds)
	require.NotNil(crd)

	assert.Equal("Topic", crd.Names.Camel)
	assert.Equal("topic", crd.Names.CamelLower)
	assert.Equal("topic", crd.Names.Snake)

	// The SNS Topic API is a little weird. There are Create and Delete
	// operations ("CreateTopic", "DeleteTopic") but there is no ReadOne
	// operation (there is a "GetTopicAttributes" call though) or Update
	// operation (there is a "SetTopicAttributes" call though). And there is a
	// ReadMany operation (ListTopics)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.GetAttributes)
	assert.NotNil(crd.Ops.SetAttributes)

	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	// The SNS Topic uses an "Attributes" map[string]*string to masquerade
	// real fields. DeliveryPolicy, Policy, KMSMasterKeyID and DisplayName are
	// all examples of these masqueraded fields...
	expSpecFieldCamel := []string{
		"DeliveryPolicy",
		"DisplayName",
		"KMSMasterKeyID",
		"Name",
		"Policy",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// The SNS Topic uses an "Attributes" map[string]*string to masquerade
	// real fields. EffectiveDeliveryPolicy and Owner are
	// examples of these masqueraded fields that are ReadOnly and thus belong
	// in the Status struct
	expStatusFieldCamel := []string{
		// "TopicARN" is in the output shape for CreateTopic, but it is removed
		// because it is the primary resource object's ARN field and the
		// SDKMapper has identified it as the source for the standard
		// Status.ACKResourceMetadata.ARN field
		"EffectiveDeliveryPolicy",
		"Owner",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// The input shape for the Create operation is set from a variety of scalar
	// and non-scalar types and the SNS API features an Attributes parameter
	// that is actually a map[string]*string of real field values that are
	// unpacked by the code generator.
	expCreateInput := `
	attrMap := map[string]*string{}
	attrMap["DeliveryPolicy"] = r.ko.Spec.DeliveryPolicy
	attrMap["DisplayName"] = r.ko.Spec.DisplayName
	attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
	attrMap["Policy"] = r.ko.Spec.Policy
	res.SetAttributes(attrMap)
	res.SetName(*r.ko.Spec.Name)
	f5 := []*svcsdk.Tag{}
	for _, f5iter := range r.ko.Spec.Tags {
		f5elem := &svcsdk.Tag{}
		f5elem.SetKey(*f5iter.Key)
		f5elem.SetValue(*f5iter.Value)
		f5 = append(f5, f5elem)
	}
	res.SetTags(f5)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko.Spec", "res", 1))

	// None of the fields in the Topic resource's CreateTopicInput shape are
	// returned in the CreateTopicOutput shape, so none of them return any Go
	// code for setting a Status struct field to a corresponding Create Output
	// Shape member
	expCreateOutput := `
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The input shape for the GetAttributes operation has a single TopicArn
	// field. This field represents the ARN of the primary resource (the Topic
	// itself) and should be set specially from the ACKResourceMetadata.ARN
	// field in the TopicStatus struct
	expGetAttrsInput := `
	res.SetTopicArn(string(*r.ko.Status.ACKResourceMetadata.ARN))
`
	assert.Equal(expGetAttrsInput, crd.GoCodeGetAttributesSetInput("r.ko", "res", 1))

	// The output shape for the GetAttributes operation contains a single field
	// "Attributes" that must be unpacked into the Topic CRD's Status fields.
	// There are only three attribute keys that are *not* in the Input shape
	// (and thus in the Spec fields). Two of them are the tesource's ARN and
	// AWS Owner account ID, both of which are handled specially.
	expGetAttrsOutput := `
	ko.Status.EffectiveDeliveryPolicy = resp.Attributes["EffectiveDeliveryPolicy"]
	tmpOwnerID := ackv1alpha1.AWSAccountID(*resp.Attributes["Owner"])
	ko.Status.ACKResourceMetadata.OwnerAccountID = &tmpOwnerID
	tmpARN := ackv1alpha1.AWSResourceName(*resp.Attributes["TopicArn"])
	ko.Status.ACKResourceMetadata.ARN = &tmpARN
`
	assert.Equal(expGetAttrsOutput, crd.GoCodeGetAttributesSetOutput("resp", "ko.Status", 1))
}

func TestEC2LaunchTemplate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "ec2")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("LaunchTemplate", crds)
	require.NotNil(crd)

	assert.Equal("LaunchTemplate", crd.Names.Camel)
	assert.Equal("launchTemplate", crd.Names.CamelLower)
	assert.Equal("launch_template", crd.Names.Snake)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		// TODO(jaypipes): DryRun and ClientToken are examples of two fields in
		// the resource input shape that need to be stripped out of the CRD. We
		// need to instruct the code generator that these types of fields are
		// not germane to the resource itself...
		"ClientToken",
		"DryRun",
		"LaunchTemplateData",
		"LaunchTemplateName",
		// TODO(jaypipes): Here's an example of where we need to instruct the
		// code generator to rename the "TagSpecifications" field to simply
		// "Tags" and place it into the common Spec.Tags field.
		"TagSpecifications",
		"VersionDescription",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"CreateTime",
		"CreatedBy",
		"DefaultVersionNumber",
		"LatestVersionNumber",
		// TODO(jaypipes): Handle "Id" Fields like "LaunchTemplateId" in the
		// same way as we handle ARN-ified modern service APIs and use the
		// SDKMapper to instruct the code generator that this field represents
		// the primary resource object's identifier field.
		"LaunchTemplateID",
		// LaunchTemplateName excluded because it matches input shape.,
		// TODO(jaypipes): Tags field should be excluded because it is the same
		// as the input shape's "TagSpecifications" field...
		"Tags",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// LaunchTemplateName is in the LaunchTemplate resource's CreateTopicInput shape and also
	// returned in the CreateLaunchTemplateResult shape, so it should have
	// Go code to set the Input Shape member from the Spec field but not set a
	// Status field from the Create Output Shape member
	expCreateInput := `
	res.SetClientToken(*r.ko.Spec.ClientToken)
	res.SetDryRun(*r.ko.Spec.DryRun)
	f2 := &svcsdk.RequestLaunchTemplateData{}
	f2f0 := []*svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
	for _, f2f0iter := range r.ko.Spec.LaunchTemplateData.BlockDeviceMappings {
		f2f0elem := &svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
		f2f0elem.SetDeviceName(*f2f0iter.DeviceName)
		f2f0elemf1 := &svcsdk.LaunchTemplateEbsBlockDeviceRequest{}
		f2f0elemf1.SetDeleteOnTermination(*f2f0iter.EBS.DeleteOnTermination)
		f2f0elemf1.SetEncrypted(*f2f0iter.EBS.Encrypted)
		f2f0elemf1.SetIops(*f2f0iter.EBS.IOPS)
		f2f0elemf1.SetKmsKeyId(*f2f0iter.EBS.KMSKeyID)
		f2f0elemf1.SetSnapshotId(*f2f0iter.EBS.SnapshotID)
		f2f0elemf1.SetVolumeSize(*f2f0iter.EBS.VolumeSize)
		f2f0elemf1.SetVolumeType(*f2f0iter.EBS.VolumeType)
		f2f0elem.SetEbs(f2f0elemf1)
		f2f0elem.SetNoDevice(*f2f0iter.NoDevice)
		f2f0elem.SetVirtualName(*f2f0iter.VirtualName)
		f2f0 = append(f2f0, f2f0elem)
	}
	f2.SetBlockDeviceMappings(f2f0)
	f2f1 := &svcsdk.LaunchTemplateCapacityReservationSpecificationRequest{}
	f2f1.SetCapacityReservationPreference(*r.ko.Spec.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationPreference)
	f2f1f1 := &svcsdk.CapacityReservationTarget{}
	f2f1f1.SetCapacityReservationId(*r.ko.Spec.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget.CapacityReservationID)
	f2f1.SetCapacityReservationTarget(f2f1f1)
	f2.SetCapacityReservationSpecification(f2f1)
	f2f2 := &svcsdk.LaunchTemplateCpuOptionsRequest{}
	f2f2.SetCoreCount(*r.ko.Spec.LaunchTemplateData.CPUOptions.CoreCount)
	f2f2.SetThreadsPerCore(*r.ko.Spec.LaunchTemplateData.CPUOptions.ThreadsPerCore)
	f2.SetCpuOptions(f2f2)
	f2f3 := &svcsdk.CreditSpecificationRequest{}
	f2f3.SetCpuCredits(*r.ko.Spec.LaunchTemplateData.CreditSpecification.CPUCredits)
	f2.SetCreditSpecification(f2f3)
	f2.SetDisableApiTermination(*r.ko.Spec.LaunchTemplateData.DisableAPITermination)
	f2.SetEbsOptimized(*r.ko.Spec.LaunchTemplateData.EBSOptimized)
	f2f6 := []*svcsdk.ElasticGpuSpecification{}
	for _, f2f6iter := range r.ko.Spec.LaunchTemplateData.ElasticGPUSpecifications {
		f2f6elem := &svcsdk.ElasticGpuSpecification{}
		f2f6elem.SetType(*f2f6iter.Type)
		f2f6 = append(f2f6, f2f6elem)
	}
	f2.SetElasticGpuSpecifications(f2f6)
	f2f7 := []*svcsdk.LaunchTemplateElasticInferenceAccelerator{}
	for _, f2f7iter := range r.ko.Spec.LaunchTemplateData.ElasticInferenceAccelerators {
		f2f7elem := &svcsdk.LaunchTemplateElasticInferenceAccelerator{}
		f2f7elem.SetCount(*f2f7iter.Count)
		f2f7elem.SetType(*f2f7iter.Type)
		f2f7 = append(f2f7, f2f7elem)
	}
	f2.SetElasticInferenceAccelerators(f2f7)
	f2f8 := &svcsdk.LaunchTemplateHibernationOptionsRequest{}
	f2f8.SetConfigured(*r.ko.Spec.LaunchTemplateData.HibernationOptions.Configured)
	f2.SetHibernationOptions(f2f8)
	f2f9 := &svcsdk.LaunchTemplateIamInstanceProfileSpecificationRequest{}
	f2f9.SetArn(*r.ko.Spec.LaunchTemplateData.IAMInstanceProfile.ARN)
	f2f9.SetName(*r.ko.Spec.LaunchTemplateData.IAMInstanceProfile.Name)
	f2.SetIamInstanceProfile(f2f9)
	f2.SetImageId(*r.ko.Spec.LaunchTemplateData.ImageID)
	f2.SetInstanceInitiatedShutdownBehavior(*r.ko.Spec.LaunchTemplateData.InstanceInitiatedShutdownBehavior)
	f2f12 := &svcsdk.LaunchTemplateInstanceMarketOptionsRequest{}
	f2f12.SetMarketType(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.MarketType)
	f2f12f1 := &svcsdk.LaunchTemplateSpotMarketOptionsRequest{}
	f2f12f1.SetBlockDurationMinutes(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.SpotOptions.BlockDurationMinutes)
	f2f12f1.SetInstanceInterruptionBehavior(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior)
	f2f12f1.SetMaxPrice(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.SpotOptions.MaxPrice)
	f2f12f1.SetSpotInstanceType(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.SpotOptions.SpotInstanceType)
	f2f12f1.SetValidUntil(*r.ko.Spec.LaunchTemplateData.InstanceMarketOptions.SpotOptions.ValidUntil.Time)
	f2f12.SetSpotOptions(f2f12f1)
	f2.SetInstanceMarketOptions(f2f12)
	f2.SetInstanceType(*r.ko.Spec.LaunchTemplateData.InstanceType)
	f2.SetKernelId(*r.ko.Spec.LaunchTemplateData.KernelID)
	f2.SetKeyName(*r.ko.Spec.LaunchTemplateData.KeyName)
	f2f16 := []*svcsdk.LaunchTemplateLicenseConfigurationRequest{}
	for _, f2f16iter := range r.ko.Spec.LaunchTemplateData.LicenseSpecifications {
		f2f16elem := &svcsdk.LaunchTemplateLicenseConfigurationRequest{}
		f2f16elem.SetLicenseConfigurationArn(*f2f16iter.LicenseConfigurationARN)
		f2f16 = append(f2f16, f2f16elem)
	}
	f2.SetLicenseSpecifications(f2f16)
	f2f17 := &svcsdk.LaunchTemplateInstanceMetadataOptionsRequest{}
	f2f17.SetHttpEndpoint(*r.ko.Spec.LaunchTemplateData.MetadataOptions.HTTPEndpoint)
	f2f17.SetHttpPutResponseHopLimit(*r.ko.Spec.LaunchTemplateData.MetadataOptions.HTTPPutResponseHopLimit)
	f2f17.SetHttpTokens(*r.ko.Spec.LaunchTemplateData.MetadataOptions.HTTPTokens)
	f2.SetMetadataOptions(f2f17)
	f2f18 := &svcsdk.LaunchTemplatesMonitoringRequest{}
	f2f18.SetEnabled(*r.ko.Spec.LaunchTemplateData.Monitoring.Enabled)
	f2.SetMonitoring(f2f18)
	f2f19 := []*svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
	for _, f2f19iter := range r.ko.Spec.LaunchTemplateData.NetworkInterfaces {
		f2f19elem := &svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
		f2f19elem.SetAssociatePublicIpAddress(*f2f19iter.AssociatePublicIPAddress)
		f2f19elem.SetDeleteOnTermination(*f2f19iter.DeleteOnTermination)
		f2f19elem.SetDescription(*f2f19iter.Description)
		f2f19elem.SetDeviceIndex(*f2f19iter.DeviceIndex)
		f2f19elemf4 := []*string{}
		for _, f2f19elemf4iter := range f2f19iter.Groups {
			f2f19elemf4elem.SetGroups(*f2f19elemf4iter)
			f2f19elemf4 = append(f2f19elemf4, f2f19elemf4elem)
		}
		f2f19elem.SetGroups(f2f19elemf4)
		f2f19elem.SetInterfaceType(*f2f19iter.InterfaceType)
		f2f19elem.SetIpv6AddressCount(*f2f19iter.IPv6AddressCount)
		f2f19elemf7 := []*svcsdk.InstanceIpv6AddressRequest{}
		for _, f2f19elemf7iter := range f2f19iter.IPv6Addresses {
			f2f19elemf7elem := &svcsdk.InstanceIpv6AddressRequest{}
			f2f19elemf7elem.SetIpv6Address(*f2f19elemf7iter.IPv6Address)
			f2f19elemf7 = append(f2f19elemf7, f2f19elemf7elem)
		}
		f2f19elem.SetIpv6Addresses(f2f19elemf7)
		f2f19elem.SetNetworkInterfaceId(*f2f19iter.NetworkInterfaceID)
		f2f19elem.SetPrivateIpAddress(*f2f19iter.PrivateIPAddress)
		f2f19elemf10 := []*svcsdk.PrivateIpAddressSpecification{}
		for _, f2f19elemf10iter := range f2f19iter.PrivateIPAddresses {
			f2f19elemf10elem := &svcsdk.PrivateIpAddressSpecification{}
			f2f19elemf10elem.SetPrimary(*f2f19elemf10iter.Primary)
			f2f19elemf10elem.SetPrivateIpAddress(*f2f19elemf10iter.PrivateIPAddress)
			f2f19elemf10 = append(f2f19elemf10, f2f19elemf10elem)
		}
		f2f19elem.SetPrivateIpAddresses(f2f19elemf10)
		f2f19elem.SetSecondaryPrivateIpAddressCount(*f2f19iter.SecondaryPrivateIPAddressCount)
		f2f19elem.SetSubnetId(*f2f19iter.SubnetID)
		f2f19 = append(f2f19, f2f19elem)
	}
	f2.SetNetworkInterfaces(f2f19)
	f2f20 := &svcsdk.LaunchTemplatePlacementRequest{}
	f2f20.SetAffinity(*r.ko.Spec.LaunchTemplateData.Placement.Affinity)
	f2f20.SetAvailabilityZone(*r.ko.Spec.LaunchTemplateData.Placement.AvailabilityZone)
	f2f20.SetGroupName(*r.ko.Spec.LaunchTemplateData.Placement.GroupName)
	f2f20.SetHostId(*r.ko.Spec.LaunchTemplateData.Placement.HostID)
	f2f20.SetHostResourceGroupArn(*r.ko.Spec.LaunchTemplateData.Placement.HostResourceGroupARN)
	f2f20.SetPartitionNumber(*r.ko.Spec.LaunchTemplateData.Placement.PartitionNumber)
	f2f20.SetSpreadDomain(*r.ko.Spec.LaunchTemplateData.Placement.SpreadDomain)
	f2f20.SetTenancy(*r.ko.Spec.LaunchTemplateData.Placement.Tenancy)
	f2.SetPlacement(f2f20)
	f2.SetRamDiskId(*r.ko.Spec.LaunchTemplateData.RamDiskID)
	f2f22 := []*string{}
	for _, f2f22iter := range r.ko.Spec.LaunchTemplateData.SecurityGroupIDs {
		f2f22elem.SetSecurityGroupIds(*f2f22iter)
		f2f22 = append(f2f22, f2f22elem)
	}
	f2.SetSecurityGroupIds(f2f22)
	f2f23 := []*string{}
	for _, f2f23iter := range r.ko.Spec.LaunchTemplateData.SecurityGroups {
		f2f23elem.SetSecurityGroups(*f2f23iter)
		f2f23 = append(f2f23, f2f23elem)
	}
	f2.SetSecurityGroups(f2f23)
	f2f24 := []*svcsdk.LaunchTemplateTagSpecificationRequest{}
	for _, f2f24iter := range r.ko.Spec.LaunchTemplateData.TagSpecifications {
		f2f24elem := &svcsdk.LaunchTemplateTagSpecificationRequest{}
		f2f24elem.SetResourceType(*f2f24iter.ResourceType)
		f2f24elemf1 := []*svcsdk.Tag{}
		for _, f2f24elemf1iter := range f2f24iter.Tags {
			f2f24elemf1elem := &svcsdk.Tag{}
			f2f24elemf1elem.SetKey(*f2f24elemf1iter.Key)
			f2f24elemf1elem.SetValue(*f2f24elemf1iter.Value)
			f2f24elemf1 = append(f2f24elemf1, f2f24elemf1elem)
		}
		f2f24elem.SetTags(f2f24elemf1)
		f2f24 = append(f2f24, f2f24elem)
	}
	f2.SetTagSpecifications(f2f24)
	f2.SetUserData(*r.ko.Spec.LaunchTemplateData.UserData)
	res.SetLaunchTemplateData(f2)
	res.SetLaunchTemplateName(*r.ko.Spec.LaunchTemplateName)
	f4 := []*svcsdk.TagSpecification{}
	for _, f4iter := range r.ko.Spec.TagSpecifications {
		f4elem := &svcsdk.TagSpecification{}
		f4elem.SetResourceType(*f4iter.ResourceType)
		f4elemf1 := []*svcsdk.Tag{}
		for _, f4elemf1iter := range f4iter.Tags {
			f4elemf1elem := &svcsdk.Tag{}
			f4elemf1elem.SetKey(*f4elemf1iter.Key)
			f4elemf1elem.SetValue(*f4elemf1iter.Value)
			f4elemf1 = append(f4elemf1, f4elemf1elem)
		}
		f4elem.SetTags(f4elemf1)
		f4 = append(f4, f4elem)
	}
	res.SetTagSpecifications(f4)
	res.SetVersionDescription(*r.ko.Spec.VersionDescription)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko.Spec", "res", 1))

	// Check that we properly determined how to find the CreatedBy attribute
	// within the CreateLaunchTemplateResult shape, which has a single field called
	// "LaunchTemplate" that contains the CreatedBy field.
	expCreateOutput := `
	ko.Status.CreateTime = &metav1.Time{*resp.LaunchTemplate.CreateTime}
	ko.Status.CreatedBy = resp.LaunchTemplate.CreatedBy
	ko.Status.DefaultVersionNumber = resp.LaunchTemplate.DefaultVersionNumber
	ko.Status.LatestVersionNumber = resp.LaunchTemplate.LatestVersionNumber
	ko.Status.LaunchTemplateID = resp.LaunchTemplate.LaunchTemplateId
	f6 := []*svcapitypes.Tag{}
	for _, f6iter := range resp.LaunchTemplate.Tags {
		f6elem := &svcapitypes.Tag{}
		f6elem.Key = f6iter.Key
		f6elem.Value = f6iter.Value
		f6 = append(f6, f6elem)
	}
	ko.Status.Tags = f6
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The EC2 LaunchTemplate API has a "normal" set of CUD operations:
	//
	// * CreateLaunchTemplate
	// * ModifyLaunchTemplate
	// * DeleteLaunchTemplate
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.Update)

	// However, oddly, there is no ReadOne operation. There is only a
	// ReadMany/List operation "DescribeLaunchTemplates" :(
	//
	// TODO(jaypipes): Develop strategy for informing the code generator via
	// the SDKMapper that certain APIs don't have ReadOne but only ReadMany
	// APIs...
	assert.Nil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)
}

func TestECRRepository(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "ecr")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Repository", crds)
	require.NotNil(crd)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	// The ECR API uses a REST-like API that uses "wrapper" single-member
	// objects in the JSON response for the create/describe calls. In other
	// words, the returned result from the CreateRepository API looks like
	// this:
	//
	// {
	//   "repository": {
	//     .. bunch of fields for the repository ..
	//   }
	// }
	//
	// This test is verifying that we're properly "unwrapping" the structs and
	// putting the repository object's fields into the Spec and Status for the
	// Repository CRD.
	expSpecFieldCamel := []string{
		"ImageScanningConfiguration",
		"ImageTagMutability",
		"RepositoryName",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// ImageScanningConfiguration is in the Repository resource's
	// CreateRepositoryInput shape and also returned in the
	// CreateRepositoryOutput shape, so it should produce Go code to set the
	// appropriate input shape member.
	expCreateInput := `
	f0 := &svcsdk.ImageScanningConfiguration{}
	f0.SetScanOnPush(*r.ko.Spec.ImageScanningConfiguration.ScanOnPush)
	res.SetImageScanningConfiguration(f0)
	res.SetImageTagMutability(*r.ko.Spec.ImageTagMutability)
	res.SetRepositoryName(*r.ko.Spec.RepositoryName)
	f3 := []*svcsdk.Tag{}
	for _, f3iter := range r.ko.Spec.Tags {
		f3elem := &svcsdk.Tag{}
		f3elem.SetKey(*f3iter.Key)
		f3elem.SetValue(*f3iter.Value)
		f3 = append(f3, f3elem)
	}
	res.SetTags(f3)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko.Spec", "res", 1))

	expStatusFieldCamel := []string{
		"CreatedAt",
		// "ImageScanningConfiguration" removed because it is contained in the
		// input shape and therefore exists in the Spec
		// "ImageTagMutability" removed because it is contained in the input
		// shape and therefore exists in the Spec
		"RegistryID",
		// "RepositoryARN" removed because it refers to the primary object's
		// ARN and the SDKMapper identified it for mapping to the standard
		// Status.ACKResourceMetadata.ARN field
		// "RepositoryName" removed because it is contained in the input shape
		// and therefore exists in the Spec
		"RepositoryURI",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// Check that we properly determined how to find the RegistryID attribute
	// within the CreateRepositoryOutput shape, which has a single field called
	// "Repository" that contains the RegistryId field.
	expCreateOutput := `
	ko.Status.CreatedAt = &metav1.Time{*resp.Repository.CreatedAt}
	ko.Status.RegistryID = resp.Repository.RegistryId
	ko.Status.RepositoryURI = resp.Repository.RepositoryUri
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The ECR Repository API has just the C and R of the normal CRUD
	// operations:
	//
	// * CreateRepository
	// * DeleteRepository
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)

	// There is no DescribeRepository operation. There is a List operation for
	// Repositories, though: DescribeRepositories
	assert.Nil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	// There is no update operation (you need to call various SetXXX operations
	// on the Repository's components
	assert.Nil(crd.Ops.Update)
}

func TestCodeDeployDeployment(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "codedeploy")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Deployment", crds)
	require.NotNil(crd)

	assert.Equal("Deployment", crd.Names.Camel)
	assert.Equal("deployment", crd.Names.CamelLower)
	assert.Equal("deployment", crd.Names.Snake)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"ApplicationName",
		"AutoRollbackConfiguration",
		"DeploymentConfigName",
		"DeploymentGroupName",
		"Description",
		"FileExistsBehavior",
		"IgnoreApplicationStopFailures",
		"Revision",
		"TargetInstances",
		"UpdateOutdatedInstancesOnly",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	// However, all of the fields in the Deployment resource's
	// CreateDeploymentInput shape are returned in the GetDeploymentOutput
	// shape, and there is a DeploymentInfo wrapper struct for the output
	// shape, so the readOne accessor contains the wrapper struct's name.
	expCreateOutput := `
	ko.Status.DeploymentID = resp.DeploymentId
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	expStatusFieldCamel := []string{
		// All of the fields in the Deployment resource's CreateDeploymentInput
		// shape are returned in the CreateDeploymentOutput shape so there are
		// not Status fields
		//
		// There is a DeploymentID field in addition to the Spec fields, though.
		"DeploymentID",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// The CodeDeploy Deployment API actually CR+L operations:
	//
	// * CreateDeployment
	// * GetDeployment
	// * ListDeployments
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.ReadOne)
	assert.NotNil(crd.Ops.ReadMany)

	// But sadly, has no Update or Delete operation :(
	assert.Nil(crd.Ops.Update)
	assert.Nil(crd.Ops.Delete)
}

func TestSQSQueue(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "sqs")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	crd := getCRDByName("Queue", crds)
	require.NotNil(crd)

	assert.Equal("Queue", crd.Names.Camel)
	assert.Equal("queue", crd.Names.CamelLower)
	assert.Equal("queue", crd.Names.Snake)

	// The SQS Queue API has CD+L operations:
	//
	// * CreateQueue
	// * DeleteQueue
	// * ListQueues
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.GetAttributes)
	assert.NotNil(crd.Ops.SetAttributes)

	// But sadly, has no Update or ReadOne operation :(
	// There is, however, GetQueueUrl and GetQueueAttributes calls...
	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"ContentBasedDeduplication",
		"DelaySeconds",
		"FifoQueue",
		"KMSDataKeyReusePeriodSeconds",
		"KMSMasterKeyID",
		"MaximumMessageSize",
		"MessageRetentionPeriod",
		"Policy",
		"QueueName",
		"ReceiveMessageWaitTimeSeconds",
		"RedrivePolicy",
		"Tags",
		"VisibilityTimeout",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		// There are a set of Attribute map keys that are readonly
		// fields...
		"CreatedTimestamp",
		"LastModifiedTimestamp",
		// There is only a QueueURL field returned from CreateQueueResult shape
		"QueueURL",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	expCreateInput := `
	attrMap := map[string]*string{}
	attrMap["ContentBasedDeduplication"] = r.ko.Spec.ContentBasedDeduplication
	attrMap["DelaySeconds"] = r.ko.Spec.DelaySeconds
	attrMap["FifoQueue"] = r.ko.Spec.FifoQueue
	attrMap["KmsDataKeyReusePeriodSeconds"] = r.ko.Spec.KMSDataKeyReusePeriodSeconds
	attrMap["KmsMasterKeyId"] = r.ko.Spec.KMSMasterKeyID
	attrMap["MaximumMessageSize"] = r.ko.Spec.MaximumMessageSize
	attrMap["MessageRetentionPeriod"] = r.ko.Spec.MessageRetentionPeriod
	attrMap["Policy"] = r.ko.Spec.Policy
	attrMap["ReceiveMessageWaitTimeSeconds"] = r.ko.Spec.ReceiveMessageWaitTimeSeconds
	attrMap["RedrivePolicy"] = r.ko.Spec.RedrivePolicy
	attrMap["VisibilityTimeout"] = r.ko.Spec.VisibilityTimeout
	res.SetAttributes(attrMap)
	res.SetQueueName(*r.ko.Spec.QueueName)
	res.SetTags(r.ko.Spec.Tags)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko.Spec", "res", 1))

	// There are no fields other than QueueID in the returned CreateQueueResult
	// shape
	expCreateOutput := `
	ko.Status.QueueURL = resp.QueueUrl
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The input shape for the GetAttributes operation technically has two
	// fields in it: an AttributeNames list of attribute keys to file
	// attributes for and a QueueUrl field. We only care about the QueueUrl
	// field, since we look for all attributes for a queue.
	expGetAttrsInput := `
	res.SetQueueUrl(*r.ko.Status.QueueURL)
`
	assert.Equal(expGetAttrsInput, crd.GoCodeGetAttributesSetInput("r.ko", "res", 1))

	// The output shape for the GetAttributes operation contains a single field
	// "Attributes" that must be unpacked into the Queue CRD's Status fields.
	// There are only three attribute keys that are *not* in the Input shape
	// (and thus in the Spec fields). One of them is the resource's ARN which
	// is handled specially.
	expGetAttrsOutput := `
	ko.Status.CreatedTimestamp = resp.Attributes["CreatedTimestamp"]
	ko.Status.LastModifiedTimestamp = resp.Attributes["LastModifiedTimestamp"]
	tmpARN := ackv1alpha1.AWSResourceName(*resp.Attributes["QueueArn"])
	ko.Status.ACKResourceMetadata.ARN = &tmpARN
`
	assert.Equal(expGetAttrsOutput, crd.GoCodeGetAttributesSetOutput("resp", "ko.Status", 1))
}
