package petstore

import (
	"flag"
	"fmt"
	"testing"

	"github.com/aws/aws-service-operator-k8s/test/integration/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/kubernetes/test/e2e/framework"
)

func init() {
	services.RegisterFlags(flag.CommandLine)
}

func TestIntegration(t *testing.T) {
	flag.Parse()
	framework.AfterReadingAllFlags(&framework.TestContext)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Amazon Service Operator K8s Petstore")
}

var _ = BeforeSuite(func() {
	fmt.Printf("Using KUBECONFIG=\"%s\"\n", framework.TestContext.KubeConfig)
})

var _ = Describe("[petstore-integration]", func() {
	It("should able to run", func() {
		fmt.Println("Tests running successfully")
	})
})
