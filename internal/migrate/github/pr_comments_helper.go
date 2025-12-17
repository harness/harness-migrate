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

package github

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

type changeType byte

const (
	added               changeType = '+'
	removed             changeType = '-'
	unchanged           changeType = ' '
	maxIdentifierLength            = 100
)

const logMessage = "[%s] Skipped mapping %q branch rule for pattern %q of repo %q as we do not support it as of now."

func (e *Export) convertPRCommentsList(from []*codeComment, repo string, pr int) []*types.PRComment {
	var to []*types.PRComment
	for _, v := range from {
		to = append(to, e.convertPRComment(v, repo, pr))
	}
	return to
}

func (e *Export) convertPRComment(from *codeComment, repo string, pr int) *types.PRComment {
	var parentID int
	var metadata *types.CodeComment
	// If the comment is a reply, we don't need the metadata
	// If the comment is on a file, diff_hunk will not be present
	if from.InReplyToID != 0 {
		parentID = from.InReplyToID
	} else if from.OriginalLine != nil && from.SubjectType != "file" {
		hunkHeader, err := extractHunkInfo(from)
		if err != nil {
			e.fileLogger.Log(fmt.Sprintf("Importing code comment %d on PR %d of repo %s as a PR comment: %v", from.ID, pr, repo, err))
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

	hunkHeader, err := common.FormatHunkHeader(int(oldLine), int(oldSpan), int(newLine), int(newSpan), "")
	if err != nil {
		return "", fmt.Errorf("invalid hunk header values: %w", err)
	}

	return hunkHeader, nil
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
	hunk, ok := migrate.ParseDiffHunkHeader(hunkHeader)
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
