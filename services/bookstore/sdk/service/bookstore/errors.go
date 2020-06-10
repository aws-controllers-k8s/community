// This code was modified from the AppMesh API in
// aws-sdk-go/service/appmesh/errors.go and aws-sdk-go/service/appmesh/api.go

package bookstore

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/private/protocol"
)

const (

	// ErrCodeBadRequestException for service response error code
	// "BadRequestException".
	//
	// The request syntax was malformed. Check your request syntax and try again.
	ErrCodeBadRequestException = "BadRequestException"

	// ErrCodeConflictException for service response error code
	// "ConflictException".
	//
	// The request contains a client token that was used for a previous update resource
	// call with different specifications. Try the request again with a new client
	// token.
	ErrCodeConflictException = "ConflictException"

	// ErrCodeForbiddenException for service response error code
	// "ForbiddenException".
	//
	// You don't have permissions to perform this action.
	ErrCodeForbiddenException = "ForbiddenException"

	// ErrCodeInternalServerErrorException for service response error code
	// "InternalServerErrorException".
	//
	// The request processing has failed because of an unknown error, exception,
	// or failure.
	ErrCodeInternalServerErrorException = "InternalServerErrorException"

	// ErrCodeLimitExceededException for service response error code
	// "LimitExceededException".
	//
	// You have exceeded a service limit for your account. For more information,
	// see Service Limits (https://docs.aws.amazon.com/app-mesh/latest/userguide/service_limits.html)
	// in the AWS App Mesh User Guide.
	ErrCodeLimitExceededException = "LimitExceededException"

	// ErrCodeNotFoundException for service response error code
	// "NotFoundException".
	//
	// The specified resource doesn't exist. Check your request syntax and try again.
	ErrCodeNotFoundException = "NotFoundException"

	// ErrCodeResourceInUseException for service response error code
	// "ResourceInUseException".
	//
	// You can't delete the specified resource because it's in use or required by
	// another resource.
	ErrCodeResourceInUseException = "ResourceInUseException"

	// ErrCodeServiceUnavailableException for service response error code
	// "ServiceUnavailableException".
	//
	// The request has failed due to a temporary failure of the service.
	ErrCodeServiceUnavailableException = "ServiceUnavailableException"

	// ErrCodeTooManyRequestsException for service response error code
	// "TooManyRequestsException".
	//
	// The maximum request rate permitted by the App Mesh APIs has been exceeded
	// for your account. For best results, use an increasing or variable sleep interval
	// between requests.
	ErrCodeTooManyRequestsException = "TooManyRequestsException"

	// ErrCodeTooManyTagsException for service response error code
	// "TooManyTagsException".
	//
	// The request exceeds the maximum allowed number of tags allowed per resource.
	// The current limit is 50 user tags per resource. You must reduce the number
	// of tags in the request. None of the tags in this request were applied.
	ErrCodeTooManyTagsException = "TooManyTagsException"
)

// The request syntax was malformed. Check your request syntax and try again.
type BadRequestException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s BadRequestException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s BadRequestException) GoString() string {
	return s.String()
}

func newErrorBadRequestException(v protocol.ResponseMetadata) error {
	return &BadRequestException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *BadRequestException) Code() string {
	return "BadRequestException"
}

// Message returns the exception's message.
func (s *BadRequestException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *BadRequestException) OrigErr() error {
	return nil
}

func (s *BadRequestException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *BadRequestException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *BadRequestException) RequestID() string {
	return s.RespMetadata.RequestID
}

// You don't have permissions to perform this action.
type ForbiddenException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s ForbiddenException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ForbiddenException) GoString() string {
	return s.String()
}

func newErrorForbiddenException(v protocol.ResponseMetadata) error {
	return &ForbiddenException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *ForbiddenException) Code() string {
	return "ForbiddenException"
}

// Message returns the exception's message.
func (s *ForbiddenException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ForbiddenException) OrigErr() error {
	return nil
}

func (s *ForbiddenException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *ForbiddenException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *ForbiddenException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The request contains a client token that was used for a previous update resource
// call with different specifications. Try the request again with a new client
// token.
type ConflictException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s ConflictException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ConflictException) GoString() string {
	return s.String()
}

func newErrorConflictException(v protocol.ResponseMetadata) error {
	return &ConflictException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *ConflictException) Code() string {
	return "ConflictException"
}

// Message returns the exception's message.
func (s *ConflictException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ConflictException) OrigErr() error {
	return nil
}

func (s *ConflictException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *ConflictException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *ConflictException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The request processing has failed because of an unknown error, exception,
// or failure.
type InternalServerErrorException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s InternalServerErrorException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s InternalServerErrorException) GoString() string {
	return s.String()
}

func newErrorInternalServerErrorException(v protocol.ResponseMetadata) error {
	return &InternalServerErrorException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *InternalServerErrorException) Code() string {
	return "InternalServerErrorException"
}

// Message returns the exception's message.
func (s *InternalServerErrorException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *InternalServerErrorException) OrigErr() error {
	return nil
}

func (s *InternalServerErrorException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *InternalServerErrorException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *InternalServerErrorException) RequestID() string {
	return s.RespMetadata.RequestID
}

// You have exceeded a service limit for your account. For more information,
// see Service Limits (https://docs.aws.amazon.com/app-mesh/latest/userguide/service_limits.html)
// in the AWS App Mesh User Guide.
type LimitExceededException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s LimitExceededException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s LimitExceededException) GoString() string {
	return s.String()
}

func newErrorLimitExceededException(v protocol.ResponseMetadata) error {
	return &LimitExceededException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *LimitExceededException) Code() string {
	return "LimitExceededException"
}

// Message returns the exception's message.
func (s *LimitExceededException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *LimitExceededException) OrigErr() error {
	return nil
}

func (s *LimitExceededException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *LimitExceededException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *LimitExceededException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The specified resource doesn't exist. Check your request syntax and try again.
type NotFoundException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s NotFoundException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s NotFoundException) GoString() string {
	return s.String()
}

func newErrorNotFoundException(v protocol.ResponseMetadata) error {
	return &NotFoundException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *NotFoundException) Code() string {
	return "NotFoundException"
}

// Message returns the exception's message.
func (s *NotFoundException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *NotFoundException) OrigErr() error {
	return nil
}

func (s *NotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *NotFoundException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *NotFoundException) RequestID() string {
	return s.RespMetadata.RequestID
}

// You can't delete the specified resource because it's in use or required by
// another resource.
type ResourceInUseException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s ResourceInUseException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ResourceInUseException) GoString() string {
	return s.String()
}

func newErrorResourceInUseException(v protocol.ResponseMetadata) error {
	return &ResourceInUseException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *ResourceInUseException) Code() string {
	return "ResourceInUseException"
}

// Message returns the exception's message.
func (s *ResourceInUseException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ResourceInUseException) OrigErr() error {
	return nil
}

func (s *ResourceInUseException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *ResourceInUseException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *ResourceInUseException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The request has failed due to a temporary failure of the service.
type ServiceUnavailableException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s ServiceUnavailableException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ServiceUnavailableException) GoString() string {
	return s.String()
}

func newErrorServiceUnavailableException(v protocol.ResponseMetadata) error {
	return &ServiceUnavailableException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *ServiceUnavailableException) Code() string {
	return "ServiceUnavailableException"
}

// Message returns the exception's message.
func (s *ServiceUnavailableException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *ServiceUnavailableException) OrigErr() error {
	return nil
}

func (s *ServiceUnavailableException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *ServiceUnavailableException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *ServiceUnavailableException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The maximum request rate permitted by the App Mesh APIs has been exceeded
// for your account. For best results, use an increasing or variable sleep interval
// between requests.
type TooManyRequestsException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s TooManyRequestsException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TooManyRequestsException) GoString() string {
	return s.String()
}

func newErrorTooManyRequestsException(v protocol.ResponseMetadata) error {
	return &TooManyRequestsException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *TooManyRequestsException) Code() string {
	return "TooManyRequestsException"
}

// Message returns the exception's message.
func (s *TooManyRequestsException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *TooManyRequestsException) OrigErr() error {
	return nil
}

func (s *TooManyRequestsException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *TooManyRequestsException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *TooManyRequestsException) RequestID() string {
	return s.RespMetadata.RequestID
}

// The request exceeds the maximum allowed number of tags allowed per resource.
// The current limit is 50 user tags per resource. You must reduce the number
// of tags in the request. None of the tags in this request were applied.
type TooManyTagsException struct {
	_            struct{}                  `type:"structure"`
	RespMetadata protocol.ResponseMetadata `json:"-" xml:"-"`

	Message_ *string `locationName:"message" type:"string"`
}

// String returns the string representation
func (s TooManyTagsException) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TooManyTagsException) GoString() string {
	return s.String()
}

func newErrorTooManyTagsException(v protocol.ResponseMetadata) error {
	return &TooManyTagsException{
		RespMetadata: v,
	}
}

// Code returns the exception type name.
func (s *TooManyTagsException) Code() string {
	return "TooManyTagsException"
}

// Message returns the exception's message.
func (s *TooManyTagsException) Message() string {
	if s.Message_ != nil {
		return *s.Message_
	}
	return ""
}

// OrigErr always returns nil, satisfies awserr.Error interface.
func (s *TooManyTagsException) OrigErr() error {
	return nil
}

func (s *TooManyTagsException) Error() string {
	return fmt.Sprintf("%s: %s", s.Code(), s.Message())
}

// Status code returns the HTTP status code for the request's response error.
func (s *TooManyTagsException) StatusCode() int {
	return s.RespMetadata.StatusCode
}

// RequestID returns the service's response RequestID for request.
func (s *TooManyTagsException) RequestID() string {
	return s.RespMetadata.RequestID
}

var exceptionFromCode = map[string]func(protocol.ResponseMetadata) error{
	"BadRequestException":          newErrorBadRequestException,
	"ConflictException":            newErrorConflictException,
	"ForbiddenException":           newErrorForbiddenException,
	"InternalServerErrorException": newErrorInternalServerErrorException,
	"LimitExceededException":       newErrorLimitExceededException,
	"NotFoundException":            newErrorNotFoundException,
	"ResourceInUseException":       newErrorResourceInUseException,
	"ServiceUnavailableException":  newErrorServiceUnavailableException,
	"TooManyRequestsException":     newErrorTooManyRequestsException,
	"TooManyTagsException":         newErrorTooManyTagsException,
}
