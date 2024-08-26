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

package stash

import (
	"encoding/json"
	"log"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

const (
	segmentAdded             = "ADDED"
	segmentRemoved           = "REMOVED"
	segmentContext           = "CONTEXT"
	branchDevelopment        = "development"
	branchProduction         = "production"
	matcherTypeBranch        = "BRANCH"
	matcherTypePattern       = "PATTERN"
	matcherTypeModelBranch   = "MODEL_BRANCH"
	matcherTypeModelCategory = "MODEL_CATEGORY"
)

func filterOutCommentActivities(from []any) []prCommentActivity {
	var to []prCommentActivity
	for i, activity := range from {
		if activityMap, ok := activity.(map[string]any); ok {
			if action, ok := activityMap["action"].(string); ok && action == "COMMENTED" {
				// Marshal the map to JSON, then unmarshal to prCommentActivity
				data, err := json.Marshal(activityMap)
				if err != nil {
					log.Default().Printf("Error parsing JSON from activity %d: %v", i, err)
					continue
				}
				var prComment prCommentActivity
				if err := json.Unmarshal(data, &prComment); err != nil {
					log.Default().Printf("Error converting comment activity %d from JSON: %v", i, err)
					continue
				}
				to = append(to, prComment)
			}
		}
	}
	return to
}

func convertPullRequestCommentsList(from []any) []*types.PRComment {
	commentActivities := filterOutCommentActivities(from)
	var to []*types.PRComment
	for _, c := range commentActivities {
		comment := c.Comment
		to = append(to, convertPullRequestComment(comment, 0, c.CommentAnchor, c.Diff))
		childComments := comment.Comments
		// child comments are nested
		for len(childComments) > 0 {
			if len(childComments) == 0 {
				break
			}
			childComment := childComments[0]
			to = append(to, convertPullRequestComment(childComment, comment.ID, commentAnchor{}, codeDiff{}))
			childComments = childComment.Comments
		}
	}
	return to
}

func convertPullRequestComment(from pullRequestComment, parentID int, anchor commentAnchor, diff codeDiff) *types.PRComment {
	var codeComment *types.CodeComment
	if anchor.Path != "" {
		var snippet types.Hunk
		var hunkHeader string
		if anchor.Line != 0 {
			snippet = extractSnippetInfo(diff)
			hunkHeader = extractHunkInfo(anchor, diff)
		}
		codeComment = &types.CodeComment{
			Path:         anchor.Path,
			CodeSnippet:  snippet,
			Side:         getCommentSide(anchor.FileType),
			HunkHeader:   hunkHeader,
			SourceSHA:    anchor.ToHash,
			MergeBaseSHA: anchor.FromHash,
			Outdated:     anchor.Orphaned,
		}
	}
	return &types.PRComment{Comment: scm.Comment{
		ID:      from.ID,
		Body:    from.Text,
		Created: time.Unix(from.CreatedDate/1000, 0),
		Updated: time.Unix(from.UpdatedDate/1000, 0),
		Author: scm.User{
			Login: from.Author.Slug,
			Name:  from.Author.DisplayName,
			Email: from.Author.EmailAddress,
		}},
		ParentID:    parentID,
		CodeComment: codeComment,
	}
}

func getCommentSide(fileType string) string {
	if fileType == "FROM" {
		return "OLD"
	}
	return "NEW"
}

func extractSnippetInfo(diff codeDiff) types.Hunk {
	hunk := diff.Hunks[0]
	header := common.FormatHunkHeader(hunk.SourceLine, hunk.SourceSpan, hunk.DestinationLine, hunk.DestinationSpan, "")
	var lines []string
	for _, segment := range hunk.Segments {
		l := ""
		switch segment.Type {
		case segmentAdded:
			l += "+"
		case segmentRemoved:
			l += "-"
		case segmentContext:
			l += " "
		}
		for _, line := range segment.Lines {
			lines = append(lines, l+line.Line)
		}
	}
	return types.Hunk{
		Header: header,
		Lines:  lines,
	}
}

func extractHunkInfo(anchor commentAnchor, diff codeDiff) string {
	hunk := diff.Hunks[0]
	for _, segment := range hunk.Segments {
		if anchor.LineType != segment.Type {
			continue
		}
		for _, line := range segment.Lines {
			if line.CommentIDs == nil {
				continue
			}
			sourceSpan, destinationSpan := 1, 1
			if segment.Type == segmentAdded {
				sourceSpan = 0
			}
			if segment.Type == segmentRemoved {
				destinationSpan = 0
			}
			return common.FormatHunkHeader(line.Source, sourceSpan, line.Destination, destinationSpan, "")
		}
	}
	return ""
}
