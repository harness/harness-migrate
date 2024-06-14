package stash

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

const (
	defaultLimit = 25
	COMMENTED    = "COMMENTED"
)

func filterOutCommentActivities(from []interface{}, tracer tracer.Tracer) []prCommentActivity {
	to := []prCommentActivity{}
	for i, activity := range from {
		if activityMap, ok := activity.(map[string]any); ok {
			if action, ok := activityMap["action"].(string); ok && action == COMMENTED {
				// Marshal the map to JSON, then unmarshal to prCommentActivity
				data, err := json.Marshal(activityMap)
				if err != nil {
					tracer.LogError("Error parsing JSON from activity %d: %v", i, err)
					continue
				}
				var prComment prCommentActivity
				if err := json.Unmarshal(data, &prComment); err != nil {
					tracer.LogError("Error converting comment activity %d from JSON: %v", i, err)
					continue
				}
				to = append(to, prComment)
			}
		}
	}
	return to
}

func convertPullRequestCommentsList(from []interface{}, tracer tracer.Tracer) []*types.PRComment {
	commentActivities := filterOutCommentActivities(from, tracer)
	to := []*types.PRComment{}
	for _, c := range commentActivities {
		comment := c.Comment
		to = append(to, convertPullRequestComment(comment, 0, &c.CommentAnchor))
		childComments := comment.Comments
		// child comments are nested
		for len(childComments) > 0 {
			if len(childComments) == 0 {
				break
			}
			childComment := childComments[0]
			to = append(to, convertPullRequestComment(childComment, comment.ID, nil))
			childComments = childComment.Comments
		}
	}
	return to
}

func convertPullRequestComment(from pullRequestComment, parentID int, anchor *commentAnchor) *types.PRComment {
	var metadata types.CommentMetadata
	if anchor != nil && (*anchor).Path != "" {
		commentAnchor := *anchor
		metadata = types.CommentMetadata{
			Path:         commentAnchor.Path,
			Line:         commentAnchor.Line,
			LineSpan:     1,
			SourceSha:    commentAnchor.FromHash,
			MergeBaseSha: commentAnchor.ToHash,
		}
	}
	if parentID != 0 {
		metadata = types.CommentMetadata{
			ParentID: parentID,
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
		CommentMetadata: metadata,
	}
}

func (e *Error) Error() string {
	if len(e.Errors) == 0 {
		if len(e.Message) > 0 {
			return fmt.Sprintf("bitbucket: status: %d message: %s", e.Status, e.Message)
		}
		return "bitbucket: undefined error"
	}
	return e.Errors[0].Message
}

func encodeListOptions(opts types.ListOptions) string {
	params := url.Values{}
	limit := defaultLimit
	if opts.Size != 0 {
		limit = opts.Size
	}
	params.Set("limit", strconv.Itoa(limit))

	if opts.Page > 0 {
		params.Set("start", strconv.Itoa(
			(opts.Page-1)*limit),
		)
	}
	return params.Encode()
}
