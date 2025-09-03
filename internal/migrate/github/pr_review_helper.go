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
	"strconv"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/harness/harness-migrate/internal/types"
	"github.com/harness/harness-migrate/internal/types/enum"
)

func convertPRReviewsList(from []*review) []*types.PRReview {
	var to []*types.PRReview
	for _, v := range from {
		to = append(to, convertPRReview(v))
	}
	return to
}

func convertPRReview(from *review) *types.PRReview {
	var decision enum.ReviewDecision
	switch strings.ToUpper(from.State) {
	case "APPROVED":
		decision = enum.ReviewDecisionApproved
	case "CHANGES_REQUESTED", "REQUEST_CHANGES":
		decision = enum.ReviewDecisionChangeReq
	case "COMMENTED":
		decision = enum.ReviewDecisionReviewed
	case "PENDING":
		decision = enum.ReviewDecisionPending
	default:
		decision = enum.ReviewDecisionReviewed // default to reviewed for safety
	}

	return &types.PRReview{
		Review: scm.Review{
			ID:   from.ID,
			Body: from.Body,
			Author: scm.User{
				ID:    strconv.Itoa(from.User.ID),
				Login: from.User.Login,
			},
			Created: from.SubmittedAt,
			Updated: from.SubmittedAt,
			Sha:     from.CommitID,
		},
		State: decision,
	}
}

func convertRequestedReviewersList(from requestedReviewersResponse) []*types.PRReviewer {
	var to []*types.PRReviewer

	for _, user := range from.Users {
		to = append(to, &types.PRReviewer{
			User: scm.User{
				ID:     strconv.Itoa(user.ID),
				Login:  user.Login,
				Avatar: user.AvatarURL,
			},
		})
	}

	return to
}
