module github.com/aws/aws-service-operator-k8s

go 1.14

require (
	github.com/aws/aws-sdk-go v1.30.29
	github.com/dlclark/regexp2 v1.2.0
	// pin to v0.1.1 due to release problem with v0.1.2
	github.com/gertd/go-pluralize v0.1.1
	github.com/getkin/kin-openapi v0.3.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	github.com/vektra/mockery v1.1.2 // indirect
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)
