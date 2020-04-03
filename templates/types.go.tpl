{{ template "boilerplate" }}

package {{ .Version }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

{{ range $structDef := .StructDefs }}
{{ template "struct_def" $structDef }}
{{- end }}

{{- range $res := .Resources }}
{{ template "resource" $res }}
{{- end }}

func init() {
{{- range $res := .Resources }}
	SchemeBuilder.Register(&{{ $res.Kind }}{}, &{{ $res.Kind }}List{})
{{- end }}
}
