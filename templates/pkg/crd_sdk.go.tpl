{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceAlias }}"

{{- if .CRD.TypeImports }}
{{- range $packagePath, $alias := .CRD.TypeImports }}
	{{ if $alias }}{{ $alias }} {{ end }}"{{ $packagePath }}"
{{ end }}

{{- end }}

{{- if .CRD.Ops.ReadOne }}
	"github.com/aws/aws-sdk-go/aws/awserr"

	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"
{{- end }}

	svcapitypes "github.com/aws/aws-controllers-k8s/services/{{ .ServiceAlias }}/apis/{{ .APIVersion }}"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &svcsdk.{{ .SDKAPIInterfaceTypeName}}{}
	_ = &svcapitypes.{{ .CRD.Names.Camel }}{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
{{- if .CRD.Ops.ReadOne }}
	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.{{ .CRD.Ops.ReadOne.Name }}WithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}
{{- else if .CRD.Ops.GetAttributes }}
	input, err := rm.newGetAttributesRequestPayload(r)
	if err != nil {
		return nil, err
	}
	resp, err := rm.sdkapi.{{ .CRD.Ops.GetAttributes.Name }}WithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}
{{- else }}
	// TODO(jaypipes): Map out the ReadMany codepath
{{- end }}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ if .CRD.Ops.ReadOne }}
{{ GoCodeSetReadOneOutput .CRD "resp" "ko.Status" 1 }}
{{- else if .CRD.Ops.GetAttributes }}
{{ GoCodeGetAttributesSetOutput .CRD "resp" "ko.Status" 1 }}
{{- else }}
	// TODO(jaypipes): Map out the ReadMany codepath
{{- end }}
	return &resource{ko}, nil
}

{{- if .CRD.Ops.ReadOne }}
// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetReadOneInput .CRD "r.ko.Spec" "res" 1 }}
	return res, nil
}
{{- end }}

{{- if .CRD.Ops.GetAttributes }}
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
	{{ if .CRD.StatusFields }}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.Create.Name }}WithContext(ctx, input)
	if respErr != nil {
		return nil, respErr
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ GoCodeSetCreateOutput .CRD "resp" "ko.Status" 1 }}
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetCreateInput .CRD "r.ko.Spec" "res" 1 }}
	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	r *resource,
) (*resource, error) {
{{- if .CRD.Ops.Update }}
	input, err := rm.newUpdateRequestPayload(r)
	if err != nil {
		return nil, err
	}
	{{ if .CRD.StatusFields}}resp{{ else }}_{{ end }}, respErr := rm.sdkapi.{{ .CRD.Ops.Update.Name }}WithContext(ctx, input)
	if respErr != nil {
		return nil, respErr
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ GoCodeSetUpdateOutput .CRD "resp" "ko.Status" 1 }}
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
{{ GoCodeSetUpdateInput .CRD "r.ko.Spec" "res" 1 }}
	return res, nil
}
{{ end }}

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
{{ GoCodeSetDeleteInput .CRD "r.ko.Spec" "res" 1 }}
	return res, nil
}
{{- end -}}
