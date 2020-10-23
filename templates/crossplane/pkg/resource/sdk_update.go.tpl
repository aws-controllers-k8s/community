{{- define "sdk_update" -}}
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
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
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.Update.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.Update.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetUpdateInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end -}}
