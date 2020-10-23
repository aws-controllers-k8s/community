{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"
	corev1 "k8s.io/api/core/v1"
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
{{ if .CRD.Ops.ReadOne }}
    {{- template "sdk_find_read_one" . }}
{{- else if .CRD.Ops.GetAttributes }}
    {{- template "sdk_find_get_attributes" . }}
{{- else if .CRD.Ops.ReadMany }}
    {{- template "sdk_find_read_many" . }}
{{- else }}
    {{- template "sdk_find_not_implemented" . }}
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
{{ if .CRD.CustomUpdateMethodName }}
    {{- template "sdk_update_custom" . }}
{{- else if .CRD.Ops.Update }}
    {{- template "sdk_update" . }}
{{- else if .CRD.Ops.SetAttributes }}
    {{- template "sdk_update_set_attributes" . }}
{{- else }}
    {{- template "sdk_update_not_implemented" . }}
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
{{- end }}

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

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions (
	r *resource,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
			break
		}
	}

	if rm.terminalAWSError(err) {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type:   ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		terminalCondition.Status = corev1.ConditionTrue
		awsErr, _ := ackerr.AWSError(err)
		errorMessage := awsErr.Message()
		terminalCondition.Message = &errorMessage
	} else if terminalCondition != nil {
		terminalCondition.Status = corev1.ConditionFalse
		terminalCondition.Message = nil
	}

{{- if $updateConditionsCustomMethodName := .CRD.UpdateConditionsCustomMethodName }}
	// custom update conditions
	customUpdate := rm.{{ $updateConditionsCustomMethodName }}(ko, r, err)
	if terminalCondition != nil || customUpdate {
		return &resource{ko}, true // updated
	}
{{- else }}
	if terminalCondition != nil {
		return &resource{ko}, true // updated
	}
{{- end }}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
{{- if .CRD.TerminalExceptionCodes }}
	if err == nil {
		return false
	}
	awsErr, ok := ackerr.AWSError(err)
	if !ok {
		return false
	}
	switch awsErr.Code() {
	case {{ range $x, $terminalCode := .CRD.TerminalExceptionCodes -}}{{ if ne ($x) (0) }}, {{ end }} "{{ $terminalCode }}"{{ end }}:
		return true
	default:
		return false
	}
{{- else }}
	// No terminal_errors specified for this resource in generator config
	return false
{{- end }}
}
