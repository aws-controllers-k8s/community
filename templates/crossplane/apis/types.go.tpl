{{- template "boilerplate" }}

package {{ .APIVersion }}
{{- if .Imports }}
import (
{{- end -}}
{{- range $packagePath, $alias := .Imports }}
	{{ if $alias -}}{{ $alias }} {{ end -}}"{{ $packagePath }}"
{{ end -}}
{{- if .Imports }}
)
{{- end -}}
{{- range $typeDef := .TypeDefs }}

{{ template "type_def" $typeDef }}
{{- end -}}
