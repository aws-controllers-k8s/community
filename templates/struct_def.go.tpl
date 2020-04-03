{{- define "struct_def" -}}
type {{ .Name }} struct {
	{{- range $attrName, $attr := .Attrs }}
	{{ $attrName }} {{ $attr.GoType }} `json:"{{ $attr.JSONName }},omitempty"`
{{- end }}
}
{{ end -}}
