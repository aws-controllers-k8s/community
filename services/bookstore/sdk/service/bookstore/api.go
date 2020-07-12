// This code was modified from the AppMesh API in
// aws-sdk-go/service/appmesh/api.go

package bookstore

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/request"
)

// An object that represents a service book returned by a describe operation.
type BookData struct {
	_ struct{} `type:"structure"`

	// BookName is a required field
	BookName *string `locationName:"bookName" min:"1" type:"string" required:"true"`

	// Title is a required field
	Title *string `locationName:"title" min:"1" type:"string" required:"true"`

	// Author is a required field
	Author *string `locationName:"author" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s BookData) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s BookData) GoString() string {
	return s.String()
}

// SetBookName sets the BookName field's value.
func (s *BookData) SetBookName(v string) *BookData {
	s.BookName = &v
	return s
}

// An object that represents a service book returned by a list operation.
type BookRef struct {
	_ struct{} `type:"structure"`

	// Arn is a required field
	Arn *string `locationName:"arn" type:"string" required:"true"`

	// BookName is a required field
	BookName *string `locationName:"bookName" min:"1" type:"string" required:"true"`

	// Title is a required field
	Title *string `locationName:"title" min:"1" type:"string" required:"true"`

	// Author is a required field
	Author *string `locationName:"author" min:"1" type:"string" required:"true"`

	// CreateTime is a required field
	CreateTime *time.Time `locationName:"createTime" type:"timestamp" required:"true"`
}

// String returns the string representation
func (s BookRef) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s BookRef) GoString() string {
	return s.String()
}

// SetArn sets the Arn field's value.
func (s *BookRef) SetArn(v string) *BookRef {
	s.Arn = &v
	return s
}

// SetBookName sets the BookName field's value.
func (s *BookRef) SetBookName(v string) *BookRef {
	s.BookName = &v
	return s
}

// SetTitle sets the Title field's value.
func (s *BookRef) SetTitle(v string) *BookRef {
	s.Title = &v
	return s
}

// SetAuthor sets the Author field's value.
func (s *BookRef) SetAuthor(v string) *BookRef {
	s.Author = &v
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

const opCreateBook = "CreateBook"

type CreateBookInput struct {
	_ struct{} `type:"structure"`

	// BookName is a required field
	BookName *string `locationName:"bookName" min:"1" type:"string" required:"true"`

	// Title is a required field
	Title *string `locationName:"title" min:"1" type:"string" required:"true"`

	// Author is a required field
	Author *string `locationName:"author" min:"1" type:"string" required:"true"`

	Tags []*TagRef `locationName:"tags" type:"list"`
}

// String returns the string representation
func (s CreateBookInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreateBookInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *CreateBookInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "CreateBookInput"}
	if s.BookName == nil {
		invalidParams.Add(request.NewErrParamRequired("BookName"))
	}
	if s.BookName != nil && len(*s.BookName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("BookName", 1))
	}
	if s.Title != nil && len(*s.Title) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Title", 1))
	}
	if s.Author != nil && len(*s.Author) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Author", 1))
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

// SetBookName sets the BookName field's value.
func (s *CreateBookInput) SetBookName(v string) *CreateBookInput {
	s.BookName = &v
	return s
}

// SetTitle sets the Title field's value.
func (s *CreateBookInput) SetTitle(v string) *CreateBookInput {
	s.Title = &v
	return s
}

// SetAuthor sets the Author field's value.
func (s *CreateBookInput) SetAuthor(v string) *CreateBookInput {
	s.Author = &v
	return s
}

// SetTags sets the Tags field's value.
func (s *CreateBookInput) SetTags(v []*TagRef) *CreateBookInput {
	s.Tags = v
	return s
}

type CreateBookOutput struct {
	_ struct{} `type:"structure" payload:"Book"`

	// An object that represents a service book returned by a describe operation.
	//
	// Book is a required field
	Book *BookData `locationName:"book" type:"structure" required:"true"`

	// the timestamp the book was created
	//
	// CreateTime is a required field
	CreateTime *time.Time `locationName:"createTime" type:"timestamp" required:"true"`
}

// String returns the string representation
func (s CreateBookOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreateBookOutput) GoString() string {
	return s.String()
}

// SetBook sets the Book field's value.
func (s *CreateBookOutput) SetBook(v *BookData) *CreateBookOutput {
	s.Book = v
	return s
}

// SetCreateTime sets the CreateTime field's value.
func (s *CreateBookOutput) SetCreateTime(v *time.Time) *CreateBookOutput {
	s.CreateTime = v
	return s
}

// CreateBookRequest generates a "aws/request.Request" representing the
// client's request for the CreateBook operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See CreateBook for more information on using the CreateBook
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the CreateBookRequest method.
//    req, resp := client.CreateBookRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *BookstoreAPI) CreateBookRequest(input *CreateBookInput) (req *request.Request, output *CreateBookOutput) {
	op := &request.Operation{
		Name:       opCreateBook,
		HTTPMethod: "PUT",
		HTTPPath:   "/v20190125/books",
	}

	if input == nil {
		input = &CreateBookInput{}
	}

	output = &CreateBookOutput{}
	req = c.newRequest(op, input, output)
	return
}

// CreateBook API operation for AWS Bookstore.
//
// Creates a book.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Bookstore's
// API operation CreateBook for usage and error information.
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
//   see Service Limits (https://docs.aws.amazon.com/app-book/latest/userguide/service_limits.html)
//   in the AWS Bookstore User Guide.
//
//   * NotFoundException
//   The specified resource doesn't exist. Check your request syntax and try again.
//
//   * ServiceUnavailableException
//   The request has failed due to a temporary failure of the service.
//
//   * TooManyRequestsException
//   The maximum request rate permitted by the Bookstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/awssdkapi-2019-01-25/CreateBook
func (c *BookstoreAPI) CreateBook(input *CreateBookInput) (*CreateBookOutput, error) {
	req, out := c.CreateBookRequest(input)
	return out, req.Send()
}

// CreateBookWithContext is the same as CreateBook with the addition of
// the ability to pass a context and additional request options.
//
// See CreateBook for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) CreateBookWithContext(ctx aws.Context, input *CreateBookInput, opts ...request.Option) (*CreateBookOutput, error) {
	req, out := c.CreateBookRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opDeleteBook = "DeleteBook"

type DeleteBookInput struct {
	_ struct{} `type:"structure"`

	// BookName is a required field
	BookName *string `location:"uri" locationName:"bookName" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s DeleteBookInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DeleteBookInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DeleteBookInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "DeleteBookInput"}
	if s.BookName == nil {
		invalidParams.Add(request.NewErrParamRequired("BookName"))
	}
	if s.BookName != nil && len(*s.BookName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("BookName", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetBookName sets the BookName field's value.
func (s *DeleteBookInput) SetBookName(v string) *DeleteBookInput {
	s.BookName = &v
	return s
}

type DeleteBookOutput struct {
	_ struct{} `type:"structure" payload:"Book"`

	// An object that represents a service book returned by a describe operation.
	//
	// Book is a required field
	Book *BookData `locationName:"book" type:"structure" required:"true"`
}

// String returns the string representation
func (s DeleteBookOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DeleteBookOutput) GoString() string {
	return s.String()
}

// SetBook sets the Book field's value.
func (s *DeleteBookOutput) SetBook(v *BookData) *DeleteBookOutput {
	s.Book = v
	return s
}

// DeleteBookRequest generates a "aws/request.Request" representing the
// client's request for the DeleteBook operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See DeleteBook for more information on using the DeleteBook
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the DeleteBookRequest method.
//    req, resp := client.DeleteBookRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appbook-2019-01-25/DeleteBook
func (c *BookstoreAPI) DeleteBookRequest(input *DeleteBookInput) (req *request.Request, output *DeleteBookOutput) {
	op := &request.Operation{
		Name:       opDeleteBook,
		HTTPMethod: "DELETE",
		HTTPPath:   "/v20190125/books/{bookName}",
	}

	if input == nil {
		input = &DeleteBookInput{}
	}

	output = &DeleteBookOutput{}
	req = c.newRequest(op, input, output)
	return
}

// DeleteBook API operation for AWS Bookstore.
//
// Deletes an existing book.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Bookstore's
// API operation DeleteBook for usage and error information.
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
//   The maximum request rate permitted by the Bookstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appbook-2019-01-25/DeleteBook
func (c *BookstoreAPI) DeleteBook(input *DeleteBookInput) (*DeleteBookOutput, error) {
	req, out := c.DeleteBookRequest(input)
	return out, req.Send()
}

// DeleteBookWithContext is the same as DeleteBook with the addition of
// the ability to pass a context and additional request options.
//
// See DeleteBook for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) DeleteBookWithContext(ctx aws.Context, input *DeleteBookInput, opts ...request.Option) (*DeleteBookOutput, error) {
	req, out := c.DeleteBookRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opListBooks = "ListBooks"

type ListBooksInput struct {
	_ struct{} `type:"structure"`

	Limit *int64 `location:"querystring" locationName:"limit" min:"1" type:"integer"`

	NextToken *string `location:"querystring" locationName:"nextToken" type:"string"`
}

// String returns the string representation
func (s ListBooksInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListBooksInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *ListBooksInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "ListBooksInput"}
	if s.Limit != nil && *s.Limit < 1 {
		invalidParams.Add(request.NewErrParamMinValue("Limit", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetLimit sets the Limit field's value.
func (s *ListBooksInput) SetLimit(v int64) *ListBooksInput {
	s.Limit = &v
	return s
}

// SetNextToken sets the NextToken field's value.
func (s *ListBooksInput) SetNextToken(v string) *ListBooksInput {
	s.NextToken = &v
	return s
}

type ListBooksOutput struct {
	_ struct{} `type:"structure"`

	// Books is a required field
	Books []*BookRef `locationName:"books" type:"list" required:"true"`

	NextToken *string `locationName:"nextToken" type:"string"`
}

// String returns the string representation
func (s ListBooksOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ListBooksOutput) GoString() string {
	return s.String()
}

// SetBooks sets the Books field's value.
func (s *ListBooksOutput) SetBooks(v []*BookRef) *ListBooksOutput {
	s.Books = v
	return s
}

// SetNextToken sets the NextToken field's value.
func (s *ListBooksOutput) SetNextToken(v string) *ListBooksOutput {
	s.NextToken = &v
	return s
}

// ListBooksRequest generates a "aws/request.Request" representing the
// client's request for the ListBooks operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See ListBooks for more information on using the ListBooks
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the ListBooksRequest method.
//    req, resp := client.ListBooksRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appbook-2019-01-25/ListBooks
func (c *BookstoreAPI) ListBooksRequest(input *ListBooksInput) (req *request.Request, output *ListBooksOutput) {
	op := &request.Operation{
		Name:       opListBooks,
		HTTPMethod: "GET",
		HTTPPath:   "/v20190125/books",
		Paginator: &request.Paginator{
			InputTokens:     []string{"nextToken"},
			OutputTokens:    []string{"nextToken"},
			LimitToken:      "limit",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &ListBooksInput{}
	}

	output = &ListBooksOutput{}
	req = c.newRequest(op, input, output)
	return
}

// ListBooks API operation for AWS Bookstore.
//
// Returns a list of existing books.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Bookstore's
// API operation ListBooks for usage and error information.
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
//   The maximum request rate permitted by the Bookstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appbook-2019-01-25/ListBooks
func (c *BookstoreAPI) ListBooks(input *ListBooksInput) (*ListBooksOutput, error) {
	req, out := c.ListBooksRequest(input)
	return out, req.Send()
}

// ListBooksWithContext is the same as ListBooks with the addition of
// the ability to pass a context and additional request options.
//
// See ListBooks for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) ListBooksWithContext(ctx aws.Context, input *ListBooksInput, opts ...request.Option) (*ListBooksOutput, error) {
	req, out := c.ListBooksRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

// ListBooksPages iterates over the pages of a ListBooks operation,
// calling the "fn" function with the response data for each page. To stop
// iterating, return false from the fn function.
//
// See ListBooks method for more information on how to use this operation.
//
// Note: This operation can generate multiple requests to a service.
//
//    // Example iterating over at most 3 pages of a ListBooks operation.
//    pageNum := 0
//    err := client.ListBooksPages(params,
//        func(page *appbook.ListBooksOutput, lastPage bool) bool {
//            pageNum++
//            fmt.Println(page)
//            return pageNum <= 3
//        })
//
func (c *BookstoreAPI) ListBooksPages(input *ListBooksInput, fn func(*ListBooksOutput, bool) bool) error {
	return c.ListBooksPagesWithContext(aws.BackgroundContext(), input, fn)
}

// ListBooksPagesWithContext same as ListBooksPages except
// it takes a Context and allows setting request options on the pages.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) ListBooksPagesWithContext(ctx aws.Context, input *ListBooksInput, fn func(*ListBooksOutput, bool) bool, opts ...request.Option) error {
	p := request.Pagination{
		NewRequest: func() (*request.Request, error) {
			var inCpy *ListBooksInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.ListBooksRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}

	for p.Next() {
		if !fn(p.Page().(*ListBooksOutput), !p.HasNextPage()) {
			break
		}
	}

	return p.Err()
}

const opDescribeBook = "DescribeBook"

type DescribeBookInput struct {
	_ struct{} `type:"structure"`

	// BookName is a required field
	BookName *string `location:"uri" locationName:"bookName" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s DescribeBookInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeBookInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DescribeBookInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "DescribeBookInput"}
	if s.BookName == nil {
		invalidParams.Add(request.NewErrParamRequired("BookName"))
	}
	if s.BookName != nil && len(*s.BookName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("BookName", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetBookName sets the BookName field's value.
func (s *DescribeBookInput) SetBookName(v string) *DescribeBookInput {
	s.BookName = &v
	return s
}

type DescribeBookOutput struct {
	_ struct{} `type:"structure" payload:"Book"`

	// An object that represents a service book returned by a describe operation.
	//
	// Book is a required field
	Book *BookData `locationName:"book" type:"structure" required:"true"`

	// the timestamp the book was created
	//
	// CreateTime is a required field
	CreateTime *time.Time `locationName:"createTime" type:"timestamp" required:"true"`
}

// String returns the string representation
func (s DescribeBookOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeBookOutput) GoString() string {
	return s.String()
}

// SetBook sets the Book field's value.
func (s *DescribeBookOutput) SetBook(v *BookData) *DescribeBookOutput {
	s.Book = v
	return s
}

// SetCreateTime sets the CreateTime field's value.
func (s *DescribeBookOutput) SetCreateTime(v *time.Time) *DescribeBookOutput {
	s.CreateTime = v
	return s
}

// DescribeBookRequest generates a "aws/request.Request" representing the
// client's request for the DescribeBook operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See DescribeBook for more information on using the DescribeBook
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the DescribeBookRequest method.
//    req, resp := client.DescribeBookRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appmesh-2019-01-25/DescribeBook
func (c *BookstoreAPI) DescribeBookRequest(input *DescribeBookInput) (req *request.Request, output *DescribeBookOutput) {
	op := &request.Operation{
		Name:       opDescribeBook,
		HTTPMethod: "GET",
		HTTPPath:   "/v20190125/meshes/{meshName}",
	}

	if input == nil {
		input = &DescribeBookInput{}
	}

	output = &DescribeBookOutput{}
	req = c.newRequest(op, input, output)
	return
}

// DescribeBook API operation for AWS Bookstore.
//
// Describes an existing service mesh.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Bookstore's
// API operation DescribeBook for usage and error information.
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
//   The maximum request rate permitted by the Bookstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appmesh-2019-01-25/DescribeBook
func (c *BookstoreAPI) DescribeBook(input *DescribeBookInput) (*DescribeBookOutput, error) {
	req, out := c.DescribeBookRequest(input)
	return out, req.Send()
}

// DescribeBookWithContext is the same as DescribeBook with the addition of
// the ability to pass a context and additional request options.
//
// See DescribeBook for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) DescribeBookWithContext(ctx aws.Context, input *DescribeBookInput, opts ...request.Option) (*DescribeBookOutput, error) {
	req, out := c.DescribeBookRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

const opUpdateBook = "UpdateBook"

type UpdateBookInput struct {
	_ struct{} `type:"structure"`

	// BookName is a required field
	BookName *string `location:"uri" locationName:"bookName" min:"1" type:"string" required:"true"`

	// Title is a required field
	Title *string `location:"uri" locationName:"title" min:"1" type:"string" required:"true"`

	// Author is a required field
	Author *string `location:"uri" locationName:"author" min:"1" type:"string" required:"true"`
}

// String returns the string representation
func (s UpdateBookInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UpdateBookInput) GoString() string {
	return s.String()
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *UpdateBookInput) Validate() error {
	invalidParams := request.ErrInvalidParams{Context: "UpdateBookInput"}
	if s.BookName == nil {
		invalidParams.Add(request.NewErrParamRequired("BookName"))
	}
	if s.BookName != nil && len(*s.BookName) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("BookName", 1))
	}
	if s.Title != nil && len(*s.Title) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Title", 1))
	}
	if s.Author != nil && len(*s.Author) < 1 {
		invalidParams.Add(request.NewErrParamMinLen("Author", 1))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

// SetBookName sets the BookName field's value.
func (s *UpdateBookInput) SetBookName(v string) *UpdateBookInput {
	s.BookName = &v
	return s
}

// SetTitle sets the Title field's value.
func (s *UpdateBookInput) SetTitle(v string) *UpdateBookInput {
	s.Title = &v
	return s
}

// SetAuthor sets the Author field's value.
func (s *UpdateBookInput) SetAuthor(v string) *UpdateBookInput {
	s.Author = &v
	return s
}

type UpdateBookOutput struct {
	_ struct{} `type:"structure" payload:"Book"`

	// An object that represents a service book returned by a describe operation.
	//
	// Book is a required field
	Book *BookData `locationName:"book" type:"structure" required:"true"`
}

// String returns the string representation
func (s UpdateBookOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UpdateBookOutput) GoString() string {
	return s.String()
}

// SetBook sets the Book field's value.
func (s *UpdateBookOutput) SetBook(v *BookData) *UpdateBookOutput {
	s.Book = v
	return s
}

// UpdateBookRequest generates a "aws/request.Request" representing the
// client's request for the UpdateBook operation. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
// Use "Send" method on the returned Request to send the API call to the service.
// the "output" return value is not valid until after Send returns without error.
//
// See UpdateBook for more information on using the UpdateBook
// API call, and error handling.
//
// This method is useful when you want to inject custom logic or configuration
// into the SDK's request lifecycle. Such as custom headers, or retry logic.
//
//
//    // Example sending a request using the UpdateBookRequest method.
//    req, resp := client.UpdateBookRequest(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
//
// See also, https://docs.aws.amazon.com/goto/WebAPI/appbook-2019-01-25/UpdateBook
func (c *BookstoreAPI) UpdateBookRequest(input *UpdateBookInput) (req *request.Request, output *UpdateBookOutput) {
	op := &request.Operation{
		Name:       opUpdateBook,
		HTTPMethod: "PUT",
		HTTPPath:   "/v20190125/books/{bookName}",
	}

	if input == nil {
		input = &UpdateBookInput{}
	}

	output = &UpdateBookOutput{}
	req = c.newRequest(op, input, output)
	return
}

// UpdateBook API operation for AWS Bookstore.
//
// Updates an existing book.
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for AWS Bookstore's
// API operation UpdateBook for usage and error information.
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
//   The maximum request rate permitted by the Bookstore APIs has been exceeded
//   for your account. For best results, use an increasing or variable sleep interval
//   between requests.
func (c *BookstoreAPI) UpdateBook(input *UpdateBookInput) (*UpdateBookOutput, error) {
	req, out := c.UpdateBookRequest(input)
	return out, req.Send()
}

// UpdateBookWithContext is the same as UpdateBook with the addition of
// the ability to pass a context and additional request options.
//
// See UpdateBook for details on how to use this API operation.
//
// The context must be non-nil and will be used for request cancellation. If
// the context is nil a panic will occur. In the future the SDK may create
// sub-contexts for http.Requests. See https://golang.org/pkg/context/
// for more information on using Contexts.
func (c *BookstoreAPI) UpdateBookWithContext(ctx aws.Context, input *UpdateBookInput, opts ...request.Option) (*UpdateBookOutput, error) {
	req, out := c.UpdateBookRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}
