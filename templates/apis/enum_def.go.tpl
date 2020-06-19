{{- define "enum_def" -}}
type {{ .Names.Camel }} {{ .GoType }}

const (
{{- range $val := .Values }}
	{{ $.Names.Camel }}_{{ $val.Clean }} {{ $.Names.Camel }} = {{ if eq $.GoType "string" }}"{{ end }}{{ $val.Original }}{{ if eq $.GoType "string" }}"{{ end }}
{{- end }}
)
{{- end -}}
