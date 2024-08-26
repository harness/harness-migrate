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
	"strings"
	"text/template"

	"github.com/drone/funcmap"
	"github.com/drone/go-convert/convert/drone"
	"github.com/drone/go-convert/convert/harness/downgrader"
	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/kingpin/v2"
	"github.com/mattn/go-isatty"
	"gopkg.in/yaml.v2"

	_ "embed"
)

//go:embed default.tmpl
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

	downgrade  bool
	orgSecrets bool

	color bool
	theme string
}

func (c *terraformCommand) run(ctx *kingpin.ParseContext) error {
	org, err := c.readAndUnmarshal(c.input)
	if err != nil {
		return err
	}

	in := c.createTemplateInput(org)

	if err := c.convertYaml(org); err != nil {
		return err
	}

	tmpl := defaultTmpl
	if c.tmpl != "" {
		t, err := ioutil.ReadFile(c.tmpl)
		if err != nil {
			return err
		}
		tmpl = string(t)
	}

	t, err := c.parseTemplate(tmpl)
	if err != nil {
		return err
	}

	buf, err := c.generateTerraformFile(t, &in)
	if err != nil {
		return err
	}

	return c.writeTerraformFile(buf, c.output)
}

func (c *terraformCommand) readAndUnmarshal(input string) (*types.Org, error) {
	f, err := os.ReadFile(input)
	if err != nil {
		return nil, err
	}
	org := new(types.Org)
	if err := json.Unmarshal(f, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (c *terraformCommand) createTemplateInput(org *types.Org) input {
	return input{
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
		Selections: selections{
			OrgSecrets: c.orgSecrets,
		},
	}
}

func (c *terraformCommand) convertYaml(org *types.Org) error {
	// read all organization secrets
	var orgSecrets []string
	for _, secret := range org.Secrets {
		orgSecrets = append(orgSecrets, secret.Name)
	}

	converter := drone.New(
		drone.WithDockerhub(c.dockerConn),
		drone.WithKubernetes(c.kubeName, c.kubeConn),
		drone.WithOrgSecrets(orgSecrets...),
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
				return err
			}
		}
		project.Yaml = convertedYaml
	}
	return nil
}

func (c *terraformCommand) parseTemplate(tmpl string) (*template.Template, error) {
	return template.New("_").
		Funcs(funcmap.Funcs).
		Funcs(template.FuncMap{
			"slugify": slug.Create,
			// TODO: drone/funcmap has its own toYaml which is incompatible,
			//       can we add fromYaml and make toYaml compatible?
			"fromYaml": func(in []byte) map[string]interface{} {
				out := map[string]interface{}{}
				yaml.Unmarshal(in, &out)
				return out
			},
			"toYaml": func(in interface{}) string {
				out, _ := yaml.Marshal(in)
				return string(out)
			},
			"indent": func(in string, spaces int) string {
				prefix := strings.Repeat(" ", spaces)
				lines := strings.Split(in, "\n")
				for i, line := range lines {
					lines[i] = prefix + line
				}
				return strings.Join(lines, "\n")
			},
		}).
		Parse(tmpl)
}

func (c *terraformCommand) generateTerraformFile(t *template.Template, in *input) (bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, in); err != nil {
		return buf, err
	}
	return buf, nil
}

func (c *terraformCommand) writeTerraformFile(buf bytes.Buffer, output string) error {
	if output != "" && output != "-" {
		return os.WriteFile(output, buf.Bytes(), 0644)
	}

	if c.color {
		// highlight and write to stdout
		return quick.Highlight(os.Stdout, buf.String(), "hcl", "terminal", c.theme)
	} else {
		// write to stdout
		_, err := os.Stdout.Write(buf.Bytes())
		return err
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

	cmd.Flag("org-secrets", "generate organization secrets").
		Default("true").
		BoolVar(&c.orgSecrets)
}

type (
	input struct {
		Account    account
		Auth       auth
		Connectors connectors
		Org        *types.Org
		Provider   provider
		Selections selections
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

	selections struct {
		OrgSecrets bool
	}
)
