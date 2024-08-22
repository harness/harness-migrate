// Copyright 2024 Harness, Inc.
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

package jenkinsxml

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/drone/go-convert/convert/harness/downgrader"
	"github.com/drone/go-convert/convert/jenkinsxml"

	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/kingpin/v2"
)

type convertCommand struct {
	input  string
	output string

	address  string
	username string
	password string

	name       string
	proj       string
	org        string
	repoName   string
	repoConn   string
	kubeName   string
	kubeConn   string
	dockerConn string

	downgrade bool

	color bool
	theme string
}

func (c *convertCommand) run(ctx *kingpin.ParseContext) error {
	var before []byte
	var err error

	// detect URI input if passed
	u, _ := url.ParseRequestURI(c.input)

	// download file if the URL is absolute (not a local path)
	if u.IsAbs() {
		// download the jenkins xml file
		before, err = httpGetRequest(c)
		if err != nil {
			return err
		}
	} else {
		// open the jenkins xml file
		before, err = os.ReadFile(c.input)
		if err != nil {
			return err
		}
	}

	// convert the pipeline yaml from the drone
	// format to the harness yaml format.
	converter := jenkinsxml.New(
		jenkinsxml.WithDockerhub(c.dockerConn),
		jenkinsxml.WithKubernetes(c.kubeName, c.kubeConn),
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
		return os.WriteFile(c.output, after, 0644)
	}

	if c.color {
		// highlight and write to stdout
		return quick.Highlight(os.Stdout, string(after), "yaml", "terminal", c.theme)
	} else {
		// write to stdout
		os.Stdout.Write(after)
		return nil
	}
}

// httpGetRequest sends an HTTP GET request to the input URL with provided
// username and password
//
// TODO: this should probably be moved under 'internal'
func httpGetRequest(c *convertCommand) ([]byte, error) {
	client := http.Client{
		Timeout: time.Duration(1) * time.Second,
	}
	req, err := http.NewRequest("GET", c.input, nil)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	req.Header.Add("Authorization", "Basic "+auth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// helper function registers the convert command
func registerConvert(app *kingpin.CmdClause) {
	c := new(convertCommand)

	tty := isatty.IsTerminal(os.Stdout.Fd())

	cmd := app.Command("convert", "convert a jenkins XML job").
		Action(c.run)

	cmd.Arg("input", "path to jenkins XML job").
		Default("config.xml").
		StringVar(&c.input)

	cmd.Arg("output", "path to save the converted yaml").
		StringVar(&c.output)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		Default("true").
		BoolVar(&c.downgrade)

	cmd.Flag("jenkins-username", "jenkins username").
		Envar("JENKINS_USERNAME").
		StringVar(&c.username)

	cmd.Flag("jenkins-password", "jenkins password").
		Envar("JENKINS_PASSWORD").
		StringVar(&c.password)

	cmd.Flag("color", "print with syntax highlighting").
		Envar("COLOR").
		Default(fmt.Sprint(tty)).
		BoolVar(&c.color)

	cmd.Flag("theme", "syntax highlighting theme").
		Envar("THEME").
		Default("friendly").
		StringVar(&c.theme)

	cmd.Flag("org", "harness organization").
		Default("default").
		StringVar(&c.org)

	cmd.Flag("project", "harness project").
		Default("default").
		StringVar(&c.proj)

	cmd.Flag("pipeline", "harness pipeline name").
		Default("default").
		StringVar(&c.name)

	cmd.Flag("repo-connector", "repository connector").
		StringVar(&c.repoConn)

	cmd.Flag("repo-name", "repository name").
		StringVar(&c.repoName)

	cmd.Flag("kube-connector", "kubernetes connector").
		StringVar(&c.kubeConn)

	cmd.Flag("kube-namespace", "kubernetes namespace").
		StringVar(&c.kubeName)

	cmd.Flag("docker-connector", "dockerhub connector").
		StringVar(&c.dockerConn)
}
