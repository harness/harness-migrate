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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func convertPRCommentsList(from []codeComment) []*types.PRComment {
	var to []*types.PRComment
	for _, v := range from {
		to = append(to, convertPRComment(v))
	}
	return to
}

func convertPRComment(from codeComment) *types.PRComment {
	if from.Type != "pullrequest_comment" {
		return nil
	}
	var metadata *types.CodeComment

	if from.Inline != nil { // Check if the comment is on a file
		metadata = &types.CodeComment{
			Path:         from.Inline.Path,
			CodeSnippet:  extractSnippetInfo(from.Inline.ContextLines),
			Side:         getSide(from.Inline),
			HunkHeader:   extractHunkInfo(from.Inline),
			SourceSHA:    from.Inline.SrcRev,
			MergeBaseSHA: from.Inline.DestRev,
			Outdated:     from.Inline.Outdated,
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

func getSide(inline *inline) string {
	if inline.From == nil {
		return "NEW"
	}
	return "OLD"
}

func extractSnippetInfo(diffHunk string) types.Hunk {
	lines := strings.Split(diffHunk, "\n")
	return types.Hunk{
		Header: lines[0],
		Lines:  lines[1:],
	}
}

func extractHunkInfo(inline *inline) string {
	re := regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) @@`)

	oldLine := 0
	newLine := 0
	oldSpan := 1
	newSpan := 1
	var err error

	// Find the first match
	matches := re.FindStringSubmatch(inline.ContextLines)
	if len(matches) != 5 {
		return ""
	}

	// Parse each number from the match results
	oldLine, err = strconv.Atoi(matches[1])
	if err != nil {
		return common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, inline.ContextLines)
	}

	oldSpan, err = strconv.Atoi(matches[2])
	if err != nil {
		return common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, inline.ContextLines)
	}

	newLine, err = strconv.Atoi(matches[3])
	if err != nil {
		return common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, inline.ContextLines)
	}

	newSpan, err = strconv.Atoi(matches[4])
	if err != nil {
		return common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, inline.ContextLines)
	}

	fmt.Printf("OLD SPAM NEW SPAN %d %d %d %d", oldLine, oldSpan, newLine, newSpan)
	return common.FormatHunkHeader(oldLine, oldSpan, newLine, newSpan, inline.ContextLines)
}
