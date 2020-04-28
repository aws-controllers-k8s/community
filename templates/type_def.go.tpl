{{- define "type_def" -}}
type {{ .Name }} struct {
{{- range $attrName, $attr := .Attrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.JSON }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}
{{- end -}}
