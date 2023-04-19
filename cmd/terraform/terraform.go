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

	"github.com/harness/harness-migrate/internal/slug"
	"github.com/harness/harness-migrate/internal/types"

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

	account  string
	endpoint string
	token    string

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
			Token:    c.token,
		},
		Account: account{
			ID: c.account,
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

	// parse the terraform template
	t, err := template.New("_").
		Funcs(funcmap.Funcs).
		Funcs(template.FuncMap{
			"slugify": slug.Create,
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

	cmd.Flag("template", "path to the terraform template").
		StringVar(&c.tmpl)

	cmd.Flag("account", "harness account").
		StringVar(&c.account)

	cmd.Flag("endpoint", "harness endpoint").
		Default("https://app.harness.io/gateway").
		StringVar(&c.endpoint)

	cmd.Flag("token", "harness token").
		StringVar(&c.token)

	cmd.Flag("color", "print with syntax highlighting").
		Envar("COLOR").
		Default(fmt.Sprint(tty)).
		BoolVar(&c.color)

	cmd.Flag("theme", "syntax highlighting theme").
		Envar("THEME").
		Default("friendly").
		StringVar(&c.theme)
}

type (
	input struct {
		Account account
		Auth    auth
		Org     *types.Org
	}

	account struct {
		ID  string
		Key string
	}

	auth struct {
		Endpoint string
		Token    string
	}
)
