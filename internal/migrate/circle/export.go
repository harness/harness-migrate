// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package circle

import (
	"context"

	"github.com/drone/go-convert/convert/circle"

	"github.com/harness/harness-migrate/internal/migrate/circle/client"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
)

// Exporter exports data from Circle.
type Exporter struct {
	Circle    client.Client
	CircleOrg string

	Tracer tracer.Tracer
}

// Export exports Circle data.
func (m *Exporter) Export(ctx context.Context) (*types.Org, error) {

	m.Tracer.Start("export organization")

	// find the circle organization by uuid
	srcOrg, err := m.Circle.FindOrgID(m.CircleOrg)
	if err != nil {
		return nil, err
	}

	// convert the circle org to the common format.
	dstOrg := &types.Org{
		Name: srcOrg.Name,
	}

	m.Tracer.Stop("export organization %s [done]", srcOrg.Name)

	// retrieve a list of all circle projects in the organization.
	srcProjects, err := m.Circle.ListProjects(srcOrg.ID)
	if err != nil {
		return nil, err
	}

	// convert each circle project to a harness project.
	for _, srcProject := range srcProjects {

		m.Tracer.Start("export project %s", srcProject.Name)

		// get a list of recent pipeline executions
		pipelines, err := m.Circle.ListPipelines(srcProject.Slug)
		if err != nil {
			return nil, err
		}
		if len(pipelines) == 0 {
			continue
		}

		// use the latest pipeline as our reference
		srcPipeline := pipelines[0]

		// find the circle pipeline
		config, err := m.Circle.FindConfig(srcPipeline.ID)
		if err != nil {
			return nil, err
		}

		// convert the circle project to a common format.
		dstProject := &types.Project{
			Name: srcProject.Name,
			Yaml: []byte(config.Source),
		}

		converter := circle.New()
		newYaml, err := converter.ConvertString(config.Source)
		if err != nil {
			return nil, err
		}
		dstProject.Yaml = newYaml
		// extract the repository details from the pipeline.
		switch {
		case srcPipeline.Params.Gitlab.ID != "":
			dstProject.Type = "gitlab"
			dstProject.Repo = srcPipeline.Params.Gitlab.DeepLink + ".git"
			dstProject.Branch = srcPipeline.Params.Gitlab.BranchDefault
		default:
			continue
		}

		// find circle environment variables
		srcEnvs, err := m.Circle.ListEnvs(srcProject.Slug)
		if err != nil {
			return nil, err
		}

		// for each environment variable
		for _, srcEnv := range srcEnvs {
			dstSecret := &types.Secret{
				Name:  srcEnv.Name,
				Desc:  srcEnv.Value,
				Value: srcEnv.Value,
			}
			// append the environment variable as a secret
			dstProject.Secrets = append(dstProject.Secrets, dstSecret)
		}

		// append projects to the org
		dstOrg.Projects = append(dstOrg.Projects, dstProject)

		m.Tracer.Stop("export project %s [done]", srcProject.Name)
	}

	return dstOrg, nil
}
