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

package drone

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/kingpin/v2"
	"github.com/drone/go-convert/convert/drone"
	"github.com/drone/go-convert/convert/harness/downgrader"
	"github.com/mattn/go-isatty"
)

type convertCommand struct {
	input  string
	output string

	name       string
	proj       string
	org        string
	repoName   string
	repoConn   string
	kubeName   string
	kubeConn   string
	dockerConn string

	orgSecrets string

	downgrade   bool
	beforeAfter bool

	color  bool
	theme  string
	format string
}

func (c *convertCommand) run(ctx *kingpin.ParseContext) error {
	// open the drone yaml
	before, err := ioutil.ReadFile(c.input)
	if err != nil {
		return err
	}

	// convert the pipeline yaml from the drone
	// format to the harness yaml format.
	converter := drone.New(
		drone.WithDockerhub(c.dockerConn),
		drone.WithKubernetes(c.kubeName, c.kubeConn),
		drone.WithOrgSecrets(
			strings.Split(c.orgSecrets, ",")...,
		),
	)
	after, err := converter.ConvertBytes(before)
	if err != nil {
		return err
	}

	// downgrade from the v1 harness yaml format
	// to the v0 harness yaml format.
	if c.downgrade {
		// downgrade to the v0 yaml
		d := downgrader.New(
			downgrader.WithCodebase(c.repoName, c.repoConn),
			downgrader.WithDockerhub(c.dockerConn),
			downgrader.WithKubernetes(c.kubeName, c.kubeConn),
			downgrader.WithName(c.name),
			downgrader.WithOrganization(c.org),
			downgrader.WithProject(c.proj),
		)
		after, err = d.Downgrade(after)
		if err != nil {
			return err
		}
	}

	// write the converted yaml to the output file
	if c.output != "" && c.output != "-" {
		return ioutil.WriteFile(c.output, after, 0644)
	}

	// write the original yaml to the buffer
	if c.beforeAfter {
		// if the original yaml has separator and terminator
		// lines, strip these before showing the before / after
		before = bytes.TrimPrefix(before, []byte("---\n"))
		before = bytes.TrimSuffix(before, []byte("...\n"))
		before = bytes.TrimSuffix(before, []byte("..."))
		before = bytes.TrimSuffix(before, []byte("\n"))

		var buf bytes.Buffer
		buf.WriteString("---\n")
		buf.Write(before)
		buf.WriteString("\n---\n")
		buf.Write(after)
		buf.WriteString("...\n")

		// combine the before and after
		after = buf.Bytes()
	}

	if c.color {
		// hightlight and write to stdout
		return quick.Highlight(os.Stdout, string(after), "yaml", c.format, c.theme)
	} else {
		// write to stdout
		os.Stdout.Write(after)
		return nil
	}
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	// determine if tty
	tty := isatty.IsTerminal(os.Stdout.Fd())

	cmd := app.Command("convert", "converts a drone yaml").
		Action(c.run)

	cmd.Arg("input", "path to the drone yaml").
		Default(".drone.yml").
		StringVar(&c.input)

	cmd.Arg("output", "path to save the converted yaml").
		StringVar(&c.output)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		Default("true").
		BoolVar(&c.downgrade)

	cmd.Flag("before-after", "print the before and after").
		BoolVar(&c.beforeAfter)

	cmd.Flag("color", "print with syntax highlighting").
		Envar("COLOR").
		Default(fmt.Sprint(tty)).
		BoolVar(&c.color)

	cmd.Flag("theme", "syntax highlighting theme").
		Envar("THEME").
		Default("github").
		StringVar(&c.theme)

	cmd.Flag("formatter", "syntax highlighting formatter").
		Hidden().
		Envar("FORMAT").
		Default("terminal").
		StringVar(&c.format)

	cmd.Flag("org", "harness organization").
		Default("default").
		StringVar(&c.org)

	cmd.Flag("project", "harness project").
		Default("default").
		StringVar(&c.proj)

	cmd.Flag("pipeline", "harness pipeline name").
		Default("").
		StringVar(&c.name)

	cmd.Flag("repo-connector", "repository connector").
		Default("").
		StringVar(&c.repoConn)

	cmd.Flag("repo-name", "repository name").
		Default("").
		StringVar(&c.repoName)

	cmd.Flag("kube-connector", "kubernetes connector").
		Default("").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernetes namespace").
		Default("").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		Default("").
		StringVar(&c.dockerConn)

	cmd.Flag("org-secrets", "optional list of secrets for pipelines with organization secrets").
		Envar("ORG_SECRETS").
		StringVar(&c.orgSecrets)
}
