package github

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type changeType byte

const (
	added     changeType = '+'
	removed   changeType = '-'
	unchanged changeType = ' '
)

var regExpHunkHeader = regexp.MustCompile(`^@@ -([0-9]+)(,([0-9]+))? \+([0-9]+)(,([0-9]+))? @@( (.+))?$`)

func convertPRCommentsList(from []*codeComment) []*types.PRComment {
	to := []*types.PRComment{}
	for _, v := range from {
		to = append(to, convertPRComment(v))
	}
	return to
}

func convertPRComment(from *codeComment) *types.PRComment {
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
			SourceSHA:    from.OriginalCommitID,
			MergeBaseSHA: from.CommitID,
			Outdated:     from.Line == nil,
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
func extractHunkInfo(comment *codeComment) string {
	var (
		oldLine   int
		oldSpan   int
		newLine   int
		newSpan   int
		otherLine int
		otherSpan int
		err       error
	)

	multiline := comment.StartLine != nil || comment.OriginalStartLine != nil
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
			log.Default().Printf("hunk information wrong: %v", err)
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
			log.Default().Printf("hunk information missing: %v", err)
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
	return common.FormatHunkHeader(int(oldLine), int(oldSpan), int(newLine), int(newSpan), "")
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
