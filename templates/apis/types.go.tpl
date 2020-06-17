{{ template "boilerplate" }}

package {{ .APIVersion }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
{{- range $typeDef := .TypeDefs }}

{{ template "type_def" $typeDef }}
{{- end -}}
