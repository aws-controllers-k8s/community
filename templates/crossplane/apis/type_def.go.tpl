{{- define "type_def" -}}
type {{ .Names.Camel }} struct {
{{- range $attrName, $attr := .Attrs }}
	{{ $attr.Names.Camel }} {{ $attr.GoType }} `json:"{{ $attr.Names.CamelLower }},omitempty"`
{{- end }}
}
{{- end -}}
