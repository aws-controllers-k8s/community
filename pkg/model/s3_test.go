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

func TestS3_Bucket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	sh := testutil.NewSchemaHelperForService(t, "s3")

	crds, err := sh.GetCRDs()
	require.Nil(err)

	// Pronounced "Boo-Kay".
	crd := getCRDByName("Bucket", crds)
	require.NotNil(crd)

	assert.Equal("Bucket", crd.Names.Camel)
	assert.Equal("bucket", crd.Names.CamelLower)
	assert.Equal("bucket", crd.Names.Snake)

	// The ListBucketsResult shape has no defined error codes (in fact, none of
	// the S3 API shapes do). We will need to create exceptions config in the
	// generate.yaml for S3, but this will take quite some manual work. For
	// now, return UNKNOWN
	assert.Equal("UNKNOWN", crd.ExceptionCode(404))

	// The S3 Bucket API is a whole lot of weird. There are Create and Delete
	// operations ("CreateBucket", "DeleteBucket") but there is no ReadOne
	// operation (there are separate API calls for each and every attribute of
	// a Bucket. For instance, there is a GetBucketCord API call, a
	// GetBucketAnalyticsConfiguration API call, a GetBucketLocation call,
	// etc...) or Update operation (there are separate API calls for each and
	// every attribute of a Bucket, though, for instance PutBucketAcl). There
	// is a ReadMany operation (ListBuckets)
	require.NotNil(crd.Ops)

	assert.NotNil(crd.Ops.Create)
	assert.NotNil(crd.Ops.Delete)
	assert.NotNil(crd.Ops.ReadMany)

	assert.Nil(crd.Ops.GetAttributes)
	assert.Nil(crd.Ops.SetAttributes)
	assert.Nil(crd.Ops.ReadOne)
	assert.Nil(crd.Ops.Update)

	specFields := crd.SpecFields
	statusFields := crd.StatusFields

	expSpecFieldCamel := []string{
		"ACL",
		"CreateBucketConfiguration",
		"GrantFullControl",
		"GrantRead",
		"GrantReadACP",
		"GrantWrite",
		"GrantWriteACP",
		// NOTE(jaypipes): Original field name in CreateBucket input is
		// "Bucket" but should be renamed to "Name" from the generator.yaml (in
		// order to match with the name of the field in the Output shape for a
		// ListBuckets API call...
		"Name",
		"ObjectLockEnabledForBucket",
	}
	assert.Equal(expSpecFieldCamel, attrCamelNames(specFields))

	expStatusFieldCamel := []string{
		"Location",
	}
	assert.Equal(expStatusFieldCamel, attrCamelNames(statusFields))

	expCreateInput := `
	if r.ko.Spec.ACL != nil {
		res.SetACL(*r.ko.Spec.ACL)
	}
	if r.ko.Spec.Name != nil {
		res.SetBucket(*r.ko.Spec.Name)
	}
	if r.ko.Spec.CreateBucketConfiguration != nil {
		f2 := &svcsdk.CreateBucketConfiguration{}
		if r.ko.Spec.CreateBucketConfiguration.LocationConstraint != nil {
			f2.SetLocationConstraint(*r.ko.Spec.CreateBucketConfiguration.LocationConstraint)
		}
		res.SetCreateBucketConfiguration(f2)
	}
	if r.ko.Spec.GrantFullControl != nil {
		res.SetGrantFullControl(*r.ko.Spec.GrantFullControl)
	}
	if r.ko.Spec.GrantRead != nil {
		res.SetGrantRead(*r.ko.Spec.GrantRead)
	}
	if r.ko.Spec.GrantReadACP != nil {
		res.SetGrantReadACP(*r.ko.Spec.GrantReadACP)
	}
	if r.ko.Spec.GrantWrite != nil {
		res.SetGrantWrite(*r.ko.Spec.GrantWrite)
	}
	if r.ko.Spec.GrantWriteACP != nil {
		res.SetGrantWriteACP(*r.ko.Spec.GrantWriteACP)
	}
	if r.ko.Spec.ObjectLockEnabledForBucket != nil {
		res.SetObjectLockEnabledForBucket(*r.ko.Spec.ObjectLockEnabledForBucket)
	}
`
	assert.Equal(expCreateInput, crd.GoCodeSetInput(model.OpTypeCreate, "r.ko", "res", 1))

	expCreateOutput := `
	if resp.Location != nil {
		ko.Status.Location = resp.Location
	}
`
	assert.Equal(expCreateOutput, crd.GoCodeSetOutput(model.OpTypeCreate, "resp", "ko.Status", 1))

	expDeleteInput := `
	if r.ko.Spec.Name != nil {
		res.SetBucket(*r.ko.Spec.Name)
	}
`
	assert.Equal(expDeleteInput, crd.GoCodeSetInput(model.OpTypeDelete, "r.ko", "res", 1))

	expReadManyOutput := `
	if len(resp.Buckets) == 0 {
		return nil, ackerr.NotFound
	}
	found := false
	for _, elem := range resp.Buckets {
		if elem.Name != nil {
			if ko.Spec.Name != nil {
				if *elem.Name != *ko.Spec.Name {
					continue
				}
			}
			ko.Spec.Name = elem.Name
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
