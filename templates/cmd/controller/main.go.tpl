{{ template "boilerplate" }}

package main

import (
	"os"

	ackrt "github.com/aws/aws-controllers-k8s/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"

	svcresource "github.com/aws/aws-controllers-k8s/services/{{ .ServiceAlias }}/pkg/resource"
	svctypes "github.com/aws/aws-controllers-k8s/services/{{ .ServiceAlias }}/apis/{{ .APIVersion }}"

	{{ $serviceAlias := .ServiceAlias }} {{range $crdName := .SnakeCasedCRDNames }}_ "github.com/aws/aws-controllers-k8s/services/{{ $serviceAlias }}/pkg/resource/{{ $crdName }}"
    {{end}}
)

var (
	awsServiceAPIGroup = "{{ .ServiceAlias }}.services.k8s.aws"
	awsServiceAlias    = "{{ .ServiceAlias }}"
	scheme             = runtime.NewScheme()
	setupLog           = ctrlrt.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = svctypes.AddToScheme(scheme)
}

func main() {
	var ackCfg ackrt.Config
	ackCfg.BindFlags()
	flag.Parse()
	ackCfg.SetupLogger()

	mgr, err := ctrlrt.NewManager(ctrlrt.GetConfigOrDie(), ctrlrt.Options{
		Scheme:             scheme,
		Port:               ackCfg.BindPort,
		MetricsBindAddress: ackCfg.MetricsAddr,
		LeaderElection:     ackCfg.EnableLeaderElection,
		LeaderElectionID:   awsServiceAPIGroup,
	})
	if err != nil {
		setupLog.Error(
			err, "unable to create controller manager",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}

	stopChan := ctrlrt.SetupSignalHandler()

	setupLog.Info(
		"initializing service controller",
		"aws.service", awsServiceAlias,
	)
	sc := ackrt.NewServiceController(
		awsServiceAlias, awsServiceAPIGroup,
	).WithLogger(
		ctrlrt.Log,
	).WithResourceManagerFactories(
		svcresource.GetManagerFactories(),
	)
	if err = sc.BindControllerManager(mgr); err != nil {
		setupLog.Error(
			err, "unable bind to controller manager to service controller",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}

	setupLog.Info(
		"starting manager",
		"aws.service", awsServiceAlias,
	)
	if err := mgr.Start(stopChan); err != nil {
		setupLog.Error(
			err, "unable to start controller manager",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}
}
