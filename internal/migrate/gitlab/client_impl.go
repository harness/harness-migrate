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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"
	externalTypes "github.com/harness/harness-migrate/types"

	"github.com/drone/go-scm/scm"
)

type (
	Export struct {
		gitlab  *scm.Client
		group   string
		project string

		checkpointManager *checkpoint.CheckpointManager

		tracer     tracer.Tracer
		fileLogger *gitexporter.FileLogger
		report     map[string]*report.Report

		userMap map[string]user
	}
)

func (e *Export) GetHighestMRNumber(
	ctx context.Context,
	repoSlug string,
) (int, error) {
	path := fmt.Sprintf("api/v4/projects/%s/merge_requests?state=all&sort=desc&per_page=1", encode(repoSlug))

	var out []struct {
		IID int `json:"iid"`
	}
	_, err := e.do(ctx, "GET", path, nil, &out)
	if err != nil {
		return 0, err
	}

	if len(out) == 0 {
		return 0, nil
	}

	return out[0].IID, nil
}

func (e *Export) FindPR(
	ctx context.Context,
	repoSlug string,
	prNumber int,
) (*types.PRResponse, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/projects/%s/merge_requests/%d", encode(repoSlug), prNumber)
	var out mergeRequest
	res, err := e.do(ctx, "GET", path, nil, &out)
	if res.Status == 404 {
		return nil, res, fmt.Errorf("merge request not found: %w", harness.ErrNotFound)
	}
	return convertPR(out), res, err
}

func (e *Export) projectInfo(
	ctx context.Context,
	repoSlug string,
) (*repoInfo, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/projects/%s", encode(repoSlug))
	out := &repoInfo{}
	res, err := e.do(ctx, "GET", path, nil, &out)
	return out, res, err
}

func (e *Export) ListPRComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/projects/%s/merge_requests/%d/discussions?%s", encode(repoSlug), prNumber, encodeListOptions(opts))
	var out []*discussion
	res, err := e.do(ctx, "GET", path, nil, &out)
	return e.convertPRCommentsList(out, prNumber), res, err
}

func (e *Export) ListRepoLabels(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]externalTypes.Label, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/projects/%s/labels?include_ancestor_groups&%s", encode(repoSlug), encodeListOptions(opts))
	var out []*types.LabelResponse
	res, err := e.do(ctx, "GET", path, nil, &out)
	return convertLabels(out), res, err
}

func (e *Export) ListBranchRulesInternal(ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/projects/%s/protected_branches?%s", encode(repoSlug), encodeListOptions(opts))
	var out []*branchRule
	res, err := e.do(ctx, "GET", path, nil, &out)
	return e.convertBranchRules(ctx, out, repoSlug), res, err
}

func (e *Export) GetMergeRequestRules(ctx context.Context, repoSlug string) (*types.BranchRule, error) {
	path := fmt.Sprintf("api/v4/projects/%s", encode(repoSlug))
	var out project
	_, err := e.do(ctx, "GET", path, nil, &out)
	return e.convertMergeRequestRule(out.mergeRequestRules), err
}

func (e *Export) GetUserByUserName(
	ctx context.Context,
	userName string,
) (*types.User, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/users?username=%s", userName)
	var out []*types.User
	res, err := e.do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, res, fmt.Errorf("failed to find user: %w", err)
	}

	if len(out) == 0 {
		return nil, res, fmt.Errorf("user response is empty: %w", harness.ErrNotFound)
	}

	return out[0], res, err
}

func (e *Export) GetUserByID(
	ctx context.Context,
	userID int,
) (*user, *scm.Response, error) {
	path := fmt.Sprintf("api/v4/users/%d", userID)
	var out *user
	res, err := e.do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, res, fmt.Errorf("failed to find user: %w", err)
	}
	if res.Status == 404 {
		return nil, res, fmt.Errorf("user not found: %w", harness.ErrNotFound)
	}

	return out, res, err
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
	res, err := e.gitlab.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse the request id.
	res.ID = res.Header.Get("X-Request-Id")

	// parse the rate limit details.
	res.Rate.Limit, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Limit"),
	)
	res.Rate.Remaining, _ = strconv.Atoi(
		res.Header.Get("RateLimit-Remaining"),
	)
	res.Rate.Reset, _ = strconv.ParseInt(
		res.Header.Get("RateLimit-Reset"), 10, 64,
	)

	// snapshot the request rate limit
	e.gitlab.SetRate(res.Rate)

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

func encode(s string) string {
	return strings.Replace(s, "/", "%2F", -1)
}

func encodeListOptions(opts types.ListOptions) string {
	params := url.Values{}
	limit := common.DefaultLimit
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		limit = opts.Size
	}

	params.Set("per_page", strconv.Itoa(limit))
	return params.Encode()
}
