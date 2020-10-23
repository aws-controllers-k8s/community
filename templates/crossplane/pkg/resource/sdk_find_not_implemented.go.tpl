{{- define "sdk_find_not_implemented" -}}
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	// Believe it or not, there are API resources that can be created but there
	// is no read operation. Point in case: RDS' CreateDBInstanceReadReplica
	// has no corresponding read operation that I know of...
	return nil, ackerr.NotImplemented
}
{{- end -}}
