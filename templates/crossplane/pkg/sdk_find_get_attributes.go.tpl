{{- define "sdk_find_get_attributes" -}}
// Generate{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }} returns input for read
// operation.
func Generate{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }}(cr *svcapitypes.{{ .CRD.Names.Camel }}) *svcsdk.{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }} {
	res := &svcsdk.{{ .CRD.Ops.GetAttributes.InputRef.Shape.ShapeName }}{}
{{ GoCodeGetAttributesSetInput .CRD "cr" "res" 1 }}
	return res
}

// Generate{{ .CRD.Names.Camel }} returns the current state in the form of *svcapitypes.{{ .CRD.Names.Camel }}.
func Generate{{ .CRD.Names.Camel }}(resp *svcsdk.{{ .CRD.Ops.GetAttributes.OutputRef.Shape.ShapeName }}) *svcapitypes.{{ .CRD.Names.Camel }} {
	cr := &svcapitypes.{{ .CRD.Names.Camel }}{}
{{ GoCodeGetAttributesSetOutput .CRD "resp" "cr" 1 }}
return cr
}
{{- end -}}