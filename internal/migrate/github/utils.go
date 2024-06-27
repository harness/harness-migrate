package github

import (
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func convertPRCommentsList(from []*prComment) []*types.PRComment {
	to := []*types.PRComment{}
	for _, v := range from {
		to = append(to, convertPRComment(v))
	}
	return to
}

func convertPRComment(from *prComment) *types.PRComment {
	var parentID int
	var metadata *types.CodeComment
	if from.InReplyToID != 0 {
		parentID = from.InReplyToID
	} else {
		metadata = &types.CodeComment{
			Path:         from.Path,
			CodeSnippet:  extractSnippetInfo(from.DiffHunk),
			Side:         getCommentSide(from.Side),
			HunkHeader:   extractHunkInfo(from),
			SourceSha:    from.OriginalCommitID,
			MergeBaseSha: from.CommitID,
		}
	}
	return &types.PRComment{Comment: scm.Comment{
		ID:      from.ID,
		Body:    from.Body,
		Created: from.CreatedAt,
		Updated: from.UpdatedAt,
		Author: scm.User{
			Login:  from.User.Login,
			Avatar: from.User.AvatarURL,
		}},
		ParentID:    parentID,
		CodeComment: metadata,
	}
}

func getCommentSide(side string) string {
	if side == "LEFT" {
		return "OLD"
	}
	return "NEW"
}

func extractSnippetInfo(diffHunk string) types.Hunk {
	lines := strings.Split(diffHunk, "\n")
	return types.Hunk{
		Header: lines[0],
		Lines:  lines[1:],
	}
}

func extractHunkInfo(comment *prComment) string {
	sourceSpan, destinationSpan := 1, 1
	return fmt.Sprintf("@@ -%d,%d +%d,%d @@", comment.OriginalLine, sourceSpan, comment.Line, destinationSpan)

}
