{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"

	ackerr "github.com/aws/aws-service-operator-k8s/pkg/errors"

	svcsdk "github.com/aws/aws-sdk-go/service/{{ .ServiceAlias }}"
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
{{- else }}
	// TODO(jaypipes): Map out the ReadMany codepath
{{- end }}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ if .CRD.Ops.ReadOne }}
{{- range $_, $field := .CRD.SpecFields -}}
{{- $goCode := GoCodeSetFieldFromReadOneOutput $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end -}}
{{- range $_, $field := .CRD.StatusFields -}}
{{- $goCode := GoCodeSetFieldFromReadOneOutput $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end -}}
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
	res = &svcsdk.{{ .CRD.Ops.ReadOne.InputRef.Shape.ShapeName }}{}
{{ range $_, $field := .CRD.SpecFields -}}
{{- $goCode := GoCodeSetReadOneInputFromField $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end }}
	return res, nil
}
{{- else }}
 // TODO(jaypipes): Map out the ReadMany codepath
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
	resp, err := rm.sdkapi.{{ .CRD.Ops.Create.Name }}WithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ range $_, $field := .CRD.StatusFields -}}
{{- $goCode := GoCodeSetFieldFromCreateOutput $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end }}
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}, error) {
	res = &svcsdk.{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}{}
{{ range $_, $field := .CRD.SpecFields -}}
{{- $goCode := GoCodeSetCreateInputFromField $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end }}
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
	resp, err := rm.sdkapi.UpdateBookWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()
{{ range $_, $field := .CRD.StatusFields -}}
{{- $goCode := GoCodeSetFieldFromCreateOutput $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end }}
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
) (*svcsdk.UpdateBookInput, error) {
	res = &svcsdk.{{ .CRD.Ops.Update.InputRef.Shape.ShapeName }}{}
{{ range $_, $field := .CRD.SpecFields -}}
{{- $goCode := GoCodeSetUpdateInputFromField $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end }}
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
	_, err = rm.sdkapi.DeleteBookWithContext(ctx, input)
	return err
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
	res = &svcsdk.{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}{}
{{ range $_, $field := .CRD.SpecFields -}}
{{- $goCode := GoCodeSetDeleteInputFromField $field -}}
{{- if $goCode }}
	{{ $goCode }}
{{- end -}}
{{- end -}}
	return res, nil
}
{{- end -}}
