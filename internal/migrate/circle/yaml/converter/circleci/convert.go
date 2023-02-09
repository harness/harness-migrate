package circleci

import (
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/jobs"

	harness "github.com/drone/spec/dist/go"
	"github.com/ghodss/yaml"
)

func Convert(opts commons.Opts, d []byte) (*harness.Pipeline, error) {
	jdata, err := yaml.YAMLToJSON(d)
	if err != nil {
		panic(err)
	}

	config, err := config.UnmarshalConfig(jdata)
	if err != nil {
		panic(err)
	}

	jobMap := make(map[string]*harness.Stage)
	for k, j := range config.Jobs {
		s, err := jobs.Convert(opts, j)
		if err != nil {
			return nil, err
		}
		jobMap[k] = s
	}

	p := &harness.Pipeline{
		Version: 1,
		Name:    "CI pipeline",
	}
	for k, w := range config.Workflows {
		if w.Workflow == nil {
			continue
		}

		for _, j := range w.Workflow.Jobs {
			if j.String == nil {
				// TODO: handle jobref map
				continue
			}

			if s, ok := jobMap[*j.String]; ok {
				s.Name = k
				p.Stages = append(p.Stages, s)
			}
		}
	}
	return p, nil
}
