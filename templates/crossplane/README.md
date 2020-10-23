# Crossplane Provider Generation

This folder includes the templates to generate AWS Crossplane Provider. Run the
following to generate:

```console
go run -tags codegen cmd/ack-generate/main.go crossplane apis ecr --provider-dir <directory for provider>
```