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

package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"
)

func (e *Export) convertPRCommentsList(from []*discussion, pr int) []*types.PRComment {
	var to []*types.PRComment
	for _, v := range from {
		to = append(to, e.convertPRComments(v, pr)...)
	}
	return to
}

func (e *Export) convertPRComments(from *discussion, pr int) []*types.PRComment {
	var parentID int
	var comments []*types.PRComment
	for _, note := range from.Notes {
		// ignore system generate notes as they should exported as diff PR activity rather comments
		if note.System {
			return nil
		}

		comment := &types.PRComment{
			Comment: scm.Comment{
				ID:   note.ID,
				Body: note.Body,
				Author: scm.User{
					Login:  note.Author.Username,
					Name:   note.Author.Name,
					Avatar: note.Author.AvatarURL,
				},
				Created: note.CreatedAt,
				Updated: note.UpdatedAt,
			},
			ParentID: parentID,
		}

		if note.Type == "DiffNote" && note.Position != nil {
			hunkHeader := extractHunkInfo(note)
			if hunkHeader == "" {
				e.fileLogger.Log(fmt.Sprintf("Importing comment %d on PR %d of repo %s as a PR comment due to invalid hunk header values", note.ID, pr, e.project))
				continue // add comment as regular comment
			}
			comment.CodeComment = &types.CodeComment{
				Path: note.Position.NewPath,
				// Gitlab doesnt return code diffs on discussion API
				CodeSnippet: types.Hunk{
					Header: "",
					Lines:  []string{},
				},
				HunkHeader:   hunkHeader,
				Side:         getSide(*note.Position),
				SourceSHA:    note.Position.HeadSHA,
				MergeBaseSHA: note.Position.BaseSHA,
			}
		}

		parentID = note.ID
		comments = append(comments, comment)
	}

	return comments
}

func extractHunkInfo(comment codeComment) string {
	leftLineStart, rightLineStart := getOldNewLines(comment.Position.LineRange.Start.LineCode)
	leftLineEnd, rightLineEnd := getOldNewLines(comment.Position.LineRange.End.LineCode)

	sourceSpan := leftLineEnd - leftLineStart + 1
	desSpan := rightLineEnd - rightLineStart + 1

	hunkHeader, err := common.FormatHunkHeader(leftLineStart, sourceSpan, rightLineStart, desSpan, "")
	if err != nil {
		// Return empty string if validation fails - caller will handle by converting to regular comment
		return ""
	}
	return hunkHeader
}

func getOldNewLines(lineCode string) (int, int) {
	// e.g: 44394fc87fbb7499b38bb65b9e85eaefcef1396f_6_3
	parts := strings.Split(lineCode, "_")
	if len(parts) < 3 {
		return 0, 0
	}
	oldLine, _ := strconv.Atoi(parts[1])
	newLine, _ := strconv.Atoi(parts[2])

	return oldLine, newLine
}

// getSide gets the comments side in a heuristic manner, as Gitlab comments appear all in one side.
func getSide(pos position) string {
	old := "old"
	if pos.LineRange.Start.Type == old &&
		pos.LineRange.End.Type == old {
		return "OLD"
	}
	return "NEW"
}
