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

package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/drone/go-convert/convert/drone"
	"github.com/drone/go-convert/convert/harness/downgrader"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/types"

	"gopkg.in/yaml.v2"

	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/kingpin/v2"
	"github.com/drone/funcmap"
	"github.com/mattn/go-isatty"

	_ "embed"
)

//go:embed template.tf
var defaultTmpl string

type terraformCommand struct {
	input  string
	output string
	tmpl   string

	account         string
	endpoint        string
	organization    string
	providerSource  string
	providerVersion string
	repoConn        string
	kubeName        string
	kubeConn        string
	dockerConn      string

	downgrade bool

	color bool
	theme string
}

func (c *terraformCommand) run(ctx *kingpin.ParseContext) error {
	// open the export file and unmarshal
	f, err := os.ReadFile(c.input)
	if err != nil {
		return err
	}
	org := new(types.Org)
	if err := json.Unmarshal(f, org); err != nil {
		return err
	}

	// create the template input
	in := input{
		Org: org,
		Auth: auth{
			Endpoint: c.endpoint,
		},
		Account: account{
			ID:           c.account,
			Organization: c.organization,
		},
		Connectors: connectors{
			Repo: c.repoConn,
		},
		Provider: provider{
			Source:  c.providerSource,
			Version: c.providerVersion,
		},
	}

	tmpl := defaultTmpl

	// if the user provides an alternate template,
	// read and parse.
	if c.tmpl != "" {
		t, err := ioutil.ReadFile(c.tmpl)
		if err != nil {
			return err
		}
		tmpl = string(t)
	}

	// convert all yaml into v1 or v0 format
	converter := drone.New(
		drone.WithDockerhub(c.dockerConn),
		drone.WithKubernetes(c.kubeName, c.kubeConn),
	)
	for _, project := range org.Projects {
		// convert to v1
		convertedYaml, err := converter.ConvertBytes(project.Yaml)
		if err != nil {
			return err
		}
		// downgrade to v0 if needed
		if c.downgrade {
			d := downgrader.New(
				downgrader.WithCodebase(project.Name, c.repoConn),
				downgrader.WithDockerhub(c.dockerConn),
				downgrader.WithKubernetes(c.kubeName, c.kubeConn),
				downgrader.WithName(project.Name),
				downgrader.WithOrganization(c.organization),
				downgrader.WithProject(slug.Create(project.Name)),
			)
			convertedYaml, err = d.Downgrade(convertedYaml)
			if err != nil {
				return nil
			}
		}
		project.Yaml = convertedYaml
	}

	// parse the terraform template
	t, err := template.New("_").
		Funcs(funcmap.Funcs).
		Funcs(template.FuncMap{
			"slugify": slug.Create,
			// TODO: drone/funcmap has its own toYaml which is incompatible,
			//       can we add fromYaml and make toYaml compatible?
			"fromYaml": func(in []byte) map[string]interface{} {
				out := map[string]interface{}{}
				yaml.Unmarshal(in, out)
				return out
			},
			"toYaml": func(in interface{}) string {
				out, _ := yaml.Marshal(in)
				return string(out)
			},
		}).
		Parse(tmpl)
	if err != nil {
		return err
	}
	// generate the terraform file from template
	var buf bytes.Buffer
	if err := t.Execute(&buf, &in); err != nil {
		return err
	}

	// write the tf to the output file
	if c.output != "" && c.output != "-" {
		return os.WriteFile(c.output, buf.Bytes(), 0644)
	}

	if c.color {
		// highlight and write to stdout
		return quick.Highlight(os.Stdout, buf.String(), "hcl", "terminal", c.theme)
	} else {
		// write to stdout
		os.Stdout.Write(buf.Bytes())
		return nil
	}
}

// Register registers the terraform generation command.
func Register(app *kingpin.Application) {
	c := new(terraformCommand)

	tty := isatty.IsTerminal(os.Stdout.Fd())

	cmd := app.Command("terraform", "generate terraform script from data export file").
		Action(c.run)

	cmd.Arg("input", "path to the data export file").
		Default("export.json").
		StringVar(&c.input)

	cmd.Arg("output", "path to save the terraform file").
		StringVar(&c.output)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		// TODO: unhide when the pipeline tf resource supports v1 yaml,
		//       until then, all pipelines must be downgraded to v0
		Hidden().
		Default("true").
		BoolVar(&c.downgrade)

	cmd.Flag("template", "path to the terraform template").
		StringVar(&c.tmpl)

	cmd.Flag("account", "harness account ID").
		StringVar(&c.account)

	cmd.Flag("endpoint", "harness endpoint").
		Default("https://app.harness.io/gateway").
		StringVar(&c.endpoint)

	cmd.Flag("org", "harness organization").
		Default("default").
		StringVar(&c.organization)

	cmd.Flag("color", "print with syntax highlighting").
		Envar("COLOR").
		Default(fmt.Sprint(tty)).
		BoolVar(&c.color)

	cmd.Flag("theme", "syntax highlighting theme").
		Envar("THEME").
		Default("friendly").
		StringVar(&c.theme)

	cmd.Flag("provider-source", "harness terraform provider source").
		Default("harness/harness").
		StringVar(&c.providerSource)

	cmd.Flag("provider-version", "harness terraform provider version").
		Default("0.19.1").
		StringVar(&c.providerVersion)

	cmd.Flag("kube-connector", "kubernetes connector").
		Envar("KUBE_CONN").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernetes namespace").
		Envar("KUBE_NAMESPACE").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		Default("").
		StringVar(&c.dockerConn)

	cmd.Flag("repo-connector", "repository connector").
		Default("").
		StringVar(&c.repoConn)
}

type (
	input struct {
		Account    account
		Auth       auth
		Connectors connectors
		Org        *types.Org
		Provider   provider
	}

	account struct {
		ID           string
		Key          string
		Organization string
	}

	auth struct {
		Endpoint string
	}

	connectors struct {
		Repo string
	}

	provider struct {
		Source  string
		Version string
	}
)
