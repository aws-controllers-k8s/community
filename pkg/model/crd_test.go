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

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"Attributes",
		"Name",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		// "TopicARN" is the only field in the output shape for CreateTopic,
		// but it is removed because it is the primary resource object's ARN
		// field and the SDKMapper has identified it as the source for the
		// standard Status.ACKResourceMetadata.ARN field
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	// The input shape for the Create operation is set from a variety of scalar
	// and non-scalar types...
	expCreateInput := `
	res.SetAttributes(r.ko.Spec.Attributes)
	res.SetName(*r.ko.Spec.Name)
	tmp0 := []*svcsdk.Tag{}
	for _, elem0 := range res.Tags {
		tmpElem0 := &svcsdk.Tag{}
		tmpElem0.SetKey(*elem0.Key)
		tmpElem0.SetValue(*elem0.Value)
		tmp0 = append(tmp0, tmpElem0)
	}
	res.Tags = tmp0
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "res", "r.ko.Spec", 1))

	// None of the fields in the Topic resource's CreateTopicInput shape are
	// returned in the CreateTopicOutput shape, so none of them return any Go
	// code for setting a Status struct field to a corresponding Create Output
	// Shape member
	expCreateOutput := `
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The SNS Topic API is a little weird. There are Create and Delete
	// operations ("CreateTopic", "DeleteTopic") but there is no ReadOne
	// operation (there is a "GetTopicAttributes" call though) or Update
	// operation (there is a "SetTopicAttributes" call though). And there is a
	// ReadMany operation (ListTopics)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)

	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)
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
	tmp0 := &svcsdk.RequestLaunchTemplateData{}
	tmp0f1 := []*svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
	for _, elem1 := range r.ko.Spec.LaunchTemplateData.BlockDeviceMappings {
		tmpElem1 := &svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
		tmpElem1.SetDeviceName(*elem1.DeviceName)
		tmpElem1f0 := &svcsdk.LaunchTemplateEbsBlockDeviceRequest{}
		tmpElem1f0.SetDeleteOnTermination(*elem1.DeleteOnTermination)
		tmpElem1f0.SetEncrypted(*elem1.Encrypted)
		tmpElem1f0.SetIops(*elem1.IOPS)
		tmpElem1f0.SetKmsKeyId(*elem1.KMSKeyID)
		tmpElem1f0.SetSnapshotId(*elem1.SnapshotID)
		tmpElem1f0.SetVolumeSize(*elem1.VolumeSize)
		tmpElem1f0.SetVolumeType(*elem1.VolumeType)
		tmpElem1.Ebs = tmpElem1f0
		tmpElem1.SetNoDevice(*elem1.NoDevice)
		tmpElem1.SetVirtualName(*elem1.VirtualName)
		tmp0f1 = append(tmp0f1, tmpElem1)
	}
	tmp0.BlockDeviceMappings = tmp0f1
	tmp0f0 := &svcsdk.LaunchTemplateCapacityReservationSpecificationRequest{}
	tmp0f0.SetCapacityReservationPreference(*r.ko.Spec.LaunchTemplateData.CapacityReservationPreference)
	tmp0f0f0 := &svcsdk.CapacityReservationTarget{}
	tmp0f0f0.SetCapacityReservationId(*r.ko.Spec.LaunchTemplateData.CapacityReservationID)
	tmp0f0.CapacityReservationTarget = tmp0f0f0
	tmp0.CapacityReservationSpecification = tmp0f0
	tmp0f0 := &svcsdk.LaunchTemplateCpuOptionsRequest{}
	tmp0f0.SetCoreCount(*r.ko.Spec.LaunchTemplateData.CoreCount)
	tmp0f0.SetThreadsPerCore(*r.ko.Spec.LaunchTemplateData.ThreadsPerCore)
	tmp0.CpuOptions = tmp0f0
	tmp0f0 := &svcsdk.CreditSpecificationRequest{}
	tmp0f0.SetCpuCredits(*r.ko.Spec.LaunchTemplateData.CPUCredits)
	tmp0.CreditSpecification = tmp0f0
	tmp0.SetDisableApiTermination(*r.ko.Spec.LaunchTemplateData.DisableAPITermination)
	tmp0.SetEbsOptimized(*r.ko.Spec.LaunchTemplateData.EBSOptimized)
	tmp0f2 := []*svcsdk.ElasticGpuSpecification{}
	for _, elem2 := range r.ko.Spec.LaunchTemplateData.ElasticGPUSpecifications {
		tmpElem2 := &svcsdk.ElasticGpuSpecification{}
		tmpElem2.SetType(*elem2.Type)
		tmp0f2 = append(tmp0f2, tmpElem2)
	}
	tmp0.ElasticGpuSpecifications = tmp0f2
	tmp0f3 := []*svcsdk.LaunchTemplateElasticInferenceAccelerator{}
	for _, elem3 := range r.ko.Spec.LaunchTemplateData.ElasticInferenceAccelerators {
		tmpElem3 := &svcsdk.LaunchTemplateElasticInferenceAccelerator{}
		tmpElem3.SetCount(*elem3.Count)
		tmpElem3.SetType(*elem3.Type)
		tmp0f3 = append(tmp0f3, tmpElem3)
	}
	tmp0.ElasticInferenceAccelerators = tmp0f3
	tmp0f0 := &svcsdk.LaunchTemplateHibernationOptionsRequest{}
	tmp0f0.SetConfigured(*r.ko.Spec.LaunchTemplateData.Configured)
	tmp0.HibernationOptions = tmp0f0
	tmp0f0 := &svcsdk.LaunchTemplateIamInstanceProfileSpecificationRequest{}
	tmp0f0.SetArn(*r.ko.Spec.LaunchTemplateData.ARN)
	tmp0f0.SetName(*r.ko.Spec.LaunchTemplateData.Name)
	tmp0.IamInstanceProfile = tmp0f0
	tmp0.SetImageId(*r.ko.Spec.LaunchTemplateData.ImageID)
	tmp0.SetInstanceInitiatedShutdownBehavior(*r.ko.Spec.LaunchTemplateData.InstanceInitiatedShutdownBehavior)
	tmp0f0 := &svcsdk.LaunchTemplateInstanceMarketOptionsRequest{}
	tmp0f0.SetMarketType(*r.ko.Spec.LaunchTemplateData.MarketType)
	tmp0f0f0 := &svcsdk.LaunchTemplateSpotMarketOptionsRequest{}
	tmp0f0f0.SetBlockDurationMinutes(*r.ko.Spec.LaunchTemplateData.BlockDurationMinutes)
	tmp0f0f0.SetInstanceInterruptionBehavior(*r.ko.Spec.LaunchTemplateData.InstanceInterruptionBehavior)
	tmp0f0f0.SetMaxPrice(*r.ko.Spec.LaunchTemplateData.MaxPrice)
	tmp0f0f0.SetSpotInstanceType(*r.ko.Spec.LaunchTemplateData.SpotInstanceType)
	tmp0f0f0.SetValidUntil(*r.ko.Spec.LaunchTemplateData.ValidUntil)
	tmp0f0.SpotOptions = tmp0f0f0
	tmp0.InstanceMarketOptions = tmp0f0
	tmp0.SetInstanceType(*r.ko.Spec.LaunchTemplateData.InstanceType)
	tmp0.SetKernelId(*r.ko.Spec.LaunchTemplateData.KernelID)
	tmp0.SetKeyName(*r.ko.Spec.LaunchTemplateData.KeyName)
	tmp0f4 := []*svcsdk.LaunchTemplateLicenseConfigurationRequest{}
	for _, elem4 := range r.ko.Spec.LaunchTemplateData.LicenseSpecifications {
		tmpElem4 := &svcsdk.LaunchTemplateLicenseConfigurationRequest{}
		tmpElem4.SetLicenseConfigurationArn(*elem4.LicenseConfigurationARN)
		tmp0f4 = append(tmp0f4, tmpElem4)
	}
	tmp0.LicenseSpecifications = tmp0f4
	tmp0f0 := &svcsdk.LaunchTemplateInstanceMetadataOptionsRequest{}
	tmp0f0.SetHttpEndpoint(*r.ko.Spec.LaunchTemplateData.HTTPEndpoint)
	tmp0f0.SetHttpPutResponseHopLimit(*r.ko.Spec.LaunchTemplateData.HTTPPutResponseHopLimit)
	tmp0f0.SetHttpTokens(*r.ko.Spec.LaunchTemplateData.HTTPTokens)
	tmp0.MetadataOptions = tmp0f0
	tmp0f0 := &svcsdk.LaunchTemplatesMonitoringRequest{}
	tmp0f0.SetEnabled(*r.ko.Spec.LaunchTemplateData.Enabled)
	tmp0.Monitoring = tmp0f0
	tmp0f5 := []*svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
	for _, elem5 := range r.ko.Spec.LaunchTemplateData.NetworkInterfaces {
		tmpElem5 := &svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
		tmpElem5.SetAssociatePublicIpAddress(*elem5.AssociatePublicIPAddress)
		tmpElem5.SetDeleteOnTermination(*elem5.DeleteOnTermination)
		tmpElem5.SetDescription(*elem5.Description)
		tmpElem5.SetDeviceIndex(*elem5.DeviceIndex)
		tmpElem5f0 := []*string{}
		for _, elem0 := range elem5.Groups {
			tmpElem0.SetSecurityGroupIdStringList(*elem0.Groups.SecurityGroupIDStringList)
			tmpElem5f0 = append(tmpElem5f0, tmpElem0)
		}
		tmpElem5.Groups = tmpElem5f0
		tmpElem5.SetInterfaceType(*elem5.InterfaceType)
		tmpElem5.SetIpv6AddressCount(*elem5.IPv6AddressCount)
		tmpElem5f1 := []*svcsdk.InstanceIpv6AddressRequest{}
		for _, elem1 := range elem5.IPv6Addresses {
			tmpElem1 := &svcsdk.InstanceIpv6AddressRequest{}
			tmpElem1.SetIpv6Address(*elem1.IPv6Address)
			tmpElem5f1 = append(tmpElem5f1, tmpElem1)
		}
		tmpElem5.Ipv6Addresses = tmpElem5f1
		tmpElem5.SetNetworkInterfaceId(*elem5.NetworkInterfaceID)
		tmpElem5.SetPrivateIpAddress(*elem5.PrivateIPAddress)
		tmpElem5f2 := []*svcsdk.PrivateIpAddressSpecification{}
		for _, elem2 := range elem5.PrivateIPAddresses {
			tmpElem2 := &svcsdk.PrivateIpAddressSpecification{}
			tmpElem2.SetPrimary(*elem2.Primary)
			tmpElem2.SetPrivateIpAddress(*elem2.PrivateIPAddress)
			tmpElem5f2 = append(tmpElem5f2, tmpElem2)
		}
		tmpElem5.PrivateIpAddresses = tmpElem5f2
		tmpElem5.SetSecondaryPrivateIpAddressCount(*elem5.SecondaryPrivateIPAddressCount)
		tmpElem5.SetSubnetId(*elem5.SubnetID)
		tmp0f5 = append(tmp0f5, tmpElem5)
	}
	tmp0.NetworkInterfaces = tmp0f5
	tmp0f0 := &svcsdk.LaunchTemplatePlacementRequest{}
	tmp0f0.SetAffinity(*r.ko.Spec.LaunchTemplateData.Affinity)
	tmp0f0.SetAvailabilityZone(*r.ko.Spec.LaunchTemplateData.AvailabilityZone)
	tmp0f0.SetGroupName(*r.ko.Spec.LaunchTemplateData.GroupName)
	tmp0f0.SetHostId(*r.ko.Spec.LaunchTemplateData.HostID)
	tmp0f0.SetHostResourceGroupArn(*r.ko.Spec.LaunchTemplateData.HostResourceGroupARN)
	tmp0f0.SetPartitionNumber(*r.ko.Spec.LaunchTemplateData.PartitionNumber)
	tmp0f0.SetSpreadDomain(*r.ko.Spec.LaunchTemplateData.SpreadDomain)
	tmp0f0.SetTenancy(*r.ko.Spec.LaunchTemplateData.Tenancy)
	tmp0.Placement = tmp0f0
	tmp0.SetRamDiskId(*r.ko.Spec.LaunchTemplateData.RamDiskID)
	tmp0f6 := []*string{}
	for _, elem6 := range r.ko.Spec.LaunchTemplateData.SecurityGroupIDs {
		tmpElem6.SetSecurityGroupIdStringList(*elem6.SecurityGroupIDs.SecurityGroupIDStringList)
		tmp0f6 = append(tmp0f6, tmpElem6)
	}
	tmp0.SecurityGroupIds = tmp0f6
	tmp0f7 := []*string{}
	for _, elem7 := range r.ko.Spec.LaunchTemplateData.SecurityGroups {
		tmpElem7.SetSecurityGroupStringList(*elem7.SecurityGroups.SecurityGroupStringList)
		tmp0f7 = append(tmp0f7, tmpElem7)
	}
	tmp0.SecurityGroups = tmp0f7
	tmp0f8 := []*svcsdk.LaunchTemplateTagSpecificationRequest{}
	for _, elem8 := range r.ko.Spec.LaunchTemplateData.TagSpecifications {
		tmpElem8 := &svcsdk.LaunchTemplateTagSpecificationRequest{}
		tmpElem8.SetResourceType(*elem8.ResourceType)
		tmpElem8f0 := []*svcsdk.Tag{}
		for _, elem0 := range elem8.Tags {
			tmpElem0 := &svcsdk.Tag{}
			tmpElem0.SetKey(*elem0.Key)
			tmpElem0.SetValue(*elem0.Value)
			tmpElem8f0 = append(tmpElem8f0, tmpElem0)
		}
		tmpElem8.Tags = tmpElem8f0
		tmp0f8 = append(tmp0f8, tmpElem8)
	}
	tmp0.TagSpecifications = tmp0f8
	tmp0.SetUserData(*r.ko.Spec.LaunchTemplateData.UserData)
	res.LaunchTemplateData = tmp0
	res.SetLaunchTemplateName(*r.ko.Spec.LaunchTemplateName)
	tmp9 := []*svcsdk.TagSpecification{}
	for _, elem9 := range res.TagSpecifications {
		tmpElem9 := &svcsdk.TagSpecification{}
		tmpElem9.SetResourceType(*elem9.ResourceType)
		tmpElem9f0 := []*svcsdk.Tag{}
		for _, elem0 := range elem9.Tags {
			tmpElem0 := &svcsdk.Tag{}
			tmpElem0.SetKey(*elem0.Key)
			tmpElem0.SetValue(*elem0.Value)
			tmpElem9f0 = append(tmpElem9f0, tmpElem0)
		}
		tmpElem9.Tags = tmpElem9f0
		tmp9 = append(tmp9, tmpElem9)
	}
	res.TagSpecifications = tmp9
	res.SetVersionDescription(*r.ko.Spec.VersionDescription)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "res", "r.ko.Spec", 1))

	// Check that we properly determined how to find the CreatedBy attribute
	// within the CreateLaunchTemplateResult shape, which has a single field called
	// "LaunchTemplate" that contains the CreatedBy field.
	expCreateOutput := `
	ko.Status.CreateTime = resp.LaunchTemplate.CreateTime
	ko.Status.CreatedBy = resp.LaunchTemplate.CreatedBy
	ko.Status.DefaultVersionNumber = resp.LaunchTemplate.DefaultVersionNumber
	ko.Status.LatestVersionNumber = resp.LaunchTemplate.LatestVersionNumber
	ko.Status.LaunchTemplateID = resp.LaunchTemplate.LaunchTemplateId
	tmp0 := []*svcapitypes.Tag{}
	for _, elem0 := range ko.Status.Tags {
		tmpElem0 := &svcapitypes.Tag{}
		tmpElem0.Key = elem0
		tmpElem0.Value = elem0
		tmp0 = append(tmp0, tmpElem0)
	}
	ko.Status.Tags = tmp0
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
	tmp0 := &svcsdk.ImageScanningConfiguration{}
	tmp0.SetScanOnPush(*r.ko.Spec.ImageScanningConfiguration.ScanOnPush)
	res.ImageScanningConfiguration = tmp0
	res.SetImageTagMutability(*r.ko.Spec.ImageTagMutability)
	res.SetRepositoryName(*r.ko.Spec.RepositoryName)
	tmp1 := []*svcsdk.Tag{}
	for _, elem1 := range res.Tags {
		tmpElem1 := &svcsdk.Tag{}
		tmpElem1.SetKey(*elem1.Key)
		tmpElem1.SetValue(*elem1.Value)
		tmp1 = append(tmp1, tmpElem1)
	}
	res.Tags = tmp1
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "res", "r.ko.Spec", 1))

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
	ko.Status.CreatedAt = resp.Repository.CreatedAt
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

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"Attributes",
		"QueueName",
		"Tags",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		// There is only a QueueURL field returned from CreateQueueResult shape
		"QueueURL",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	expCreateInput := `
	res.SetAttributes(r.ko.Spec.Attributes)
	res.SetQueueName(*r.ko.Spec.QueueName)
	res.SetTags(r.ko.Spec.Tags)
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "res", "r.ko.Spec", 1))

	// There are no fields other than QueueID in the returned CreateQueueResult
	// shape
	expCreateOutput := `
	ko.Status.QueueURL = resp.QueueUrl
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	// The SQS Queue API has CD+L operations:
	//
	// * CreateQueue
	// * DeleteQueue
	// * ListQueues
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.ReadMany)
	assert.NotNil(crd.Ops.Delete)

	// But sadly, has no Update or ReadOne operation :(
	// There is, however, GetQueueUrl and GetQueueAttributes calls...
	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)
}
