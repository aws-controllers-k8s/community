# Crossplane Provider Generation

This folder includes the templates to generate AWS Crossplane Provider. Run the
following to generate:

```console
go run -tags codegen cmd/ack-generate/main.go crossplane apis ecr --provider-dir <directory for provider>
```

Then you will need to run `go generate ./...` in `provider-dir` so that `kubebuiler`
and other generation tools can generate the rest of machinerty for CRDs and Crossplane.