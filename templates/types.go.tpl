{{ template "boilerplate" }}

package {{ .APIVersion }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
{{- range $enumDef := .EnumDefs }}

{{ template "enum_def" $enumDef }}
{{- end -}}
{{- range $typeDef := .TypeDefs }}

{{ template "type_def" $typeDef }}
{{- end -}}
