package stash

import (
	"context"
	"github.com/alecthomas/kingpin/v2"
	scmstash "github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"github.com/harness/harness-migrate/cmd/util"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/migrate/stash"
	"github.com/harness/harness-migrate/internal/tracer"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
)

type exportCommand struct {
	debug bool
	trace bool
	file  string

	stashOrg   string
	stashUser  string
	stashToken string
	stashUrl   string
}

func (c *exportCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// create the stash client (url, token, org)
	client, err := scmstash.New(c.stashUrl)
	if err != nil {
		return err
	}
	// provide a custom http.Client with a transport
	// that injects the private stash token through
	// the PRIVATE_TOKEN header variable.
	t := &transport.BasicAuth{
		Base:     nil,
		Username: c.stashUser,
		Password: c.stashToken,
	}

	client.Client = &http.Client{
		Transport: t,
	}

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	// extract the data
	e := stash.Export{
		Stash:    client,
		Tracer:   tracer_,
		StashOrg: c.stashOrg,
	}

	exporter := common.Exporter{Exporter: e, ZipLocation: c.file}
	exporter.Export(ctx)
	return nil
}

// helper function registers the export command
func registerGit(app *kingpin.CmdClause) {
	c := new(exportCommand)

	cmd := app.Command("git", "export stash git data").
		Hidden().
		Action(c.run)

	cmd.Arg("save", "save the output to a folder").
		Default("harness").
		StringVar(&c.file)

	cmd.Flag("host", "stash host url").
		Required().
		Envar("stash_HOST").
		StringVar(&c.stashUrl)

	cmd.Flag("org", "stash organization").
		Required().
		Envar("stash_ORG").
		StringVar(&c.stashOrg)

	cmd.Flag("token", "stash token").
		Required().
		Envar("stash_TOKEN").
		StringVar(&c.stashToken)

	cmd.Flag("username", "stash username").
		Required().
		Envar("stash_USERNAME").
		StringVar(&c.stashToken)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)

	cmd.Flag("trace", "enable trace logging").
		BoolVar(&c.trace)
}

// defaultTransport provides a default http.Transport.
// If skip verify is true, the transport will skip ssl verification.
// Otherwise, it will append all the certs from the provided path.
func defaultTransport(proxy string) http.RoundTripper {
	if proxy == "" {
		return &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
	}

	proxyURL, _ := url.Parse(proxy)

	return &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
}
