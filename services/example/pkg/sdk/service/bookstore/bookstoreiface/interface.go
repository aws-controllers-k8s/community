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

package bookapiiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	// "github.com/aws/aws-sdk-go/service/{{ .AWSServiceAlias }}"
	awssdkapi "github.com/aws/aws-service-operator-k8s/services/example/pkg/sdk/service/bookstore"
)

type BookstoreAPI interface {
	CreateBook(*awssdkapi.CreateBookInput) (*awssdkapi.CreateBookOutput, error)
	CreateBookWithContext(aws.Context, *awssdkapi.CreateBookInput, ...request.Option) (*awssdkapi.CreateBookOutput, error)
	CreateBookRequest(*awssdkapi.CreateBookInput) (*request.Request, *awssdkapi.CreateBookOutput)

	DeleteBook(*awssdkapi.DeleteBookInput) (*awssdkapi.DeleteBookOutput, error)
	DeleteBookWithContext(aws.Context, *awssdkapi.DeleteBookInput, ...request.Option) (*awssdkapi.DeleteBookOutput, error)
	DeleteBookRequest(*awssdkapi.DeleteBookInput) (*request.Request, *awssdkapi.DeleteBookOutput)

	ListBooks(*awssdkapi.ListBooksInput) (*awssdkapi.ListBooksOutput, error)
	ListBooksWithContext(aws.Context, *awssdkapi.ListBooksInput, ...request.Option) (*awssdkapi.ListBooksOutput, error)
	ListBooksRequest(*awssdkapi.ListBooksInput) (*request.Request, *awssdkapi.ListBooksOutput)

	UpdateBook(*awssdkapi.UpdateBookInput) (*awssdkapi.UpdateBookOutput, error)
	UpdateBookWithContext(aws.Context, *awssdkapi.UpdateBookInput, ...request.Option) (*awssdkapi.UpdateBookOutput, error)
	UpdateBookRequest(*awssdkapi.UpdateBookInput) (*request.Request, *awssdkapi.UpdateBookOutput)
}

// Verify our fake BookstoreAPI implements the above interface
var _ BookstoreAPI = (*awssdkapi.BookstoreAPI)(nil)
