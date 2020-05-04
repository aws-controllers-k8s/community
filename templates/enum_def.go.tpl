{{- define "enum_def" -}}
type {{ .Names.GoExported }} {{ .GoType }}

const (
{{- range $val := .Values }}
	{{ $.Names.GoExported }}_{{ $val.Clean }} {{ $.GoType }} = {{ if eq $.GoType "string" }}"{{ end }}{{ $val.Original }}{{ if eq $.GoType "string" }}"{{ end }}
{{- end }}
)
{{- end -}}
