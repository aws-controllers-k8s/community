{{- define "sdk_find_read_many" -}}
// Generate{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }} returns input for read
// operation.
func Generate{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}(cr *svcapitypes.{{ .CRD.Names.Camel }}) *svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }} {
	res := &svcsdk.{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}{}
{{ GoCodeSetReadManyInput .CRD "cr" "res" 1 }}
	return res
}

// Generate{{ .CRD.Names.Camel }} returns the current state in the form of *svcapitypes.{{ .CRD.Names.Camel }}.
func Generate{{ .CRD.Names.Camel }}(resp *svcsdk.{{ .CRD.Ops.ReadMany.OutputRef.Shape.ShapeName }}) *svcapitypes.{{ .CRD.Names.Camel }} {
	cr := &svcapitypes.{{ .CRD.Names.Camel }}{}
{{ GoCodeSetReadManyOutput .CRD "resp" "cr" 1 false }}
return cr
}
{{- end -}}
