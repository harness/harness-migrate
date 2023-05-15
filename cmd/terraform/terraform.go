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
	"errors"
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

//go:embed template.tf
var defaultTmpl string

type terraformCommand struct {
	input  string
	output string
	tmpl   string

	// harness info
	harnessAccount         string
	harnessOrg             string
	harnessAddress         string
	harnessProviderSource  string
	harnessProviderVersion string

	// TODO: support passing list of repos
	//repositoryList string

	githubToken    string
	githubURL      string
	githubUser     string
	gitlabToken    string
	gitlabURL      string
	bitbucketToken string
	bitbucketURL   string

	// repo connector
	repoConn string

	// kube connector
	kubeName   string
	kubeConn   string
	dockerConn string

	// docker connector
	downgrade bool

	// output formatting
	color bool
	theme string
}

func (c *terraformCommand) run(ctx *kingpin.ParseContext) error {
	// validate flags
	if len(strings.Fields(fmt.Sprintf("%s %s %s", c.bitbucketToken, c.githubToken, c.gitlabToken))) > 1 {
		return errors.New("multiple repository tokens passed")
	}
	if c.bitbucketURL != "" && c.bitbucketToken == "" {
		return errors.New("--bitbucket-url requires flag --bitbucket-token")
	}
	if c.githubURL != "" && c.githubToken == "" {
		return errors.New("--github-url requires flag --github-token")
	}
	if c.githubToken != "" && c.githubUser == "" {
		return errors.New("--github-url requires flag --github-user")
	}
	if c.gitlabURL != "" && c.gitlabToken == "" {
		return errors.New("--gitlab-url requires flag --gitlab-token")
	}
	if c.repoConn == "" && (c.gitlabToken == "" && c.githubToken == "" && c.bitbucketToken == "") {
		return errors.New("either specify a repo connector or a token")
	}
	if c.repoConn != "" && (c.gitlabToken != "" || c.githubToken != "" || c.bitbucketToken != "") {
		return errors.New("token not required when passing repo connector")
	}

	// read exported json file
	org, err := c.readAndUnmarshal(c.input)
	if err != nil {
		return err
	}

	// create input for template
	in := c.createTemplateInput(org)

	// convert all yaml files
	if err := c.convertYaml(org); err != nil {
		return err
	}

	// read in conversion template
	tmpl := defaultTmpl
	if c.tmpl != "" {
		t, err := ioutil.ReadFile(c.tmpl)
		if err != nil {
			return err
		}
		tmpl = string(t)
	}

	// parse terraform template
	t, err := c.parseTemplate(tmpl)
	if err != nil {
		return err
	}

	// generate terraform file
	buf, err := c.generateTerraformFile(t, &in)
	if err != nil {
		return err
	}

	// write terraform file
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
	repoToken := ""
	repoType := ""
	repoURL := ""

	if c.bitbucketToken != "" {
		repoToken = c.bitbucketToken
		repoType = "bitbucket"
		repoURL = c.bitbucketURL
	} else if c.githubToken != "" {
		repoToken = c.githubToken
		repoType = "github"
		repoURL = c.githubURL
	} else if c.gitlabToken != "" {
		repoToken = c.gitlabToken
		repoType = "gitlab"
		repoURL = c.gitlabURL
	}

	return input{
		Org: org,
		Auth: auth{
			Endpoint: c.harnessAddress,
		},
		Account: account{
			ID:           c.harnessAccount,
			Organization: c.harnessOrg,
		},
		Connector: connector{
			Repo:  c.repoConn,
			Token: repoToken,
			Type:  repoType,
			URL:   repoURL,
			User:  c.githubUser,
		},
		Provider: provider{
			Source:  c.harnessProviderSource,
			Version: c.harnessProviderVersion,
		},
	}
}

func (c *terraformCommand) convertYaml(org *types.Org) error {
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
				downgrader.WithOrganization(c.harnessOrg),
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
		StringVar(&c.harnessAccount)

	cmd.Flag("endpoint", "harness endpoint").
		Default("https://app.harness.io/gateway").
		StringVar(&c.harnessAddress)

	cmd.Flag("harness-org", "harness organization").
		Required().
		Envar("HARNESS_ORG").
		StringVar(&c.harnessOrg)

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("github-url", "github url").
		Envar("GITHUB_URL").
		StringVar(&c.githubURL)

	cmd.Flag("github-user", "github username associated with token").
		Envar("GITHUB_USER").
		StringVar(&c.githubUser)

	cmd.Flag("gitlab-token", "gitlab token").
		// TODO: add gitlab support to template
		Hidden().
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("gitlab-url", "gitlab url").
		// TODO: add gitlab support to template
		Hidden().
		Envar("GITLAB_URL").
		StringVar(&c.gitlabURL)

	cmd.Flag("bitbucket-token", "bitbucket token").
		// TODO: add bitbucket support to template
		Hidden().
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("bitbucket-url", "bitbucket url").
		// TODO: add bitbucket support to template
		Hidden().
		Envar("BITBUCKET_URL").
		StringVar(&c.bitbucketURL)

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
		StringVar(&c.harnessProviderSource)

	cmd.Flag("provider-version", "harness terraform provider version").
		Default("0.19.1").
		StringVar(&c.harnessProviderVersion)

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
		Account   account
		Auth      auth
		Connector connector
		Org       *types.Org
		Provider  provider
	}

	account struct {
		ID           string
		Key          string
		Organization string
	}

	auth struct {
		Endpoint string
	}

	connector struct {
		Repo  string
		Token string
		Type  string
		URL   string
		User  string
	}

	provider struct {
		Source  string
		Version string
	}
)
