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

package petapiiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	awssdkapi "github.com/aws/aws-service-operator-k8s/services/petstore/sdk/service/petstore"
)

type PetstoreAPI interface {
	CreatePet(*awssdkapi.CreatePetInput) (*awssdkapi.CreatePetOutput, error)
	CreatePetWithContext(aws.Context, *awssdkapi.CreatePetInput, ...request.Option) (*awssdkapi.CreatePetOutput, error)
	CreatePetRequest(*awssdkapi.CreatePetInput) (*request.Request, *awssdkapi.CreatePetOutput)

	DeletePet(*awssdkapi.DeletePetInput) (*awssdkapi.DeletePetOutput, error)
	DeletePetWithContext(aws.Context, *awssdkapi.DeletePetInput, ...request.Option) (*awssdkapi.DeletePetOutput, error)
	DeletePetRequest(*awssdkapi.DeletePetInput) (*request.Request, *awssdkapi.DeletePetOutput)

	ListPets(*awssdkapi.ListPetsInput) (*awssdkapi.ListPetsOutput, error)
	ListPetsWithContext(aws.Context, *awssdkapi.ListPetsInput, ...request.Option) (*awssdkapi.ListPetsOutput, error)
	ListPetsRequest(*awssdkapi.ListPetsInput) (*request.Request, *awssdkapi.ListPetsOutput)

	DescribePet(*awssdkapi.DescribePetInput) (*awssdkapi.DescribePetOutput, error)
	DescribePetWithContext(aws.Context, *awssdkapi.DescribePetInput, ...request.Option) (*awssdkapi.DescribePetOutput, error)
	DescribePetRequest(*awssdkapi.DescribePetInput) (*request.Request, *awssdkapi.DescribePetOutput)

	UpdatePet(*awssdkapi.UpdatePetInput) (*awssdkapi.UpdatePetOutput, error)
	UpdatePetWithContext(aws.Context, *awssdkapi.UpdatePetInput, ...request.Option) (*awssdkapi.UpdatePetOutput, error)
	UpdatePetRequest(*awssdkapi.UpdatePetInput) (*request.Request, *awssdkapi.UpdatePetOutput)
}

// Verify our fake PetstoreAPI implements the above interface
var _ PetstoreAPI = (*awssdkapi.PetstoreAPI)(nil)
