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
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/migrate"
	"github.com/harness/harness-migrate/internal/types"
)

const commentFields = "values.id,values.type,values.parent.id,values.content.raw,values.user.account_id,values.user.display_name,values.created_on,values.updated_on,values.inline.*"

func (e *Export) convertPRCommentsList(from []codeComment, prNumber int, repoSlug string) []*types.PRComment {
	var to []*types.PRComment
	for _, v := range from {
		to = append(to, e.convertPRComment(v, prNumber, repoSlug))
	}
	return to
}

func (e *Export) convertPRComment(from codeComment, prNumber int, repo string) *types.PRComment {
	if from.Type != "pullrequest_comment" {
		return nil
	}
	var metadata *types.CodeComment

	isReply := false
	if from.Parent.ID != 0 {
		isReply = true
	}

	if from.Inline != nil {
		hunkHeader, err := extractHunkInfo(from.Inline, isReply)
		if err != nil {
			e.fileLogger.Log(fmt.Sprintf("Importing code comment %d on PR %d of repo %s as a PR comment: %v", from.ID, prNumber, repo, err))
		} else {
			metadata = &types.CodeComment{
				Path:         from.Inline.Path,
				CodeSnippet:  extractSnippetInfo(from.Inline.ContextLines),
				Side:         getSide(from.Inline),
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
				ID:   from.User.AccountID,
				Name: from.User.DisplayName,
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
	if len(lines) < 4 {
		return types.Hunk{}
	}
	return types.Hunk{
		Header: lines[2],
		Lines:  lines[3:],
	}
}

func getSide(inline *inline) string {
	if inline.From != nil {
		return "OLD"
	}
	return "NEW"
}

func extractHunkInfo(inline *inline, isReply bool) (string, error) {
	// reply comments dont need hunkheader, only parents do.
	if isReply {
		return "", nil
	}

	var (
		oldLine int
		oldSpan int
		newLine int
		newSpan int
	)

	// get the first 4k of the diff if it's too large.
	context := inline.ContextLines
	if len(context) > 4096 {
		context = context[:4096]
	}

	scan := bufio.NewScanner(strings.NewReader(context))

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

	oldLine = hunk.OldLine
	if inline.From != nil {
		oldLine = *inline.From
		oldSpan = 1
	}
	newLine = hunk.NewLine
	if inline.To != nil {
		newLine = *inline.To
		newSpan = 1
	}

	hunkHeader, err := common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, "")
	if err != nil {
		return "", fmt.Errorf("invalid hunk header values: %w", err)
	}

	return hunkHeader, nil
}
