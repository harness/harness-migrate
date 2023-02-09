package jobs

import (
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/steps"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/utils"

	harness "github.com/drone/spec/dist/go"
)

func Convert(opts commons.Opts, j config.JobValue) (*harness.Stage, error) {
	hci := harness.StageCI{}

	for _, step := range j.Steps {
		s, err := steps.Convert(opts, step)
		if err != nil {
			return nil, err
		}

		if s != nil {
			hci.Steps = append(hci.Steps, s)
		}
	}
	hci.Envs = utils.ConvertEnvs(j.Environment)

	return &harness.Stage{
		Name: "ci stage",
		Type: "ci",
		Spec: hci,
	}, nil
}
