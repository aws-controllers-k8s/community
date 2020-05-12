{{- define "type_def" -}}
type {{ .Names.GoExported }} struct `aws:"{{ .Names.Original }}"` {
{{- range $attrName, $attr := .Attrs }}
	{{ $attr.Names.GoExported }} {{ $attr.GoType }} `json:"{{ $attr.Names.GoUnexported }},omitempty" aws:"{{ $attr.Names.Original }}"`
{{- end }}
}
{{- end -}}
