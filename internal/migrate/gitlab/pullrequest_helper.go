// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlab

import (
	"fmt"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/types"
)

func convertPR(from mergeRequest) *types.PRResponse {
	var labels []scm.Label
	for _, label := range from.Labels {
		labels = append(labels, scm.Label{
			Name: label,
		})
	}

	return &types.PRResponse{
		PullRequest: scm.PullRequest{
			Number: from.Number,
			Title:  from.Title,
			Body:   from.Desc,
			Sha:    from.Sha,
			Ref:    fmt.Sprintf("refs/merge-requests/%d/head", from.Number),
			Source: from.SourceBranch,
			Target: from.TargetBranch,
			Link:   from.Link,
			Draft:  from.WorkInProgress,
			Closed: from.State != "opened",
			Merged: from.State == "merged",
			Author: scm.User{
				Name:   from.Author.Name,
				Login:  from.Author.Username,
				Avatar: from.Author.Avatar,
			},
			Created: from.Created,
			Updated: from.Updated,
			Labels:  labels,
			Head: scm.Reference{
				Sha:  from.DiffRefs.HeadSha,
				Name: from.SourceBranch,
				Path: scm.ExpandRef(from.SourceBranch, "refs/heads"),
			},
			Base: scm.Reference{
				Sha:  from.DiffRefs.BaseSha,
				Name: from.TargetBranch,
				Path: scm.ExpandRef(from.TargetBranch, "refs/heads"),
			},
		},
	}
}
