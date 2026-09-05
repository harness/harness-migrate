// Copyright © 2026 Harness Inc.
// SPDX-License-Identifier: Apache-2.0

package migrateplugin

import (
	"net/http"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/transport/oauth2"
)

// oauth2BearerClient builds an http.Client that injects token as a bearer
// token, matching every provider's transport in the standalone commands.
func oauth2BearerClient(token string) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Scheme: oauth2.SchemeBearer,
			Source: oauth2.StaticTokenSource(&scm.Token{Token: token}),
		},
	}
}
