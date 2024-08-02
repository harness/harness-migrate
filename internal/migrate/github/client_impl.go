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
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

const graphqlUrl = "https://api.github.com/graphql"

type (
	// wrapper wraps the Client to provide high level helper functions
	// for making http requests and unmarshalling the response.
	wrapper struct {
		*scm.Client
	}

	Export struct {
		github     *wrapper
		org        string
		repository string

		checkpointManager *checkpoint.CheckpointManager

		tracer     tracer.Tracer
		fileLogger *gitexporter.FileLogger

		userMap map[string]types.User
	}

	graphQLRequest struct {
		Query string `json:"query"`
	}
)

func (c *wrapper) ListPRComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments?%s", repoSlug, prNumber, encodeListOptions(opts))
	out := []*codeComment{}
	res, err := c.do(ctx, "GET", path, nil, &out)
	return convertPRCommentsList(out, repoSlug, prNumber), res, err
}

func (c *wrapper) GetUserByUserName(
	ctx context.Context,
	userName string,
) (*types.User, *scm.Response, error) {
	path := fmt.Sprintf("users/%s", userName)
	out := &types.User{}
	res, err := c.do(ctx, "GET", path, nil, &out)
	return out, res, err
}

func (c *wrapper) ListBranchRules(
	ctx context.Context,
	repoSlug string,
	logger gitexporter.Logger,
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
	res, err := c.do(ctx, "POST", graphqlUrl, body, &out)
	if out.Data.Repository.BranchProtectionRules.PageInfo.HasNextPage {
		res.Page.NextURL = out.Data.Repository.BranchProtectionRules.PageInfo.EndCursor
	}
	return convertBranchRulesList(out, repoSlug, logger), res, err
}

func (c *wrapper) ListBranchRulesets(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/rulesets?%s", repoSlug, encodeListOptions(opts))
	out := []*ruleSet{}
	res, err := c.do(ctx, "GET", path, nil, &out)
	return convertBranchRulesetsList(out), res, err
}

func (c *wrapper) FindBranchRuleset(
	ctx context.Context,
	repoSlug string,
	logger gitexporter.Logger,
	ruleID int,
) (*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/rulesets/%d", repoSlug, ruleID)
	out := new(detailedRuleSet)
	res, err := c.do(ctx, "GET", path, nil, &out)
	return convertBranchRuleset(out, repoSlug, logger), res, err
}

func (c *wrapper) do(ctx context.Context, method, path string, in, out interface{}) (*scm.Response, error) {
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
	res, err := c.Client.Do(ctx, req)
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
	c.Client.SetRate(res.Rate)

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
