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

package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/harness"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type Export struct {
	bitbucket  *scm.Client
	workspace  string
	repository string

	checkpointManager *checkpoint.CheckpointManager

	tracer     tracer.Tracer
	fileLogger *gitexporter.FileLogger
	report     map[string]*report.Report

	userMap map[string]user
}

func (e *Export) GetUserByAccountID(
	ctx context.Context,
	accID string,
) (*user, *scm.Response, error) {
	path := fmt.Sprintf("2.0/workspaces/%s/members?fields=values.user.*&q=user.account_id=%s", e.workspace, accID)
	out := &user{}
	res, err := e.do(ctx, "GET", path, nil, &out)
	if res.Status == 404 {
		return nil, res, fmt.Errorf("user not found: %w", harness.ErrNotFound)
	}
	if res.Status == 403 {
		return nil, res, fmt.Errorf("failed to find user: %w", harness.ErrForbidden)
	}

	if err != nil {
		return nil, res, fmt.Errorf("failed to find user: %w", err)
	}

	return out, res, err
}

func (e *Export) ListPRComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	path := fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/comments?fields=%s&%s", repoSlug, prNumber, commentFields, encodeListOptions(opts))
	var out comments
	res, err := e.do(ctx, "GET", path, nil, &out)
	copyPagination(out.pagination, res)
	return convertPRCommentsList(out.Values, prNumber, repoSlug), res, err
}

func (e *Export) ListBranchRulesInternal(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	path := fmt.Sprintf("/2.0/repositories/%s/branch-restrictions?%s", repoSlug, encodeListOptions(opts))
	var out rules
	res, err := e.do(ctx, "GET", path, nil, &out)
	copyPagination(out.pagination, res)
	return e.convertBranchRules(ctx, out, repoSlug), res, err
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
	res, err := e.bitbucket.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse the bitbucket rate limit details.
	res.Rate.Limit, _ = strconv.Atoi(
		res.Header.Get("X-RateLimit-Limit"),
	)
	nearLimit, _ := strconv.ParseBool(
		res.Header.Get("X-RateLimit-NearLimit"),
	)

	if nearLimit {
		e.tracer.Debug().Log("Near Bitbucket rate limit. Less than 20%% of the requests remained.")
	}

	// if an error is encountered, unmarshal and return the
	// error response.
	if res.Status == 401 {
		return res, scm.ErrNotAuthorized
	} else if res.Status > 300 {
		err := new(Error)
		json.NewDecoder(res.Body).Decode(err)
		return res, err
	}

	if out == nil {
		return res, nil
	}

	// if raw output is expected, copy to the provided
	// buffer and exit.
	if w, ok := out.(io.Writer); ok {
		io.Copy(w, res.Body)
		return res, nil
	}

	// if a json response is expected, parse and return
	// the json response.
	return res, json.NewDecoder(res.Body).Decode(out)
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

	params.Set("pagelen", strconv.Itoa(limit))
	return params.Encode()
}

func copyPagination(from pagination, to *scm.Response) error {
	if to == nil {
		return nil
	}
	to.Page.NextURL = from.Next
	uri, err := url.Parse(from.Next)
	if err != nil {
		return err
	}
	page := uri.Query().Get("page")
	to.Page.First = 1
	to.Page.Next, _ = strconv.Atoi(page)
	return nil
}
