package drone

import (
	"net/http"
	"os"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/migrate/drone"

	"golang.org/x/exp/slog"
)

// helper function creates a logger
func createLogger(debug bool) slog.Logger {
	opts := new(slog.HandlerOptions)
	if debug {
		opts.Level = slog.DebugLevel
	}
	return slog.New(
		opts.NewTextHandler(os.Stdout),
	)
}

// helper function creates an importer
func createImporter(harnessAccount, harnessOrg, harnessToken, githubToken, gitlabToken, bitbucketToken, harnessAddress string) *drone.Importer {
	importer := &drone.Importer{
		Harness:    harness.New(harnessAccount, harnessToken, harness.WithAddress(harnessAddress)),
		HarnessOrg: harnessOrg,
	}
	switch {
	case githubToken != "":
		importer.ScmType = "github"
		importer.ScmToken = githubToken
	case gitlabToken != "":
		importer.ScmType = "gitlab"
		importer.ScmToken = gitlabToken
	case bitbucketToken != "":
		importer.ScmType = "bitbucket"
		importer.ScmToken = bitbucketToken
	}
	return importer
}

// helper function creates an scm client
func createClient(githubToken, gitlabToken, bitbucketToken string) *scm.Client {
	var client *scm.Client
	switch {
	case githubToken != "":
		// create the gitHub client and create an oauth2
		// transport to authenticate requests using the token
		client = github.NewDefault()
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: githubToken,
					},
				),
			},
		}
	case gitlabToken != "":
		// create the gitlab client and create an oauth2
		// transport to authenticate requests using the token
		client = gitlab.NewDefault()
		client.Client = &http.Client{
			Transport: &transport.PrivateToken{
				Token: gitlabToken,
			},
		}
	case bitbucketToken != "":
		// create the bitbucket client and create an oauth2
		// transport to authenticate requests using the token
		client = bitbucket.NewDefault()
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: bitbucketToken,
					},
				),
			},
		}
	}
	return client
}
