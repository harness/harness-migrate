package testreport

import (
	harness "github.com/drone/spec/dist/go"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
)

const (
	stepType   = "script"
	namePrefix = "report"
)

func Convert(c config.StoreTestResults) (*harness.Step, error) {
	name := namePrefix
	if c.Name != nil && *c.Name != "" {
		name = *c.Name
	}

	return &harness.Step{
		Name: name,
		Type: stepType,
		Spec: harness.StepExec{
			Run: "echo Storing test report",
			Reports: []*harness.Report{
				{
					Path: []string{c.Path},
					Type: "junit",
				},
			},
		},
	}, nil
}
