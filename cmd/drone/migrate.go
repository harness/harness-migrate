package drone

import (
	"context"

	"github.com/harness/harness-migrate/internal/migrate/drone"
	"github.com/harness/harness-migrate/internal/migrate/drone/repo"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/alecthomas/kingpin/v2"
	"golang.org/x/exp/slog"
)

type migrateCommand struct {
	debug bool
	trace bool

	Driver     string
	Datasource string
	namespace  string

	harnessToken   string
	harnessAccount string
	harnessOrg     string
	harnessAddress string

	githubToken    string
	gitlabToken    string
	bitbucketToken string
}

func (c *migrateCommand) run(*kingpin.ParseContext) error {
	// create the logger
	log := createLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	droneRepo, err := repo.NewRepository(c.Driver, c.Datasource)

	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	exporter := &drone.Exporter{
		Repository: droneRepo,
		Namespace:  c.namespace,
		Tracer:     tracer_,
	}
	data, err := exporter.Export(ctx)
	if err != nil {
		return err
	}

	importer := createImporter(
		c.harnessAccount,
		c.harnessOrg,
		c.harnessToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.harnessAddress,
	)

	importer.Tracer = tracer_
	// create scm client to verify the token
	// and retrieve the user id.
	client := createClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
	)

	// get the current user id.
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	// provide the user id to the importer. the user id
	// is required by the connector despite the fact that
	// it can be retrieved using the token itself (like we just did)
	importer.ScmLogin = user.Login

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	importer.ScmClient = client

	// execute the import routine.
	return importer.Import(ctx, data)
}

func registerMigrate(app *kingpin.CmdClause) {
	c := new(migrateCommand)

	cmd := app.Command("migrate", "migrate drone data to harness").
		Action(c.run)

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

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("namespace", "drone namespace").
		Required().
		Envar("DRONE_NAMESPACE").
		StringVar(&c.namespace)

	cmd.Flag("driver", "drone db type").
		Default("sqlite3").
		StringVar(&c.Driver)

	cmd.Flag("datasource", "drone database datasource").
		Envar("DRONE_DATABASE_DATASOURCE").
		Default("database.sqlite3").
		StringVar(&c.Datasource)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}
