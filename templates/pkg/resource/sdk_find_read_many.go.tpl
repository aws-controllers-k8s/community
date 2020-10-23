{{- define "sdk_find_read_many" -}}
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
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
}

// newListRequestPayload returns SDK-specific struct for the HTTP request
// payload of the List API call for the resource
func (rm *resourceManager) newListRequestPayload(
	r *resource,
) (*svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}, error) {
	res := &svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetReadManyInput .CRD "r.ko" "res" 1 }}
	return res, nil
}
{{- end -}}
