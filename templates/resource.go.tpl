{{- define "resource" -}}
// {{ .Kind }}Spec defines the desired state of {{ .Kind }}
type {{ .Kind }}Spec struct {
	// The Arn attr is on all AWS service API CRs. It represents the Amazon
	// Resource Name for the object. CRs of this Kind that are created without
	// an Arn attr will be created by the controller. CRs of this Kind that
	// are created with a non-nil Arn attr are considered by the controller to
	// already exist in the backend AWS service API.
	Arn string `json:"arn,omitempty"`
	{{- range $attrName, $attr := .SpecAttrs }}
	{{ $attrName }} {{ $attr.GoType }} `json:"{{ $attr.JSONName }},omitempty"`
{{- end }}
}

// {{ .Kind }}Status defines the observed state of {{ .Kind }}
type {{ .Kind }}Status struct {
	{{- range $attrName, $attr := .StatusAttrs }}
	{{ $attrName }} {{ $attr.GoType }} `json:"{{ $attr.JSONName }},omitempty"`
{{- end }}
}

// {{ .Kind }} is the Schema for the {{ .Plural }} API
type {{ .Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   {{ .Kind }}Spec   `json:"spec,omitempty"`
	Status {{ .Kind }}Status `json:"status,omitempty"`
}

// {{ .Kind }}List contains a list of {{ .Kind }}
type {{ .Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []{{ .Kind }} `json:"items"`
}
{{ end -}}
