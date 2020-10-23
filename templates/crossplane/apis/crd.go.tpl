{{- template "boilerplate" }}

package {{ .APIVersion }}

import (
{{- if .CRD.TypeImports }}
{{- range $packagePath, $alias := .CRD.TypeImports }}
    {{ if $alias }}{{ $alias }} {{ end }}"{{ $packagePath }}"
{{ end }}

{{- end }}
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// {{ .CRD.Kind }}Parameters defines the desired state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Parameters struct {
	{{- range $fieldName, $field := .CRD.SpecFields }}
	{{ $field.Names.Camel }} {{ $field.GoType }} `json:"{{ $field.Names.CamelLower }},omitempty"`
{{- end }}
}

// {{ .CRD.Kind }}Spec defines the desired state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Spec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider {{ .CRD.Kind }}Parameters `json:"forProvider"`
}

// {{ .CRD.Kind }}Observation defines the observed state of {{ .CRD.Kind }}
type {{ .CRD.Kind }}Observation struct {
	{{- range $fieldName, $field := .CRD.StatusFields }}
	{{ $field.Names.Camel }} {{ $field.GoType }} `json:"{{ $field.Names.CamelLower }},omitempty"`
{{- end }}
}

// {{ .CRD.Kind }}Status defines the observed state of {{ .CRD.Kind }}.
type {{ .CRD.Kind }}Status struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider {{ .CRD.Kind }}Observation `json:"atProvider"`
}


// +kubebuilder:object:root=true

// {{ .CRD.Kind }} is the Schema for the {{ .CRD.Plural }} API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type {{ .CRD.Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   {{ .CRD.Kind }}Spec   `json:"spec,omitempty"`
	Status {{ .CRD.Kind }}Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// {{ .CRD.Kind }}List contains a list of {{ .CRD.Plural }}
type {{ .CRD.Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []{{ .CRD.Kind }} `json:"items"`
}

// Repository type metadata.
var (
	{{ .CRD.Kind }}Kind             = "{{ .CRD.Kind }}"
	{{ .CRD.Kind }}GroupKind        = schema.GroupKind{Group: Group, Kind: {{ .CRD.Kind }}Kind}.String()
	{{ .CRD.Kind }}KindAPIVersion   = {{ .CRD.Kind }}Kind + "." + GroupVersion.String()
	{{ .CRD.Kind }}GroupVersionKind = GroupVersion.WithKind({{ .CRD.Kind }}Kind)
)

func init() {
	SchemeBuilder.Register(&{{ .CRD.Kind }}{}, &{{ .CRD.Kind }}List{})
}

