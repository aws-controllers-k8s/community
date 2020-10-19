module github.com/aws/aws-controllers-k8s

go 1.14

require (
	github.com/aws/aws-sdk-go v1.35.9
	github.com/dlclark/regexp2 v1.2.0
	// pin to v0.1.1 due to release problem with v0.1.2
	github.com/gertd/go-pluralize v0.1.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.3.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.10.0
	golang.org/x/mod v0.2.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)
