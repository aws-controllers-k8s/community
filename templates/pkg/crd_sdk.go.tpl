{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"
{{- if .CRD.TypeImports }}
{{- range $packagePath, $alias := .CRD.TypeImports }}
	{{ if $alias }}{{ $alias }} {{ end }}"{{ $packagePath }}"
{{ end }}

{{- end }}

	ackv1alpha1 "github.com/aws/aws-controllers-k8s/apis/core/v1alpha1"
	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws/aws-controllers-k8s/services/{{ .ServiceIDClean }}/apis/{{ .APIVersion }}"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &aws.JSONValue{}
	_ = &svcsdk.{{ .SDKAPIInterfaceTypeName}}{}
	_ = &svcapitypes.{{ .CRD.Names.Camel }}{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
{{- if .CRD.Ops.ReadOne }}
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadOneInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}
{{ $setCode := GoCodeSetReadOneOutput .CRD "resp" "ko" 1 true }}
	{{ if not ( Empty $setCode ) }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.ReadOne.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "{{ .CRD.Ops.ReadOne.Name }}", respErr)
	if respErr != nil {
		if awsErr, ok := ackerr.AWSError(respErr); ok && awsErr.Code() == "{{ ResourceExceptionCode .CRD 404 }}" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ $setCode }}
	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
{{- else if .CRD.Ops.GetAttributes }}
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromGetAttributesInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newGetAttributesRequestPayload(r)
	if err != nil {
		return nil, err
	}
{{ $setCode := GoCodeGetAttributesSetOutput .CRD "resp" "ko.Status" 1 }}
	{{ if not ( Empty $setCode ) }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.GetAttributes.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("GET_ATTRIBUTES", "{{ .CRD.Ops.GetAttributes.Name }}", respErr)
	if respErr != nil {
		if awsErr, ok := ackerr.AWSError(respErr); ok && awsErr.Code() == "{{ ResourceExceptionCode .CRD 404 }}" {
			return nil, ackerr.NotFound
		}
		return nil, respErr
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ $setCode }}
	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
{{- else if .CRD.Ops.ReadMany }}
	input, err := rm.newListRequestPayload(r)
	if err != nil {
		return nil, err
	}
{{ $setCode := GoCodeSetReadManyOutput .CRD "resp" "ko" 1 true }}
	{{ if not ( Empty $setCode ) }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.ReadMany.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "{{ .CRD.Ops.ReadMany.Name }}", respErr)
	if respErr != nil {
		if awsErr, ok := ackerr.AWSError(respErr); ok && awsErr.Code() == "{{ ResourceExceptionCode .CRD 404 }}" {
			return nil, ackerr.NotFound
		}
		return nil, respErr
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ $setCode }}
	rm.setStatusDefaults(ko)
{{ if $setOutputCustomMethodName := .CRD.SetOutputCustomMethodName .CRD.Ops.ReadMany }}
	// custom set output from response
	rm.{{ $setOutputCustomMethodName }}(r, resp, ko)
{{ end }}
	return &resource{ko}, nil
{{- else }}
	// Believe it or not, there are API resources that can be created but there
	// is no read operation. Point in case: RDS' CreateDBInstanceReadReplica
	// has no corresponding read operation that I know of...
	return nil, ackerr.NotImplemented
{{- end }}
}

{{- if .CRD.Ops.ReadOne }}
// requiredFieldsMissingFromReadOneInput returns true if there are any fields
// for the ReadOne Input shape that are required by not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadOneInput(
	r *resource,
) bool {
{{ GoCodeRequiredFieldsMissingFromReadOneInput .CRD "r.ko" 1 }}
}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetReadOneInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end }}

{{- if .CRD.Ops.ReadMany }}
// newListRequestPayload returns SDK-specific struct for the HTTP request
// payload of the List API call for the resource
func (rm *resourceManager) newListRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetReadManyInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end }}

{{- if .CRD.Ops.GetAttributes }}
// requiredFieldsMissingFromGetAtttributesInput returns true if there are any
// fields for the GetAttributes Input shape that are required by not present in
// the resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromGetAttributesInput(
	r *resource,
) bool {
{{ GoCodeRequiredFieldsMissingFromGetAttributesInput .CRD "r.ko" 1 }}
}

// newGetAttributesRequestPayload returns SDK-specific struct for the HTTP
// request payload of the GetAttributes API call for the resource
func (rm *resourceManager) newGetAttributesRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }}{}
{{ GoCodeGetAttributesSetInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end }}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a new resource with any fields in the Status field filled in
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	input, err := rm.newCreateRequestPayload(r)
	if err != nil {
		return nil, err
	}
{{ $createCode := GoCodeSetCreateOutput .CRD "resp" "ko" 1 false }}
	{{ if not ( Empty $createCode ) }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.Create.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "{{ .CRD.Ops.Create.Name }}", respErr)
	if respErr != nil {
		return nil, respErr
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ $createCode }}
	rm.setStatusDefaults(ko)
	{{ if $setOutputCustomMethodName := .CRD.SetOutputCustomMethodName .CRD.Ops.Create }}
		// custom set output from response
		rm.{{ $setOutputCustomMethodName }}(r, resp, ko)
	{{ end }}
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetCreateInput .CRD "r.ko" "res" 1 }}
	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
{{- if .CRD.CustomUpdateMethodName }}
	return rm.{{ .CRD.CustomUpdateMethodName }}(ctx, desired, latest, diffReporter)
{{- else if .CRD.Ops.Update }}

{{ $customMethod := .CRD.GetCustomImplementation .CRD.Ops.Update }}
{{ if $customMethod }}
	customResp, customRespErr := rm.{{ $customMethod }}(ctx, desired, latest, diffReporter)
	if customResp != nil || customRespErr != nil {
		return customResp, customRespErr
	}
{{ end }}

	input, err := rm.newUpdateRequestPayload(desired)
	if err != nil {
		return nil, err
	}

{{ $setCode := GoCodeSetUpdateOutput .CRD "resp" "ko" 1 false }}
	{{ if not ( Empty $setCode ) }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.Update.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "{{ .CRD.Ops.Update.Name }}", respErr)
	if respErr != nil {
		return nil, respErr
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()
{{ $setCode }}
	rm.setStatusDefaults(ko)
{{ if $setOutputCustomMethodName := .CRD.SetOutputCustomMethodName .CRD.Ops.Update }}
	// custom set output from response
	rm.{{ $setOutputCustomMethodName }}(desired, resp, ko)
{{ end }}
	return &resource{ko}, nil
{{- else if .CRD.Ops.SetAttributes }}
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. And sdkUpdate should never be called if this is the
	// case, and it's an error in the generated code if it is...
	if rm.requiredFieldsMissingFromSetAttributesInput(desired) {
		panic("Required field in SetAttributes input shape missing!")
	}

	input, err := rm.newSetAttributesRequestPayload(desired)
	if err != nil {
		return nil, err
	}

	// NOTE(jaypipes): SetAttributes calls return a response but they don't
	// contain any useful information. Instead, below, we'll be returning a
	// DeepCopy of the supplied desired state, which should be fine because
	// that desired state has been constructed from a call to GetAttributes...
	_, respErr := rm.sdkapi.{{ .CRD.Ops.SetAttributes.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("SET_ATTRIBUTES", "{{ .CRD.Ops.SetAttributes.Name }}", respErr)
	if respErr != nil {
		if awsErr, ok := ackerr.AWSError(respErr); ok && awsErr.Code() == "{{ ResourceExceptionCode .CRD 404 }}" {
			// Technically, this means someone deleted the backend resource in
			// between the time we got a result back from sdkFind() and here...
			return nil, ackerr.NotFound
		}
		return nil, respErr
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()
	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
{{- else }}
	// TODO(jaypipes): Figure this out...
	return nil, nil
{{- end }}
}

{{- if .CRD.Ops.Update }}
// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Update.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.Update.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetUpdateInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{ end }}

{{- if .CRD.Ops.SetAttributes }}
// requiredFieldsMissingFromSetAtttributesInput returns true if there are any
// fields for the SetAttributes Input shape that are required by not present in
// the resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromSetAttributesInput(
	r *resource,
) bool {
{{ GoCodeRequiredFieldsMissingFromSetAttributesInput .CRD "r.ko" 1 }}
}

// newSetAttributesRequestPayload returns SDK-specific struct for the HTTP
// request payload of the SetAttributes API call for the resource
func (rm *resourceManager) newSetAttributesRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.SetAttributes.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.SetAttributes.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetAttributesSetInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end }}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) error {
{{- if .CRD.Ops.Delete }}
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return err
	}
	_, respErr := rm.sdkapi.{{ .CRD.Ops.Delete.Name }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "{{ .CRD.Ops.Delete.Name }}", respErr)
	return respErr
{{- else }}
	// TODO(jaypipes): Figure this out...
	return nil
{{ end }}
}

{{ if .CRD.Ops.Delete -}}
// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetDeleteInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end -}}

{{ "" }}
// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults (
	ko *svcapitypes.{{ .CRD.Names.Camel }},
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}
