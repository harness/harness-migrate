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

package harness

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	pathSeparator        = "/"
	encodedPathSeparator = "%252F"

	accountIdentifier = "accountIdentifier"
	projectIdentifier = "projectIdentifier"
	orgIdentifier     = "orgIdentifier"
	routingId         = "routingId"
)

func getQueryParamsFromRepoRef(repoRef string) (string, error) {
	params := url.Values{}
	s := strings.TrimSuffix(repoRef, "/+")
	repoRefParts := strings.Split(s, "/")
	// valid repoRef: "Acc/Repo", "Acc/Org/Repo", "Acc/Org/Projct/Repo"
	if len(repoRefParts) < 2 || len(repoRefParts) > 4 {
		return "", fmt.Errorf("repo ref %s segments is invalid, got %d want 2-4", repoRef, len(repoRefParts))
	}
	params.Set(accountIdentifier, repoRefParts[0])
	params.Set(routingId, repoRefParts[0])

	switch len(repoRefParts) {
	case 3:
		params.Set(orgIdentifier, repoRefParts[1])
	case 4:
		params.Set(orgIdentifier, repoRefParts[1])
		params.Set(projectIdentifier, repoRefParts[2])
	}

	return params.Encode(), nil
}
