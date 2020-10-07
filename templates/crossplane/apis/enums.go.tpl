{{- template "boilerplate" }}

package {{ .APIVersion }}
{{- range $enumDef := .EnumDefs }}

{{ template "enum_def" $enumDef }}
{{- end -}}
