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

package bitbucket

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
)

const commentFields = "values.id,values.type,values.content.raw,values.user.uuid,values.user.display_name,values.created_on,values.updated_on,values.inline.*"

func convertPRCommentsList(from []codeComment, prNumber int, repoSlug string) []*types.PRComment {
	var to []*types.PRComment
	for _, v := range from {
		to = append(to, convertPRComment(v, prNumber, repoSlug))
	}
	return to
}

func convertPRComment(from codeComment, prNumber int, repo string) *types.PRComment {
	if from.Type != "pullrequest_comment" {
		return nil
	}
	var metadata *types.CodeComment

	if from.Inline != nil {
		hunkHeader, err := extractHunkInfo(from.Inline)
		if err != nil {
			log.Default().Printf("Failed to export code comment %d on PR %d of repo %s as a PR comment: %v", from.ID, prNumber, repo, err)
		} else {
			metadata = &types.CodeComment{
				Path:         from.Inline.Path,
				CodeSnippet:  extractSnippetInfo(from.Inline.ContextLines),
				Side:         "NEW",
				HunkHeader:   hunkHeader,
				SourceSHA:    from.Inline.SrcRev,
				MergeBaseSHA: from.Inline.DestRev,
				Outdated:     from.Inline.Outdated,
			}
		}
	}

	return &types.PRComment{
		Comment: scm.Comment{
			ID:   from.ID,
			Body: from.Content.Raw,
			Author: scm.User{
				Login: from.User.UUID,
				Name:  from.User.DisplayName,
			},
			Created: from.CreatedOn,
			Updated: from.UpdatedOn,
		},
		ParentID:    from.Parent.ID,
		CodeComment: metadata,
	}
}

func extractSnippetInfo(diffHunk string) types.Hunk {
	lines := strings.Split(diffHunk, "\n")
	return types.Hunk{
		Header: lines[0],
		Lines:  lines[1:],
	}
}

func extractHunkInfo(inline *inline) (string, error) {
	scan := bufio.NewScanner(strings.NewReader(inline.ContextLines))

	// Look for the line containing the hunk header
	var hunkHeader string
	for scan.Scan() {
		line := scan.Text()
		if strings.HasPrefix(line, "@@") {
			hunkHeader = line
			break
		}
	}

	if hunkHeader == "" {
		return "", errors.New("hunk header missing")
	}

	hunk, ok := migrate.ParseDiffHunkHeader(hunkHeader)
	if !ok {
		return "", fmt.Errorf("invalid diff hunk header: %s", hunkHeader)
	}

	return common.FormatHunkHeader(hunk.OldLine, hunk.OldSpan, hunk.NewLine, hunk.NewSpan, ""), nil
}
