package run

import (
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/utils"

	harness "github.com/drone/spec/dist/go"
)

const (
	scriptType     = "script"
	backgroundType = "background"
	namePrefix     = "run"
)

func Convert(r config.Deploy) (*harness.Step, error) {
	if r.String != nil && *r.String != "" {
		return &harness.Step{
			Name: namePrefix,
			Type: scriptType,
			Spec: harness.StepExec{
				Run: *r.String,
			},
		}, nil
	}

	if r.Confi == nil {
		return nil, nil
	}

	background := isBackground(r.Confi)
	envs := utils.ConvertEnvs(r.Confi.Environment)
	cmd := r.Confi.Command
	name := getName(r.Confi)
	shell := getShell(r.Confi)

	if background {
		return &harness.Step{
			Name: name,
			Type: backgroundType,
			Spec: harness.StepBackground{
				Envs:  envs,
				Run:   cmd,
				Shell: shell,
			},
		}, nil
	}
	return &harness.Step{
		Name: name,
		Type: scriptType,
		Spec: harness.StepExec{
			Envs:  envs,
			Run:   cmd,
			Shell: shell,
		},
	}, nil
}

func getShell(c *config.Confi) string {
	shell := ""
	if c.Shell != nil && *c.Shell != "" {
		shell = *c.Shell
	}
	return shell
}

func getName(c *config.Confi) string {
	name := namePrefix
	if c.Name != nil && *c.Name != "" {
		name = *c.Name
	}
	return name
}

func isBackground(c *config.Confi) bool {
	background := false
	if c.Background != nil && *c.Background {
		background = true
	}
	return background
}
