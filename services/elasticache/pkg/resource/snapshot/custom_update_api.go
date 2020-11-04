package snapshot

import (
	"context"
	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
)

// Snapshot API has no update
func (rm *resourceManager) customUpdateSnapshot(
	ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	return desired, nil
}
