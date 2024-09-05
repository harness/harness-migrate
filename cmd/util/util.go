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

package util

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/tracer"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/harness/harness-migrate/internal/harness"
	"golang.org/x/exp/slog"
)

// CreateLogger helper function creates a logger
func CreateLogger(debug bool) *slog.Logger {
	opts := new(slog.HandlerOptions)
	if debug {
		opts.Level = slog.LevelDebug
	}
	return slog.New(
		opts.NewTextHandler(os.Stdout),
	)
}

// CreateImporter helper function creates an importer
func CreateImporter(harnessAccount, harnessOrg, harnessToken, githubToken, gitlabToken, bitbucketToken, harnessAddress string) *migrate.Importer {
	importer := &migrate.Importer{
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

// CreateClient helper function creates a scm client
func CreateClient(githubToken, gitlabToken, bitbucketToken, githubURL, gitlabURL, bitbucketURL string, skipVerify bool) *scm.Client {
	var client *scm.Client
	switch {
	case githubToken != "":
		if githubURL != "" {
			client, _ = github.New(githubURL)
		} else {
			client = github.NewDefault()
		}
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: githubToken,
					},
				),
				Base: defaultTransport(skipVerify),
			},
		}
	case gitlabToken != "":
		if gitlabURL != "" {
			client, _ = gitlab.New(gitlabURL)
		} else {
			client = gitlab.NewDefault()
		}
		client.Client = &http.Client{
			Transport: &transport.PrivateToken{
				Token: gitlabToken,
				Base:  defaultTransport(skipVerify),
			},
		}
	case bitbucketToken != "":
		if bitbucketURL != "" {
			client, _ = bitbucket.New(bitbucketURL)
		} else {
			client = bitbucket.NewDefault()
		}
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: bitbucketToken,
					},
				),
				Base: defaultTransport(skipVerify),
			},
		}
	}
	return client
}

// defaultTransport provides a default http.Transport. If
// skipVerify is true, the transport will skip ssl verification.
func defaultTransport(skipVerify bool) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify,
		},
	}
}

func CreateTracerWithLevel(debug bool) tracer.Tracer {
	tracer_ := tracer.New()
	if debug == true {
		tracer_.WithLevel(tracer.LogLevelDebug)
	}
	return tracer_
}
