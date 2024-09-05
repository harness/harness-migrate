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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

const graphqlUrl = "https://api.github.com/graphql"

type (
	Export struct {
		github     *scm.Client
		org        string
		repository string

		checkpointManager *checkpoint.CheckpointManager

		tracer     tracer.Tracer
		fileLogger *gitexporter.FileLogger
		report     map[string]*report.Report

		userMap map[string]types.User
	}

	graphQLRequest struct {
		Query string `json:"query"`
	}
)

func (e *Export) ListPRComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments?%s", repoSlug, prNumber, encodeListOptions(opts))
	var out []*codeComment
	res, err := e.do(ctx, "GET", path, nil, &out)
	return convertPRCommentsList(out, repoSlug, prNumber), res, err
}

func (e *Export) GetUserByUserName(
	ctx context.Context,
	userName string,
) (*types.User, *scm.Response, error) {
	path := fmt.Sprintf("users/%s", userName)
	out := &types.User{}
	res, err := e.do(ctx, "GET", path, nil, &out)
	return out, res, err
}

func (e *Export) ListBranchRulesInternal(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	owner, name := scm.Split(repoSlug)
	queryTemplate := `
		{
			repository(owner: "%s", name: "%s") {
				branchProtectionRules(first: %d%s) {
					edges {
						node {
							allowsDeletions
							allowsForcePushes
							blocksCreations
							bypassForcePushAllowances {
								totalCount
							}
							bypassPullRequestAllowances(first: 100) {
								totalCount
								edges {
									node {
										actor {
											... on User {
												login
												email
											}
										}
									}
								}
							}
							dismissesStaleReviews
							id
							isAdminEnforced
							lockAllowsFetchAndMerge
							lockBranch
							pattern
							pushAllowances(first: 100) {
								totalCount
								edges {
									node {
										actor {
											... on User {
												login
												email
											}
										}
									}
								}
							}
							requireLastPushApproval
							requiredApprovingReviewCount
							requiredDeploymentEnvironments
							requiresApprovingReviews
							requiresCodeOwnerReviews
							requiresCommitSignatures
							requiresConversationResolution
							requiresDeployments
							requiresLinearHistory
							requiresStatusChecks
							requiresStrictStatusChecks
							restrictsPushes
							restrictsReviewDismissals
							reviewDismissalAllowances {
								totalCount
							}
						}
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		}
	`
	var pagination string
	if opts.Size == 0 {
		opts.Size = common.DefaultLimit
	}
	if opts.URL != "" {
		pagination = fmt.Sprintf(`, after: "%s"`, opts.URL)
	}
	query := fmt.Sprintf(queryTemplate, owner, name, opts.Size, pagination)
	body := graphQLRequest{
		Query: query,
	}
	out := new(branchProtectionRulesResponse)
	res, err := e.do(ctx, "POST", graphqlUrl, body, &out)
	if out.Data.Repository.BranchProtectionRules.PageInfo.HasNextPage {
		res.Page.NextURL = out.Data.Repository.BranchProtectionRules.PageInfo.EndCursor
	}
	return e.convertBranchRulesList(out, repoSlug), res, err
}

func (e *Export) ListBranchRuleSets(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/rulesets?%s", repoSlug, encodeListOptions(opts))
	var out []*ruleSet
	res, err := e.do(ctx, "GET", path, nil, &out)
	return e.convertBranchRuleSetsList(out), res, err
}

func (e *Export) FindBranchRuleset(
	ctx context.Context,
	repoSlug string,
	ruleID int,
) (*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/rulesets/%d", repoSlug, ruleID)
	out := new(detailedRuleSet)
	res, err := e.do(ctx, "GET", path, nil, &out)
	return e.convertBranchRuleset(out, repoSlug), res, err
}

func (e *Export) do(ctx context.Context, method, path string, in, out interface{}) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
	}
	// if we are posting or putting data, we need to
	// write it to the body of the request.
	if in != nil {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(in)
		req.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}
		req.Body = buf
	}

	// execute the http request
	res, err := e.github.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse the github request id.
	res.ID = res.Header.Get("X-GitHub-Request-Id")

	// parse the github rate limit details.
	res.Rate.Limit, _ = strconv.Atoi(
		res.Header.Get("X-RateLimit-Limit"),
	)
	res.Rate.Remaining, _ = strconv.Atoi(
		res.Header.Get("X-RateLimit-Remaining"),
	)
	res.Rate.Reset, _ = strconv.ParseInt(
		res.Header.Get("X-RateLimit-Reset"), 10, 64,
	)

	// snapshot the request rate limit
	e.github.SetRate(res.Rate)

	if res.Rate.Remaining == 0 {
		return res, fmt.Errorf("Github rate limit has been reached. please wait for %d until try again.", res.Rate.Reset)
	}
	
	// if an error is encountered, unmarshal and return the
	// error response.
	if res.Status > 300 {
		err := new(Error)
		json.NewDecoder(res.Body).Decode(err)
		return res, err
	}

	if out == nil {
		return res, nil
	}

	// if a json response is expected, parse and return
	// the json response.
	return res, json.NewDecoder(res.Body).Decode(out)
}

func (e *Error) Error() string {
	return e.Message
}
