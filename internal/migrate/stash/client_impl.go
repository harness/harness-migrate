package stash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/gitexporter"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/internal/tracer"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

type (
	Export struct {
		stash      *scm.Client
		project    string
		repository string

		checkpointManager *checkpoint.CheckpointManager

		tracer     tracer.Tracer
		fileLogger gitexporter.Logger
		report     map[string]*report.Report
	}
)

func New(
	client *scm.Client,
	project string,
	repo string,
	checkpointer *checkpoint.CheckpointManager,
	logger *gitexporter.FileLogger,
	tracer tracer.Tracer,
	reporter map[string]*report.Report,
) *Export {
	return &Export{
		stash:             client,
		project:           project,
		repository:        repo,
		checkpointManager: checkpointer,
		tracer:            tracer,
		fileLogger:        logger,
		report:            reporter,
	}
}

func (e *Export) ListPRComments(
	ctx context.Context,
	repoSlug string,
	prNumber int,
	opts types.ListOptions,
) ([]*types.PRComment, *scm.Response, error) {
	namespace, name := scm.Split(repoSlug)
	path := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests/%d/activities?%s",
		namespace, name, prNumber, encodeListOptions(opts))
	out := new(activities)
	res, err := e.do(ctx, "GET", path, out)
	if !out.pagination.LastPage {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return convertPullRequestCommentsList(out.Values), res, err
}

func (e *Export) ListBranchRulesInternal(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) ([]*types.BranchRule, *scm.Response, error) {
	namespace, name := scm.Split(repoSlug)
	branchModels, _, _ := e.listBranchModels(ctx, namespace, name)
	path := fmt.Sprintf("rest/branch-permissions/2.0/projects/%s/repos/%s/restrictions?%s",
		namespace, name, encodeListOptions(opts))
	out := new(branchPermissions)
	res, err := e.do(ctx, "GET", path, out)
	if !out.pagination.LastPage {
		res.Page.First = 1
		res.Page.Next = opts.Page + 1
	}
	return e.convertBranchRulesList(out.Values, branchModels, repoSlug), res, err
}

func (e *Export) listBranchModels(
	ctx context.Context,
	namespace string,
	repoName string,
) (map[string]modelValue, *scm.Response, error) {
	path := fmt.Sprintf("rest/branch-utils/1.0/projects/%s/repos/%s/branchmodel/configuration", namespace, repoName)
	out := new(branchModels)
	res, err := e.do(ctx, "GET", path, out)
	return convertBranchModelsMap(*out), res, err
}

func (e *Export) do(ctx context.Context, method, path string, out any) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
		Header: map[string][]string{
			"Accept":            {"application/json"},
			"x-atlassian-token": {"no-check"},
		},
	}

	// execute the http request
	res, err := e.stash.Do(ctx, req)
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
