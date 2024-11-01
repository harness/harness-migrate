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

	userMap map[string]types.User
}

func (e *Export) GetUserByUUID(
	ctx context.Context,
	uuid string,
) (*types.User, *scm.Response, error) {
	path := fmt.Sprintf("2.0/user/emails", uuid)
	out := &types.User{}
	res, err := e.do(ctx, "GET", path, nil, &out)
	return out, res, err
}

func (e *Export) ListPullRequestComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	path := fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/comments?%s", repoSlug, prNumber, encodeListOptions(opts))
	var out comments
	res, err := e.do(ctx, "GET", path, nil, &out)
	return convertPRCommentsList(out.Values), res, err
}

func (e *Export) ListBranchRulesInternal(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
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

	// parse the github rate limit details.
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
