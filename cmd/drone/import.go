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
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/alecthomas/kingpin/v2"
	"golang.org/x/exp/slog"
)

type importCommand struct {
	debug bool
	file  string

	harnessToken   string
	harnessAccount string
	harnessOrg     string
	harnessAddress string

	repositoryList string

	githubToken    string
	githubURL      string
	gitlabToken    string
	gitlabURL      string
	bitbucketToken string
	bitbucketURL   string
	skipVerify     bool

	// repo connector
	repoConn string

	// kube connector
	kubeName string
	kubeConn string

	// docker connector
	dockerConn string

	downgrade bool
}

func (c *importCommand) run(*kingpin.ParseContext) error {
	log := util.CreateLogger(c.debug)
	ctx := slog.NewContext(context.Background(), log)

	if c.repoConn == "" && (c.gitlabToken == "" && c.githubToken == "" && c.bitbucketToken == "") {
		return errors.New("either specify a repo connector or a gitlab/github/bitbucket token")
	}

	org, err := c.readAndUnmarshalOrg(c.file, log)
	if err != nil {
		return err
	}

	importer, err := c.createImporter(log, ctx)
	if err != nil {
		return err
	}

	return importer.Import(ctx, org)
}

func (c *importCommand) readAndUnmarshalOrg(filename string, log *slog.Logger) (*types.Org, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Error("cannot read data file", nil)
		return nil, err
	}

	org := new(types.Org)
	if err := json.Unmarshal(data, org); err != nil {
		log.Error("cannot unmarshal data file", nil)
		return nil, err
	}

	return org, nil
}

func (c *importCommand) createImporter(log *slog.Logger, ctx context.Context) (*migrate.Importer, error) {
	tracer_ := tracer.New()
	defer tracer_.Close()

	importer := util.CreateImporter(
		c.harnessAccount,
		c.harnessOrg,
		c.harnessToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.harnessAddress,
	)

	importer.Tracer = tracer_
	importer.Downgrade = c.downgrade
	importer.DockerConn = c.dockerConn
	importer.KubeName = c.kubeName
	importer.KubeConn = c.kubeConn

	if c.repositoryList != "" {
		importer.RepositoryList = strings.Split(c.repositoryList, ",")
	}

	if c.repoConn == "" {
		client, user, err := c.createAndVerifyScmClient(log, ctx)
		if err != nil {
			return nil, err
		}
		importer.ScmLogin = user.Login
		importer.ScmClient = client
	} else {
		importer.RepoConn = c.repoConn
	}

	return importer, nil
}

func (c *importCommand) createAndVerifyScmClient(log *slog.Logger, ctx context.Context) (*scm.Client, *scm.User, error) {
	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.githubURL,
		c.gitlabURL,
		c.bitbucketURL,
		c.skipVerify,
	)

	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return nil, nil, err
	}

	log.Debug("verified user and token", slog.String("user", user.Login))
	return client, user, nil
}

// helper function registers the import command.
func registerImport(app *kingpin.CmdClause) {
	c := new(importCommand)

	cmd := app.Command("import", "import drone data").
		Action(c.run)

	cmd.Arg("file", "data file to import").
		StringVar(&c.file)

	cmd.Flag("harness-account", "harness account").
		Required().
		Envar("HARNESS_ACCOUNT").
		StringVar(&c.harnessAccount)

	cmd.Flag("harness-org", "harness organization").
		Required().
		Envar("HARNESS_ORG").
		StringVar(&c.harnessOrg)

	cmd.Flag("harness-token", "harness token").
		Required().
		Envar("HARNESS_TOKEN").
		StringVar(&c.harnessToken)

	cmd.Flag("harness-address", "harness address").
		Envar("HARNESS_ADDRESS").
		Default("https://app.harness.io").
		StringVar(&c.harnessAddress)

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("github-url", "github url").
		Envar("GITHUB_URL").
		StringVar(&c.githubURL)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("gitlab-url", "gitlab url").
		Envar("GITLAB_URL").
		StringVar(&c.gitlabURL)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("bitbucket-url", "bitbucket url").
		Envar("BITBUCKET_URL").
		StringVar(&c.bitbucketURL)

	cmd.Flag("skip-tls-verify", "skip TLS verification for SCM").
		Envar("SKIP_TLS_VERIFY").
		BoolVar(&c.skipVerify)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("downgrade", "downgrade to the legacy yaml format").
		Default("true").
		BoolVar(&c.downgrade)

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

	cmd.Flag("repository-list", "optional list of repositories to import").
		Envar("REPOSITORY_LIST").
		StringVar(&c.repositoryList)
}
