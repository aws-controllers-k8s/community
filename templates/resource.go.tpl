{{- define "resource" -}}
{{ template "boilerplate" }}

package {{ .APIVersion }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .Resource.Kind }}Spec defines the desired state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Spec struct {
	// The ARN attr is on all AWS service API CRs. It represents the Amazon
	// Resource Name for the object. CRs of this Kind that are created without
	// an ARN attr will be created by the controller. CRs of this Kind that
	// are created with a non-nil ARN attr are considered by the controller to
	// already exist in the backend AWS service API.
	ARN *string `json:"arn,omitempty"`
	{{- range $attrName, $attr := .Resource.SpecAttrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.GoUnexported }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Status struct {
	{{- range $attrName, $attr := .Resource.StatusAttrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.GoUnexported }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}

// {{ .Resource.Kind }} is the Schema for the {{ .Resource.Plural }} API
type {{ .Resource.Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   {{ .Resource.Kind }}Spec   `json:"spec,omitempty"`
	Status {{ .Resource.Kind }}Status `json:"status,omitempty"`
}

// {{ .Resource.Kind }}List contains a list of {{ .Resource.Kind }}
type {{ .Resource.Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []{{ .Resource.Kind }} `json:"items"`
}

func init() {
	SchemeBuilder.Register(&{{ .Resource.Kind }}{}, &{{ .Resource.Kind }}List{})
}
{{- end -}}
