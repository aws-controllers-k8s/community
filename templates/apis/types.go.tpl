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
{{- if .HasSecret }}
// SecretReference has enough information to retrieve secret
// in any namespace.
type SecretReference struct {
    // Namespace defines the space within which the secret name must be unique.
    Namespace string `json:"namespace,omitempty"`
    // Name of secret in a given namespace.
    Name string `json:"name,omitempty"`
    // Key of the secret.
    Key string `json:"key,omitempty"`
}
{{- end -}}
