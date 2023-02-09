package steps

import (
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/steps/artifact"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/steps/cache"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/steps/run"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/steps/testreport"

	harness "github.com/drone/spec/dist/go"
)

func Convert(opts commons.Opts, s config.Step) (*harness.Step, error) {
	if s.String != nil && *s.String != "" {
		// Either checkout or job commands
		// TODO (shubham): Handle commands
	}

	if s.StepClass != nil {
		if s.StepClass.Run != nil {
			return run.Convert(*s.StepClass.Run)
		} else if s.StepClass.StoreTestResults != nil {
			return testreport.Convert(*s.StepClass.StoreTestResults)
		} else if s.StepClass.RestoreCache != nil {
			return cache.ConvertRestore(opts, *s.StepClass.RestoreCache)
		} else if s.StepClass.SaveCache != nil {
			return cache.ConvertSave(opts, *s.StepClass.SaveCache)
		} else if s.StepClass.StoreArtifacts != nil {
			return artifact.Convert(opts, *s.StepClass.StoreArtifacts)
		}
	}
	return nil, nil
}
