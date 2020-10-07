{{- template "boilerplate" }}

package {{ .APIVersion }}

import (
{{- if .CRD.TypeImports }}
{{- range $packagePath, $alias := .CRD.TypeImports }}
    {{ if $alias }}{{ $alias }} {{ end }}"{{ $packagePath }}"
{{ end }}

{{- end }}
    cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .CRD.Kind }}SpecParams defines the desired state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}SpecParams struct {
	{{- range $fieldName, $field := .CRD.SpecFields }}
	{{ $field.Names.Camel }} {{ $field.GoType }} `json:"{{ $field.Names.CamelLower }},omitempty"`
{{- end }}
}

// {{ .CRD.Kind }}Spec defines the desired state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Spec struct {
    cpv1alpha1.ResourceSpec `json:",inline"`
    ForProvider {{ .CRD.Kind }}SpecParams `json:"forProvider"`
}

// {{ .CRD.Kind }}ExternalStatus defines the observed state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}ExternalStatus struct {
    // TODO(negz): place common Crossplane-y stuff.
	{{- range $fieldName, $field := .CRD.StatusFields }}
	{{ $field.Names.Camel }} {{ $field.GoType }} `json:"{{ $field.Names.CamelLower }},omitempty"`
{{- end }}
}

// {{ .CRD.Kind }}Status defines the observed state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Status struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider {{ .CRD.Kind }}ExternalStatus `json:"atProvider"`
}

// {{ .CRD.Kind }} is the Schema for the {{ .CRD.Plural }} API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type {{ .CRD.Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   {{ .CRD.Kind }}Spec   `json:"spec,omitempty"`
	Status {{ .CRD.Kind }}Status `json:"status,omitempty"`
}

// {{ .CRD.Kind }}List contains a list of {{ .CRD.Kind }}
// +kubebuilder:object:root=true
type {{ .CRD.Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []{{ .CRD.Kind }} `json:"items"`
}

func init() {
	SchemeBuilder.Register(&{{ .CRD.Kind }}{}, &{{ .CRD.Kind }}List{})
}

