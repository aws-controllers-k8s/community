{{- define "sdk_find_read_one" -}}
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
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
}

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
{{- end -}}
