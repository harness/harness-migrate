package github

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"

	"github.com/drone/go-scm/scm"
	gonanoid "github.com/matoous/go-nanoid"
)

type changeType byte

const (
	added               changeType = '+'
	removed             changeType = '-'
	unchanged           changeType = ' '
	maxIdentifierLength            = 100
)

const logMessage = "[%s] Skipped mapping %q branch rule for pattern %q of repo %q as we do not support it as of now. \n"

var regExpHunkHeader = regexp.MustCompile(`^@@ -([0-9]+)(,([0-9]+))? \+([0-9]+)(,([0-9]+))? @@( (.+))?$`)

func convertPRCommentsList(from []*codeComment, repo string, pr int) []*types.PRComment {
	to := []*types.PRComment{}
	for _, v := range from {
		to = append(to, convertPRComment(v, repo, pr))
	}
	return to
}

func convertPRComment(from *codeComment, repo string, pr int) *types.PRComment {
	var parentID int
	var metadata *types.CodeComment
	// If the comment is a reply, we don't need the metadata
	// If the comment is on a file, diff_hunk will not be present
	if from.InReplyToID != 0 {
		parentID = from.InReplyToID
	} else if from.OriginalLine != nil && from.SubjectType != "file" {
		hunkHeader, err := extractHunkInfo(from)
		if err != nil {
			log.Default().Printf("Importing code comment %d on PR %d of repo %s as a PR comment: %v", from.ID, pr, repo, err)
		} else {
			metadata = &types.CodeComment{
				Path:         from.Path,
				CodeSnippet:  extractSnippetInfo(from.DiffHunk),
				Side:         getCommentSide(from.Side),
				HunkHeader:   hunkHeader,
				SourceSHA:    from.OriginalCommitID,
				MergeBaseSHA: from.CommitID,
				Outdated:     from.Line == nil,
			}
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

// TODO: Revisit this method to try getting the correct hunk header
// in some cases like rebase where the otherLine and otherSpan are
// not correct because of rebase offset.
func extractHunkInfo(comment *codeComment) (string, error) {
	var (
		oldLine   int
		oldSpan   int
		newLine   int
		newSpan   int
		otherLine int
		otherSpan int
		err       error
	)

	multiline := comment.OriginalStartLine != nil
	if comment.StartLine == nil {
		comment.StartLine = comment.OriginalStartLine
	}
	if comment.Line == nil {
		comment.Line = comment.OriginalLine
	}

	if multiline {
		span := *comment.Line - *comment.StartLine + 1
		otherLine, otherSpan, err = getOtherSideLineAndSpan(comment.DiffHunk, comment.Side == "RIGHT", *comment.OriginalStartLine, span)
		if err != nil {
			return "", fmt.Errorf("diff hunk information wrong: %v", err)
		}

		if comment.Side == "RIGHT" {
			oldLine = otherLine // missing rebase offset
			oldSpan = otherSpan
			newLine = *comment.StartLine
			newSpan = int(span)
		} else {
			oldLine = *comment.StartLine
			oldSpan = span
			newLine = otherLine // missing rebase offset
			newSpan = otherSpan
		}
	} else {
		otherLine, otherSpan, err = getOtherSideLineAndSpan(comment.DiffHunk, comment.Side == "RIGHT", *comment.OriginalLine, 1)
		if err != nil {
			return "", fmt.Errorf("diff hunk information wrong: %v", err)
		}

		if comment.Side == "RIGHT" {
			oldLine = otherLine
			oldSpan = otherSpan
			newLine = *comment.Line // missing rebase offset
			newSpan = 1
		} else {
			oldLine = *comment.Line
			oldSpan = 1
			newLine = otherLine // missing rebase offset
			newSpan = otherSpan
		}
	}
	return common.FormatHunkHeader(int(oldLine), int(oldSpan), int(newLine), int(newSpan), ""), nil
}

func getOtherSideLineAndSpan(rawHunk string, newSide bool, line, span int) (int, int, error) {
	var otherLine, otherSpan int
	var haveOtherLine bool

	err := processHunk(rawHunk, func(oldLine, newLine int, change changeType) {
		inSelected :=
			newSide && newLine >= line && newLine < line+span ||
				!newSide && oldLine >= line && oldLine < line+span
		otherSide := change == unchanged || change == removed && newSide || change == added && !newSide

		if inSelected {
			// set the default value for the other side's line number (will be used if the other side's span is 0)
			if otherLine == 0 {
				if newSide {
					otherLine = oldLine
				} else {
					otherLine = newLine
				}
			}
			// if the current line belongs to the other side
			if otherSide {
				otherSpan++
				if !haveOtherLine {
					haveOtherLine = true
					// set the value for the other side's line number
					if newSide {
						otherLine = oldLine
					} else {
						otherLine = newLine
					}
				}
			}
		}
	})
	if err != nil {
		return 0, 0, err
	}

	return otherLine, otherSpan, nil
}

func processHunk(rawHunk string, fnLine func(oldLine, newLine int, change changeType)) error {
	scan := bufio.NewScanner(strings.NewReader(rawHunk))
	if !scan.Scan() {
		return errors.New("hunk header missing")
	}
	hunkHeader := scan.Text()
	hunk, ok := parseDiffHunkHeader(hunkHeader)
	if !ok {
		return fmt.Errorf("invalid diff hunk header: %s", hunkHeader)
	}

	oldLine, newLine := hunk.OldLine, hunk.NewLine

	for scan.Scan() {
		text := scan.Text()
		if text == "" {
			return errors.New("empty line in hunk body")
		}

		change := changeType(text[0])
		if change != added && change != removed && change != unchanged {
			return fmt.Errorf("invalid line in hunk body: %s", text)
		}

		fnLine(oldLine, newLine, change)

		switch change {
		case added:
			newLine++
		case removed:
			oldLine++
		case unchanged:
			oldLine++
			newLine++
		}
	}

	return nil
}

func parseDiffHunkHeader(line string) (HunkHeader, bool) {
	groups := regExpHunkHeader.FindStringSubmatch(line)
	if groups == nil {
		return HunkHeader{}, false
	}

	oldLine, _ := strconv.Atoi(groups[1])
	oldSpan := 1
	if groups[3] != "" {
		oldSpan, _ = strconv.Atoi(groups[3])
	}

	newLine, _ := strconv.Atoi(groups[4])
	newSpan := 1
	if groups[6] != "" {
		newSpan, _ = strconv.Atoi(groups[6])
	}

	return HunkHeader{
		OldLine: oldLine,
		OldSpan: oldSpan,
		NewLine: newLine,
		NewSpan: newSpan,
		Text:    groups[8],
	}, true
}

func convertBranchRulesList(from *branchProtectionRulesResponse, repo string, l gitexporter.Logger) []*types.BranchRule {
	to := []*types.BranchRule{}
	for _, edge := range from.Data.Repository.BranchProtectionRules.Edges {
		to = append(to, convertBranchRule(edge.Node, repo, l)...)
	}
	return to
}

func convertBranchRulesetsList(from []*ruleSet) []*types.BranchRule {
	to := []*types.BranchRule{}
	for _, rule := range from {
		if rule.Target == "branch" {
			to = append(to, &types.BranchRule{ID: rule.ID})
		}
	}
	return to
}

func convertBranchRule(from branchProtectionRule, repo string, l gitexporter.Logger) []*types.BranchRule {
	var logs []string
	var warningMsg string
	rules := []*types.BranchRule{}
	// randomize is set as true as rulesets might have same pattern
	ruleName, err := patternNameToIdentifier(from.Pattern, true)
	if err != nil {
		log.Default().Printf("failed to name pattern %q to identifier: %v", from.Pattern, err)
	}
	rule := &types.BranchRule{
		ID:    from.DatabaseID,
		Name:  ruleName,
		State: enum.RuleStateActive,
		Pattern: types.Pattern{
			IncludedPatterns: []string{from.Pattern},
		},
	}

	if !from.AllowsDeletions {
		rule.DeleteForbidden = true
	}
	if from.AllowsForcePushes {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "force push allowances", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.BlocksCreations {
		rule.Lifecycle.CreateForbidden = true
	}
	if from.BypassPullRequestAllowances.TotalCount > 0 {
		actorNotUser := false
		for _, actor := range from.BypassPullRequestAllowances.Edges {
			if actor.Node.Actor.Email != "" {
				rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, actor.Node.Actor.Email)
			} else {
				actorNotUser = true
			}
		}
		if actorNotUser {
			warningMsg = fmt.Sprintf("[%s] Skipped adding teams and apps to bypass list for pattern %q of repo %q as we do not support it as of now. \n",
				enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
	}
	if from.DismissesStaleReviews {
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true
	}
	if !from.IsAdminEnforced {
		rule.Definition.Bypass.RepoOwners = true
	}
	if from.LockBranch {
		rule.Lifecycle.UpdateForbidden = true
		rule.Lifecycle.DeleteForbidden = true
	}
	if from.RequireLastPushApproval {
		rule.UpdateForbidden = true
		rule.RequireLatestCommit = true
	}
	if from.RequiresApprovingReviews {
		rule.UpdateForbidden = true
		rule.RequireMinimumCount = from.RequiredApprovingReviewCount
	}
	if from.RequiresCodeOwnerReviews {
		rule.UpdateForbidden = true
		rule.RequireCodeOwners = true
	}
	if from.RequiresCommitSignatures {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required commit signatures", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresConversationResolution {
		rule.UpdateForbidden = true
		rule.RequireResolveAll = true
	}
	if from.RequiresDeployments {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required deployments", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresLinearHistory {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "required linear history", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RequiresStatusChecks {
		warningMsg = fmt.Sprintf("[%s] Skipped adding status checks. Create the status checks' pipelines as in branch rule %q for repo %q and reconfigure the branch rule. \n",
			enum.LogLevelWarning, from.Pattern, repo)
		logs = append(logs, warningMsg)
	}
	if from.RestrictsPushes {
		r := &types.BranchRule{
			ID:    from.DatabaseID,
			Name:  ruleName + "_restricts_pushes",
			State: enum.RuleStateActive,
			Pattern: types.Pattern{
				IncludedPatterns: []string{from.Pattern},
			},
		}
		r.UpdateForbidden = true
		if !from.IsAdminEnforced {
			r.Definition.Bypass.RepoOwners = true
		}
		actorNotUser := false
		for _, actor := range from.PushAllowances.Edges {
			if actor.Node.Actor.Email != "" {
				rule.Bypass.UserEmails = append(rule.Bypass.UserEmails, actor.Node.Actor.Email)
			} else {
				actorNotUser = true
			}
		}
		if actorNotUser {
			warningMsg = fmt.Sprintf("[%s] Skipped adding teams and apps to bypass list for branch rule with pattern %q of repo %q as we do not support it as of now. \n",
				enum.LogLevelWarning, from.Pattern, repo)
			logs = append(logs, warningMsg)
		}
		rules = append(rules, r)
	}
	if from.RestrictsReviewDismissals {
		warningMsg = fmt.Sprintf(logMessage, enum.LogLevelWarning, "restricts review dismissals", from.Pattern, repo)
		logs = append(logs, warningMsg)
	}

	rules = append(rules, rule)
	if err := l.Log(strings.Join(logs, "")); err != nil {
		log.Default().Printf("failed to log the not supported branch rules for repo %q: %v", repo, err)
		return rules
	}
	return rules
}

func convertBranchRuleset(from *detailedRuleSet, repo string, l gitexporter.Logger) *types.BranchRule {
	includedPatterns, includeDefault := mapPatterns(from.Conditions.RefName.Include)
	excludedPatterns, _ := mapPatterns(from.Conditions.RefName.Exclude)
	return &types.BranchRule{
		ID:         from.ID,
		Name:       from.Name,
		State:      mapRuleEnforcement(from.Enforcement),
		Definition: mapRuleDefinition(from, repo, l),
		Pattern: types.Pattern{
			IncludeDefault:   includeDefault,
			IncludedPatterns: includedPatterns,
			ExcludedPatterns: excludedPatterns,
		},
		Created: from.CreatedAt,
		Updated: from.UpdatedAt,
	}
}

func mapRuleEnforcement(enforcement string) enum.RuleState {
	switch enforcement {
	case "active":
		return enum.RuleStateActive
	case "evaluate":
		return enum.RuleStateMonitor
	case "disabled":
		return enum.RuleStateDisabled
	default:
		return enum.RuleStateDisabled
	}
}

func mapPatterns(branches []string) ([]string, bool) {
	includeDefault := false
	res := []string{}
	for _, b := range branches {
		switch b {
		case "~DEFAULT_BRANCH":
			includeDefault = true
		// This will include all the branches in the repo
		case "~ALL":
			res = append(res, "**")
		default:
			res = append(res, strings.TrimPrefix(b, "refs/heads/"))
		}
	}
	return res, includeDefault
}

func mapRuleDefinition(from *detailedRuleSet, repo string, l gitexporter.Logger) types.Definition {
	definition := types.Definition{}
	var logs []string
	var warningMsg string

	for _, r := range from.Rules {
		switch r.Type {
		case "creation":
			definition.CreateForbidden = true
		case "update":
			definition.UpdateForbidden = true
		case "deletion":
			definition.DeleteForbidden = true
		case "required_linear_history":
			definition.UpdateForbidden = true
		case "pull_request":
			definition.UpdateForbidden = true
			parameters := extractPullRequestParameters(r.Parameters)
			if parameters.RequiredApprovingReviewCount > 0 {
				definition.RequireMinimumCount = parameters.RequiredApprovingReviewCount
			}
			definition.RequireCodeOwners = parameters.RequireCodeOwnerReview
			definition.RequireLatestCommit = parameters.RequireLastPushApproval
			definition.RequireResolveAll = parameters.RequiredReviewThreadResolution
			definition.RequireLatestCommit = parameters.DismissStaleReviewsOnPush
		case "required_status_checks":
			warningMsg = fmt.Sprintf("[%s] Skipped adding status checks. Create the status checks' pipelines as in branch rule %q for repo %q and reconfigure the branch rule. \n",
				enum.LogLevelWarning, from.Name, repo)
			logs = append(logs, warningMsg)
		case "non_fast_forward":
			definition.UpdateForbidden = true
		default:
			warningMsg = fmt.Sprintf("[%s] Skipped mapping rule type %q for branch rule %q of repo %q as we do not support it as of now. \n",
				enum.LogLevelWarning, r.Type, from.Name, repo)
			logs = append(logs, warningMsg)
		}
	}
	if len(from.BypassActors) != 0 {
		msg := fmt.Sprintf("[%s] Couldn't map bypass list for branch rule %q of repo %q. Need to reconfigure the bypass list for the branch rule. \n",
			enum.LogLevelWarning, from.Name, repo)
		logs = append(logs, msg)
	}

	if err := l.Log(strings.Join(logs, "")); err != nil {
		log.Default().Printf("failed to log the not supported branch rules for repo %q: %v", repo, err)
		return definition
	}
	return definition
}

func extractPullRequestParameters(params map[string]interface{}) pullRequestParameters {
	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Default().Printf("failed to marshal branch rule pull request parameters: %v", err)
		return pullRequestParameters{}
	}

	var parameters pullRequestParameters
	if err := json.Unmarshal(jsonData, &parameters); err != nil {
		log.Default().Printf("failed to unmarshal branch rul pull request parameters: %v", err)
		return pullRequestParameters{}
	}

	return parameters
}

func encodeListOptions(opts types.ListOptions) string {
	params := url.Values{}
	limit := common.DefaultLimit
	if opts.Size != 0 {
		limit = opts.Size
	}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	params.Set("per_page", strconv.Itoa(limit))
	return params.Encode()
}

func patternNameToIdentifier(displayName string, randomize bool) (string, error) {
	const placeholder = '_'
	const specialChars = ".-_"
	// remove / replace any illegal characters
	// Identifier Regex: ^[a-zA-Z0-9-_.]*$
	identifier := strings.Map(func(r rune) rune {
		switch {
		// drop any control characters or empty characters
		case r < 32 || r == 127:
			return -1

		// keep all allowed character
		case ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') ||
			strings.ContainsRune(specialChars, r):
			return r

		// everything else is replaced with the placeholder
		default:
			return placeholder
		}
	}, displayName)

	// remove any leading/trailing special characters
	identifier = strings.Trim(identifier, specialChars)

	// ensure string doesn't start with numbers (leading '_' is valid)
	if len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9' {
		identifier = string(placeholder) + identifier
	}

	// remove consecutive special characters
	identifier = sanitizeConsecutiveChars(identifier, specialChars)

	// ensure length restrictions
	if len(identifier) > maxIdentifierLength {
		identifier = identifier[0:maxIdentifierLength]
	}

	// backfill randomized identifier if sanitization ends up with empty identifier
	if len(identifier) == 0 {
		identifier = "rule"
		randomize = true
	}

	if randomize {
		return randomizeIdentifier(identifier)
	}

	return identifier, nil
}

func sanitizeConsecutiveChars(in string, charSet string) string {
	if len(in) == 0 {
		return ""
	}

	inSet := func(b byte) bool {
		return strings.ContainsRune(charSet, rune(b))
	}

	out := strings.Builder{}
	out.WriteByte(in[0])
	for i := 1; i < len(in); i++ {
		if inSet(in[i]) && inSet(in[i-1]) {
			continue
		}
		out.WriteByte(in[i])
	}

	return out.String()
}

func randomizeIdentifier(identifier string) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 4
	const maxLength = maxIdentifierLength - length - 1 // max length of identifier to fit random suffix

	if len(identifier) > maxLength {
		identifier = identifier[0:maxLength]
	}
	suffix, err := gonanoid.Generate(alphabet, length)
	if err != nil {
		return "", fmt.Errorf("failed to generate gonanoid: %w", err)
	}

	return identifier + "_" + suffix, nil
}
