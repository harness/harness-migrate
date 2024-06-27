package stash

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

const (
	defaultLimit             = 25
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
	to := []prCommentActivity{}
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
	to := []*types.PRComment{}
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
			SourceSha:    anchor.ToHash,
			MergeBaseSha: anchor.FromHash,
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
	header := formatHunkHeader(hunk.SourceLine, hunk.SourceSpan, hunk.DestinationLine, hunk.DestinationSpan, "")
	lines := []string{}
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
			return formatHunkHeader(line.Source, sourceSpan, line.Destination, destinationSpan, "")
		}
	}
	return ""
}

func formatHunkHeader(source, sourceSpan, destination, destinationSpan int, sectionHeading string) string {
	sb := strings.Builder{}
	sb.Grow(20 + len(sectionHeading))

	sb.WriteString("@@ -")
	sb.WriteString(strconv.Itoa(source))
	if sourceSpan != 1 {
		sb.WriteByte(',')
		sb.WriteString(strconv.Itoa(sourceSpan))
	}
	sb.WriteString(" +")
	sb.WriteString(strconv.Itoa(destination))
	if destinationSpan != 1 {
		sb.WriteByte(',')
		sb.WriteString(strconv.Itoa(destinationSpan))
	}
	sb.WriteString(" @@")

	if sectionHeading != "" {
		sb.WriteByte(' ')
		sb.WriteString(sectionHeading)
	}

	return sb.String()
}

func convertBranchRulesList(from []*branchPermission, m map[string]modelValue) []*types.BranchRule {
	rules := []*types.BranchRule{}
	for _, p := range from {
		rules = append(rules, convertBranchRule(p, m))
	}
	return rules
}

func convertBranchRule(from *branchPermission, m map[string]modelValue) *types.BranchRule {
	includeDefault := false
	branches := []string{}
	includedPatterns := []string{}
	switch from.Matcher.Type.ID {
	case matcherTypeBranch:
		// displayID will give just branch name main and ID will give refs/heads/main
		branches = append(branches, from.Matcher.DisplayID)
	case matcherTypePattern:
		includedPatterns = append(includedPatterns, convertIntoGlobstar(from.Matcher.ID))
	case matcherTypeModelBranch:
		v := m[from.Matcher.ID]
		if v.UseDefault {
			includeDefault = true
		} else {
			branches = append(branches, strings.TrimPrefix(v.RefID, "refs/heads/"))
		}
	case matcherTypeModelCategory:
		includedPatterns = append(includedPatterns, convertIntoGlobstar(m[from.Matcher.ID].Prefix))
	}
	return &types.BranchRule{
		ID:               from.ID,
		Name:             from.Matcher.DisplayID,
		Type:             from.Type,
		IncludeDefault:   includeDefault,
		Branches:         branches,
		IncludedPatterns: includedPatterns,
		BypassUsers:      from.Users,
	}
}

func convertBranchModelsMap(from branchModels) map[string]modelValue {
	m := map[string]modelValue{}
	m[branchDevelopment] = modelValue{modelBranch: from.Development}
	m[branchProduction] = modelValue{modelBranch: from.Production}
	for _, c := range from.Categories {
		m[c.ID] = modelValue{Prefix: c.Prefix}
	}
	return m
}

func convertIntoGlobstar(s string) string {
	if strings.HasSuffix(s, "/") {
		return s + "**"
	}
	return s
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
