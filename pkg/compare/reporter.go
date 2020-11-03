package compare

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"strings"
)

type DiffItem struct {
	Path string
}

func (diff *DiffItem) String() string {
	return fmt.Sprintf("%#v\n", diff.Path)
}

type Reporter struct {
	path        cmp.Path
	Differences []DiffItem
}

func (reporter *Reporter) PushStep(ps cmp.PathStep) {
	reporter.path = append(reporter.path, ps)
}

func (reporter *Reporter) PopStep() {
	reporter.path = reporter.path[:len(reporter.path)-1]
}

func (reporter *Reporter) Report(result cmp.Result) {
	if !result.Equal() {
		reporter.Differences = append(reporter.Differences, DiffItem{reporter.path.String()})
	}
}

func (reporter *Reporter) String() string {
	var diffs []string
	for _, diff := range reporter.Differences {
		diffs = append(diffs, diff.String())
	}
	return strings.Join(diffs, "\n")
}
