{{- define "sdk_update_custom" -}}
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	return rm.{{ .CRD.CustomUpdateMethodName }}(ctx, desired, latest, diffReporter)
}
{{- end -}}
