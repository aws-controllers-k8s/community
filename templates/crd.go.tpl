{{- define "crd" -}}
{{ template "boilerplate" }}

package {{ .APIVersion }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .CRD.Kind }}Spec defines the desired state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Spec struct {
	// The ARN attr is on all AWS service API CRs. It represents the Amazon
	// Resource Name for the object. CRs of this Kind that are created without
	// an ARN attr will be created by the controller. CRs of this Kind that
	// are created with a non-nil ARN attr are considered by the controller to
	// already exist in the backend AWS service API.
	ARN *string `json:"arn,omitempty"`
	{{- range $attrName, $attr := .CRD.SpecAttrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.GoUnexported }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}

// {{ .CRD.Kind }}Status defines the observed state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Status struct {
	{{- range $attrName, $attr := .CRD.StatusAttrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.GoUnexported }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}

// {{ .CRD.Kind }} is the Schema for the {{ .CRD.Plural }} API
type {{ .CRD.Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   {{ .CRD.Kind }}Spec   `json:"spec,omitempty"`
	Status {{ .CRD.Kind }}Status `json:"status,omitempty"`
}

// {{ .CRD.Kind }}List contains a list of {{ .CRD.Kind }}
type {{ .CRD.Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []{{ .CRD.Kind }} `json:"items"`
}

func init() {
	SchemeBuilder.Register(&{{ .CRD.Kind }}{}, &{{ .CRD.Kind }}List{})
}
{{- end -}}
