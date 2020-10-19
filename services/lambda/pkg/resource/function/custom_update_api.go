package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
)

func (rm *resourceManager) customUpdateFunction(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	fmt.Println("----", diffReporter.String())

	empJSON, err := json.MarshalIndent(desired.ko, "", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("desired: %s\n\n", string(empJSON))
	empJSON, err = json.MarshalIndent(latest.ko, "", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("latest: %s\n", string(empJSON))
	return nil, nil
}
