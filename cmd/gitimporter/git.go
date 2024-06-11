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

package gitimporter

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/gitimporter"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/types"

	"github.com/alecthomas/kingpin/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type gitImport struct {
	debug bool
	trace bool

	harnessEndpoint string
	harnessToken    string
	harnessSpace    string
	//harnessRepo     string

	filePath string
}

type UserInvite bool

func (c *gitImport) run(*kingpin.ParseContext) error {
	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	tracer_ := tracer.New()
	defer tracer_.Close()

	importUuid := uuid.New().String()
	importer := gitimporter.NewImporter(c.harnessSpace, c.harnessToken, c.filePath, importUuid, tracer_)

	tracer_.Log("starting operation with id: %s", importUuid)

	repositoriesImportOutput, err := importer.UploadZip()

	if err != nil {
		tracer_.LogError("encountered error uploading zip: %s", err)
		return err
	}

	if err := checkAndPerformUserInvite(repositoriesImportOutput, tracer_, importer); err != nil {
		return err
	}

	if err := importer.IsComplete(); err != nil {
		return err
	}
	return nil
}

func checkAndPerformUserInvite(repositoriesImportOutput *types.RepositoriesImportOutput, tracer_ tracer.Tracer, importer *gitimporter.Importer) error {
	if repositoriesImportOutput != nil && len(repositoriesImportOutput.Users.NotPresent) != 0 {
		tracer_.Log("Found users which are not in harness and are present in import data: ")
		tracer_.Log("%v", repositoriesImportOutput.Users.NotPresent)
		userFollowUp, err := doUserFollowUp()
		if err != nil {
			return err
		}
		if userFollowUp {
			err = importer.InviteUsers(repositoriesImportOutput.Users.NotPresent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func doUserFollowUp() (UserInvite, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please select one of the following options:")
	fmt.Println("1. Map missing user to yourself")
	fmt.Println("2. Invite missing users (needs admin permission for space)")
	fmt.Print("Enter your choice (1 or 2): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("You selected Option 1")
		return false, nil
	case "2":
		fmt.Println("You selected Option 2")
		return true, nil
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
		return false, fmt.Errorf("invalid option selected for user invite")
	}
}

func registerGitImporter(app *kingpin.CmdClause) {
	c := new(gitImport)

	cmd := app.Command("git-import", "import git data into harness/gitness").
		Hidden().
		Action(c.run)

	cmd.Arg("filePath", "location of the zip file").
		Required().
		Envar("HARNESS_FILEPATH").
		StringVar(&c.filePath)

	cmd.Flag("harnessEndpoint", "url of harness code host").
		Default("https://app.harness.io/").
		Envar("HARNESS_HOST").
		StringVar(&c.harnessEndpoint)

	cmd.Flag("token", "harness api token").
		Required().
		Envar("HARNESS_TOKEN").
		StringVar(&c.harnessToken)

	cmd.Flag("space", "harness path where import should take place. Example: account/org/project").
		Required().
		Envar("HARNESS_TOKEN").
		StringVar(&c.harnessSpace)

	// cmd.Flag("repo", "Required in case of single repo import which already exists.").
	//	Envar("HARNESS_REPO").
	//	StringVar(&c.harnessRepo)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
