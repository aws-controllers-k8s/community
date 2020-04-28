{{- define "enum_def" -}}
type {{ .Names.GoExported }} {{ .GoType }}

const (
{{- range $val := .Values }}
	{{ $.Names.GoExported }}_{{ $val }} {{ $.GoType }} = {{ if eq $.GoType "string" }}"{{ end }}{{ $val }}{{ if eq $.GoType "string" }}"{{ end }}
{{- end }}
)
{{- end -}}
