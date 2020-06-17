// This code was modified from the AppMesh API in
// aws-sdk-go/service/appmesh/api.go

package petstore

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/request"
)

// An object that represents a service pet returned by a describe operation.
type PetData struct {
	_ struct{} `type:"structure"`

	// PetName is a required field
	PetName *string `locationName:"petName" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s PetData) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s PetData) GoString() string {
	return s.String()
}

// SetPetName sets the PetName field's value.
func (s *PetData) SetPetName(v string) *PetData {
	s.PetName = &v
	return s
}

// An object that represents a service pet returned by a list operation.
type PetRef struct {
	_ struct{} `type:"structure"`

	// Arn is a required field
	Arn *string `locationName:"arn" type:"string" required:"true"`

	// PetName is a required field
	PetName *string `locationName:"petName" min:"1" type:"string" required:"true"`

	// PetOwner is a required field
	PetOwner *string `locationName:"petOwner" min:"12" type:"string" required:"true"`
}

// String returns the string representation
func (s PetRef) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s PetRef) GoString() string {
	return s.String()
}

// SetArn sets the Arn field's value.
func (s *PetRef) SetArn(v string) *PetRef {
	s.Arn = &v
	return s
}

// SetPetName sets the PetName field's value.
func (s *PetRef) SetPetName(v string) *PetRef {
	s.PetName = &v
	return s
}

// SetPetOwner sets the PetOwner field's value.
func (s *PetRef) SetPetOwner(v string) *PetRef {
	s.PetOwner = &v
	return s
}

// Optional metadata that you apply to a resource to assist with categorization
// and organization. Each tag consists of a key and an optional value, both
// of which you define. Tag keys can have a maximum character length of 128
// characters, and tag values can have a maximum length of 256 characters.
type TagRef struct {
	_ struct{} `type:"structure"`

	// Key is a required field
	Key *string `locationName:"key" min:"1" type:"string" required:"true"`

	Value *string `locationName:"value" type:"string"`
}

// String returns the string representation
func (s TagRef) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TagRef) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *TagRef) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "TagRef"}
	if s.Key == nil {
		invalidParams.Add(request.NewErrParamRequired("Key"))
	}
	if s.Key != nil && len(*s.Key) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Key", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetKey sets the Key field's value.
func (s *TagRef) SetKey(v string) *TagRef {
	s.Key = &v
	return s
}

// SetValue sets the Value field's value.
func (s *TagRef) SetValue(v string) *TagRef {
	s.Value = &v
	return s
}

const opCreatePet = "CreatePet"

type CreatePetInput struct {
	_ struct{} `type:"structure"`

	// PetName is a required field
	PetName *string `locationName:"petName" min:"1" type:"string" required:"true"`

	Tags []*TagRef `locationName:"tags" type:"list"`
}

// String returns the string representation
func (s CreatePetInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreatePetInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *CreatePetInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "CreatePetInput"}
	if s.PetName == nil {
		invalidParams.Add(request.NewErrParamRequired("PetName"))
	}
	if s.PetName != nil && len(*s.PetName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("PetName", 1))
	}
	if s.Tags != nil {
		for i, v := range s.Tags {
			if v == nil {
				continue
			}
			if err := v.Validate(); err != nil {
				invalidParams.AddNested(fmt.Sprintf("%s[%v]", "Tags", i), err.(request.ErrInvalidParams))
			}
		}
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetPetName sets the PetName field's value.
func (s *CreatePetInput) SetPetName(v string) *CreatePetInput {
	s.PetName = &v
	return s
}

// SetTags sets the Tags field's value.
func (s *CreatePetInput) SetTags(v []*TagRef) *CreatePetInput {
	s.Tags = v
	return s
}

type CreatePetOutput struct {
	_ struct{} `type:"structure" payload:"Pet"`

	// An object that represents a service pet returned by a describe operation.
	//
	// Pet is a required field
	Pet *PetData `locationName:"pet" type:"structure" required:"true"`
}

// String returns the string representation
func (s CreatePetOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreatePetOutput) GoString() string {
	return s.String()
}

// SetPet sets the Pet field's value.
func (s *CreatePetOutput) SetPet(v *PetData) *CreatePetOutput {
	s.Pet = v
	return s
}

// CreatePetRequest generates a "aws/request.Request" representing the
// client's request for the CreatePet operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See CreatePet for more information on using the CreatePet
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the CreatePetRequest method.
//    req, resp := client.CreatePetRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *PetstoreAPI) CreatePetRequest(input *CreatePetInput) (req *request.Request, output *CreatePetOutput) {
	op := &request.Operation{
		Name:       opCreatePet,
		HTTPMethod: "PUT",
		HTTPPath:   "/v20190125/pets",
	}

	if input == nil {
		input = &CreatePetInput{}
	}

	output = &CreatePetOutput{}
	req = c.newRequest(op, input, output)
	return
}

// CreatePet API operation for AWS Petstore.
//
// Creates a pet.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Petstore's
// API operation CreatePet for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   The request syntax was malformed. Check your request syntax and try again.
//
//   * ConflictException
//   The request contains a client token that was used for a previous update resource
//   call with different specifications. Try the request again with a new client
//   token.
//
//   * ForbiddenException
//   You don't have permissions to perform this action.
//
//   * InternalServerErrorException
//   The request processing has failed because of an unknown error, exception,
//   or failure.
//
//   * LimitExceededException
//   You have exceeded a service limit for your account. For more information,
//   see Service Limits (https://docs.aws.amazon.com/app-pet/latest/userguide/service_limits.html)
//   in the AWS Petstore User Guide.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Petstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/awssdkapi-2019-01-25/CreatePet
func (c *PetstoreAPI) CreatePet(input *CreatePetInput) (*CreatePetOutput, error) {
	req, out := c.CreatePetRequest(input)
	return out, req.Send()
}

// CreatePetWithContext is the same as CreatePet with the addition of
// the ability to pass a context and additional request options.
//
// See CreatePet for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) CreatePetWithContext(ctx aws.Context, input *CreatePetInput, opts ...request.Option) (*CreatePetOutput, error) {
	req, out := c.CreatePetRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opDeletePet = "DeletePet"

type DeletePetInput struct {
	_ struct{} `type:"structure"`

	// PetName is a required field
	PetName *string `location:"uri" locationName:"petName" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s DeletePetInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DeletePetInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DeletePetInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "DeletePetInput"}
	if s.PetName == nil {
		invalidParams.Add(request.NewErrParamRequired("PetName"))
	}
	if s.PetName != nil && len(*s.PetName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("PetName", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetPetName sets the PetName field's value.
func (s *DeletePetInput) SetPetName(v string) *DeletePetInput {
	s.PetName = &v
	return s
}

type DeletePetOutput struct {
	_ struct{} `type:"structure" payload:"Pet"`

	// An object that represents a service pet returned by a describe operation.
	//
	// Pet is a required field
	Pet *PetData `locationName:"pet" type:"structure" required:"true"`
}

// String returns the string representation
func (s DeletePetOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DeletePetOutput) GoString() string {
	return s.String()
}

// SetPet sets the Pet field's value.
func (s *DeletePetOutput) SetPet(v *PetData) *DeletePetOutput {
	s.Pet = v
	return s
}

// DeletePetRequest generates a "aws/request.Request" representing the
// client's request for the DeletePet operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See DeletePet for more information on using the DeletePet
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the DeletePetRequest method.
//    req, resp := client.DeletePetRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/apppet-2019-01-25/DeletePet
func (c *PetstoreAPI) DeletePetRequest(input *DeletePetInput) (req *request.Request, output *DeletePetOutput) {
	op := &request.Operation{
		Name:       opDeletePet,
		HTTPMethod: "DELETE",
		HTTPPath:   "/v20190125/pets/{petName}",
	}

	if input == nil {
		input = &DeletePetInput{}
	}

	output = &DeletePetOutput{}
	req = c.newRequest(op, input, output)
	return
}

// DeletePet API operation for AWS Petstore.
//
// Deletes an existing pet.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Petstore's
// API operation DeletePet for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   The request syntax was malformed. Check your request syntax and try again.
//
//   * ForbiddenException
//   You don't have permissions to perform this action.
//
//   * InternalServerErrorException
//   The request processing has failed because of an unknown error, exception,
//   or failure.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ResourceInUseException
//   You can't delete the specified resource because it's in use or required by
//   another resource.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Petstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/apppet-2019-01-25/DeletePet
func (c *PetstoreAPI) DeletePet(input *DeletePetInput) (*DeletePetOutput, error) {
	req, out := c.DeletePetRequest(input)
	return out, req.Send()
}

// DeletePetWithContext is the same as DeletePet with the addition of
// the ability to pass a context and additional request options.
//
// See DeletePet for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) DeletePetWithContext(ctx aws.Context, input *DeletePetInput, opts ...request.Option) (*DeletePetOutput, error) {
	req, out := c.DeletePetRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opListPets = "ListPets"

type ListPetsInput struct {
	_ struct{} `type:"structure"`

	Limit *int64 `location:"querystring" locationName:"limit" min:"1" type:"integer"`

	NextToken *string `location:"querystring" locationName:"nextToken" type:"string"`
}

// String returns the string representation
func (s ListPetsInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListPetsInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *ListPetsInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "ListPetsInput"}
	if s.Limit != nil && *s.Limit < 1 {
		invalidParams.Add(request.NewErrParamMinValue("Limit", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetLimit sets the Limit field's value.
func (s *ListPetsInput) SetLimit(v int64) *ListPetsInput {
	s.Limit = &v
	return s
}

// SetNextToken sets the NextToken field's value.
func (s *ListPetsInput) SetNextToken(v string) *ListPetsInput {
	s.NextToken = &v
	return s
}

type ListPetsOutput struct {
	_ struct{} `type:"structure"`

	// Pets is a required field
	Pets []*PetRef `locationName:"pets" type:"list" required:"true"`

	NextToken *string `locationName:"nextToken" type:"string"`
}

// String returns the string representation
func (s ListPetsOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListPetsOutput) GoString() string {
	return s.String()
}

// SetPets sets the Pets field's value.
func (s *ListPetsOutput) SetPets(v []*PetRef) *ListPetsOutput {
	s.Pets = v
	return s
}

// SetNextToken sets the NextToken field's value.
func (s *ListPetsOutput) SetNextToken(v string) *ListPetsOutput {
	s.NextToken = &v
	return s
}

// ListPetsRequest generates a "aws/request.Request" representing the
// client's request for the ListPets operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See ListPets for more information on using the ListPets
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the ListPetsRequest method.
//    req, resp := client.ListPetsRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/apppet-2019-01-25/ListPets
func (c *PetstoreAPI) ListPetsRequest(input *ListPetsInput) (req *request.Request, output *ListPetsOutput) {
	op := &request.Operation{
		Name:       opListPets,
		HTTPMethod: "GET",
		HTTPPath:   "/v20190125/pets",
		Paginator: &request.Paginator{
			InputTokens:     []string{"nextToken"},
			OutputTokens:    []string{"nextToken"},
			LimitToken:      "limit",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &ListPetsInput{}
	}

	output = &ListPetsOutput{}
	req = c.newRequest(op, input, output)
	return
}

// ListPets API operation for AWS Petstore.
//
// Returns a list of existing pets.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Petstore's
// API operation ListPets for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   The request syntax was malformed. Check your request syntax and try again.
//
//   * ForbiddenException
//   You don't have permissions to perform this action.
//
//   * InternalServerErrorException
//   The request processing has failed because of an unknown error, exception,
//   or failure.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Petstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/apppet-2019-01-25/ListPets
func (c *PetstoreAPI) ListPets(input *ListPetsInput) (*ListPetsOutput, error) {
	req, out := c.ListPetsRequest(input)
	return out, req.Send()
}

// ListPetsWithContext is the same as ListPets with the addition of
// the ability to pass a context and additional request options.
//
// See ListPets for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) ListPetsWithContext(ctx aws.Context, input *ListPetsInput, opts ...request.Option) (*ListPetsOutput, error) {
	req, out := c.ListPetsRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

// ListPetsPages iterates over the pages of a ListPets operation,
// calling the "fn" function with the response data for each page. To stop
// iterating, return false from the fn function.
//
// See ListPets method for more information on how to use this operation.
//
// Note: This operation can generate multiple requests to a service.
//
//    // Example iterating over at most 3 pages of a ListPets operation.
//    pageNum := 0
//    err := client.ListPetsPages(params,
//        func(page *apppet.ListPetsOutput, lastPage bool) bool {
//            pageNum++
//            fmt.Println(page)
//            return pageNum <= 3
//        })
//
func (c *PetstoreAPI) ListPetsPages(input *ListPetsInput, fn func(*ListPetsOutput, bool) bool) error {
	return c.ListPetsPagesWithContext(aws.BackgroundContext(), input, fn)
}

// ListPetsPagesWithContext same as ListPetsPages except
// it takes a Context and allows setting request options on the pages.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) ListPetsPagesWithContext(ctx aws.Context, input *ListPetsInput, fn func(*ListPetsOutput, bool) bool, opts ...request.Option) error {
	p := request.Pagination{
		NewRequest: func() (*request.Request, error) {
			var inCpy *ListPetsInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.ListPetsRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}

	for p.Next() {
		if !fn(p.Page().(*ListPetsOutput), !p.HasNextPage()) {
			break
		}
	}

	return p.Err()
}

const opDescribePet = "DescribePet"

type DescribePetInput struct {
	_ struct{} `type:"structure"`

	// PetName is a required field
	PetName *string `location:"uri" locationName:"petName" min:"1" type:"string" required:"true"`

	PetOwner *string `location:"querystring" locationName:"petOwner" min:"12" type:"string"`
}

// String returns the string representation
func (s DescribePetInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribePetInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DescribePetInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "DescribePetInput"}
	if s.PetName == nil {
		invalidParams.Add(request.NewErrParamRequired("PetName"))
	}
	if s.PetName != nil && len(*s.PetName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("PetName", 1))
	}
	if s.PetOwner != nil && len(*s.PetOwner) < 12 {
		invalidParams.Add(request.NewErrParamMinLen("PetOwner", 12))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetPetName sets the PetName field's value.
func (s *DescribePetInput) SetPetName(v string) *DescribePetInput {
	s.PetName = &v
	return s
}

// SetPetOwner sets the PetOwner field's value.
func (s *DescribePetInput) SetPetOwner(v string) *DescribePetInput {
	s.PetOwner = &v
	return s
}

type DescribePetOutput struct {
	_ struct{} `type:"structure" payload:"Pet"`

	// An object that represents a service pet returned by a describe operation.
	//
	// Pet is a required field
	Pet *PetData `locationName:"pet" type:"structure" required:"true"`
}

// String returns the string representation
func (s DescribePetOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribePetOutput) GoString() string {
	return s.String()
}

// SetPet sets the Pet field's value.
func (s *DescribePetOutput) SetPet(v *PetData) *DescribePetOutput {
	s.Pet = v
	return s
}

// DescribePetRequest generates a "aws/request.Request" representing the
// client's request for the DescribePet operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See DescribePet for more information on using the DescribePet
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the DescribePetRequest method.
//    req, resp := client.DescribePetRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appmesh-2019-01-25/DescribePet
func (c *PetstoreAPI) DescribePetRequest(input *DescribePetInput) (req *request.Request, output *DescribePetOutput) {
	op := &request.Operation{
		Name:       opDescribePet,
		HTTPMethod: "GET",
		HTTPPath:   "/v20190125/meshes/{meshName}",
	}

	if input == nil {
		input = &DescribePetInput{}
	}

	output = &DescribePetOutput{}
	req = c.newRequest(op, input, output)
	return
}

// DescribePet API operation for AWS Petstore.
//
// Describes an existing service mesh.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Petstore's
// API operation DescribePet for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   The request syntax was malformed. Check your request syntax and try again.
//
//   * ForbiddenException
//   You don't have permissions to perform this action.
//
//   * InternalServerErrorException
//   The request processing has failed because of an unknown error, exception,
//   or failure.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Petstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appmesh-2019-01-25/DescribePet
func (c *PetstoreAPI) DescribePet(input *DescribePetInput) (*DescribePetOutput, error) {
	req, out := c.DescribePetRequest(input)
	return out, req.Send()
}

// DescribePetWithContext is the same as DescribePet with the addition of
// the ability to pass a context and additional request options.
//
// See DescribePet for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) DescribePetWithContext(ctx aws.Context, input *DescribePetInput, opts ...request.Option) (*DescribePetOutput, error) {
	req, out := c.DescribePetRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opUpdatePet = "UpdatePet"

type UpdatePetInput struct {
	_ struct{} `type:"structure"`

	// PetName is a required field
	PetName *string `location:"uri" locationName:"petName" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s UpdatePetInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UpdatePetInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *UpdatePetInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "UpdatePetInput"}
	if s.PetName == nil {
		invalidParams.Add(request.NewErrParamRequired("PetName"))
	}
	if s.PetName != nil && len(*s.PetName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("PetName", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetPetName sets the PetName field's value.
func (s *UpdatePetInput) SetPetName(v string) *UpdatePetInput {
	s.PetName = &v
	return s
}

type UpdatePetOutput struct {
	_ struct{} `type:"structure" payload:"Pet"`

	// An object that represents a service pet returned by a describe operation.
	//
	// Pet is a required field
	Pet *PetData `locationName:"pet" type:"structure" required:"true"`
}

// String returns the string representation
func (s UpdatePetOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UpdatePetOutput) GoString() string {
	return s.String()
}

// SetPet sets the Pet field's value.
func (s *UpdatePetOutput) SetPet(v *PetData) *UpdatePetOutput {
	s.Pet = v
	return s
}

// UpdatePetRequest generates a "aws/request.Request" representing the
// client's request for the UpdatePet operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See UpdatePet for more information on using the UpdatePet
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the UpdatePetRequest method.
//    req, resp := client.UpdatePetRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/apppet-2019-01-25/UpdatePet
func (c *PetstoreAPI) UpdatePetRequest(input *UpdatePetInput) (req *request.Request, output *UpdatePetOutput) {
	op := &request.Operation{
		Name:       opUpdatePet,
		HTTPMethod: "PUT",
		HTTPPath:   "/v20190125/pets/{petName}",
	}

	if input == nil {
		input = &UpdatePetInput{}
	}

	output = &UpdatePetOutput{}
	req = c.newRequest(op, input, output)
	return
}

// UpdatePet API operation for AWS Petstore.
//
// Updates an existing pet.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Petstore's
// API operation UpdatePet for usage and error information.
//
// Returned Error Types:
//   * BadRequestException
//   The request syntax was malformed. Check your request syntax and try again.
//
//   * ConflictException
//   The request contains a client token that was used for a previous update resource
//   call with different specifications. Try the request again with a new client
//   token.
//
//   * ForbiddenException
//   You don't have permissions to perform this action.
//
//   * InternalServerErrorException
//   The request processing has failed because of an unknown error, exception,
//   or failure.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Petstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
func (c *PetstoreAPI) UpdatePet(input *UpdatePetInput) (*UpdatePetOutput, error) {
	req, out := c.UpdatePetRequest(input)
	return out, req.Send()
}

// UpdatePetWithContext is the same as UpdatePet with the addition of
// the ability to pass a context and additional request options.
//
// See UpdatePet for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *PetstoreAPI) UpdatePetWithContext(ctx aws.Context, input *UpdatePetInput, opts ...request.Option) (*UpdatePetOutput, error) {
	req, out := c.UpdatePetRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}
