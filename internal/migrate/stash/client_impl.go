package stash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type (
	// wrapper wraps the Client to provide high level helper functions
	// for making http requests and unmarshalling the response.
	wrapper struct {
		*scm.Client
	}

	Export struct {
		stash    *wrapper
		stashOrg string

		checkpointManager *checkpoint.CheckpointManager

		tracer tracer.Tracer
	}
)

func New(
	client *scm.Client,
	org string,
	checkpointer *checkpoint.CheckpointManager,
	tracer tracer.Tracer,
) *Export {
	return &Export{
		stash:             &wrapper{client},
		stashOrg:          org,
		checkpointManager: checkpointer,
		tracer:            tracer,
	}
}

func (c *wrapper) ListPRComments(ctx context.Context, repoSlug string, prNumber int, opts types.ListOptions, tracer tracer.Tracer) ([]*types.PRComment, *scm.Response, error) {
	namespace, name := scm.Split(repoSlug)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests/%d/activities?%s", namespace, name, prNumber, encodeListOptions(opts))
	out := new(activities)
	res, err := c.do(ctx, "GET", path, out)
	if !out.pagination.LastPage {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return convertPullRequestCommentsList(out.Values, tracer), res, err
}

func (c *wrapper) do(ctx context.Context, method, path string, out interface{}) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
		Header: map[string][]string{
			"Accept":            {"application/json"},
			"x-atlassian-token": {"no-check"},
		},
	}

	// execute the http request
	res, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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
