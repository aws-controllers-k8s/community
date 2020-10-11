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
	empJSON, err := json.MarshalIndent(desired.ko.Spec, "", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Marshal funnction output %s\n\n", string(empJSON))
	empJSON, err = json.MarshalIndent(latest.ko.Spec, "", "    ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Marshal funnction output %s\n", string(empJSON))
	return desired, nil
}
