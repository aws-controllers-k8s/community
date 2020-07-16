package bookstore

import (
	"flag"
	"fmt"
	"testing"

	"github.com/aws/aws-controllers-k8s/test/integration/services"

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
	RunSpecs(t, "ACK BookStore")
}

var _ = BeforeSuite(func() {
	fmt.Printf("Using KUBECONFIG=\"%s\"\n", framework.TestContext.KubeConfig)
})

var _ = Describe("[bookstore-integration]", func() {
	It("should able to run", func() {
		fmt.Println("Tests running successfully")
	})
})
