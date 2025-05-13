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
	"context"
	"fmt"
	"strings"

	"github.com/harness/harness-migrate/internal/checkpoint"
	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/types"

	"github.com/drone/go-scm/scm"
)

func (e *Export) ListRepositories(
	ctx context.Context,
	params types.ListOptions,
) ([]types.RepoResponse, error) {
	e.tracer.Start(common.MsgStartRepoList, "gitlab", "group", e.group)
	opts := scm.ListOptions{Page: params.Page, Size: params.Size}
	var allRepos []*scm.Repository

	checkpointDataKey := fmt.Sprintf(common.RepoCheckpointData, e.group)
	val, ok, err := checkpoint.GetCheckpointData[[]*scm.Repository](e.checkpointManager, checkpointDataKey)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointDataRead, err)
	}
	if ok && val != nil {
		allRepos = append(allRepos, val...)
	}

	checkpointPageKey := fmt.Sprintf(common.RepoCheckpointPage, e.group)
	checkpointPageIntfc, ok := e.checkpointManager.GetCheckpoint(checkpointPageKey)
	var checkpointPage int
	if ok && checkpointPageIntfc != nil {
		checkpointPage = int(checkpointPageIntfc.(float64))
		opts.Page = checkpointPage
	}

	// all pages are done
	if checkpointPage == -1 {
		e.tracer.Stop(common.MsgCompleteRepoList, len(allRepos))
		return common.MapRepository(allRepos), nil
	}

	if e.project != "" {
		repoSlug := strings.Join([]string{e.group, e.project}, "/")
		repo, _, err := e.gitlab.Repositories.Find(ctx, repoSlug)
		if err != nil {
			e.tracer.LogError(common.ErrListRepo, err)
			return nil, fmt.Errorf("failed to get the repo %s: %w", repoSlug, err)
		}

		allRepos = append(allRepos, repo)
		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allRepos)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRepoDataSave, repoSlug, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRepoPageSave, repoSlug, err)
		}

		e.tracer.Stop(common.MsgCompleteRepoList, 1)
		return common.MapRepository([]*scm.Repository{repo}), nil
	}

	for {
		repos, resp, err := e.gitlab.Repositories.ListNamespace(ctx, e.group, opts)
		if err != nil {
			e.tracer.LogError(common.ErrListRepo, err)
			return nil, fmt.Errorf("failed to get repos for group %s: %w", e.group, err)
		}
		allRepos = append(allRepos, repos...)

		err = e.checkpointManager.SaveCheckpoint(checkpointDataKey, allRepos)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRepoDataSave, e.group, err)
		}

		err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, resp.Page.Next)
		if err != nil {
			e.tracer.LogError(common.ErrCheckpointRepoPageSave, e.group, err)
		}

		if resp.Page.Next == 0 {
			break
		}
		opts.Page = resp.Page.Next
	}

	err = e.checkpointManager.SaveCheckpoint(checkpointPageKey, -1)
	if err != nil {
		e.tracer.LogError(common.ErrCheckpointRepoDataSave, e.group, err)
	}

	e.tracer.Stop(common.MsgCompleteRepoList, len(allRepos))
	return common.MapRepository(allRepos), nil
}

func (e *Export) GetLFSEnabledSettings(ctx context.Context, repoSlug string) (bool, error) {
	e.tracer.Start(common.MsgStartRepoLFSEnabled, repoSlug)
	res, _, err := e.projectInfo(ctx, repoSlug)
	if err != nil {
		e.tracer.LogError(common.ErrRepoLFSEnabled, err)
		e.tracer.Stop(common.MsgCompleteRepoLFSEnabled, repoSlug)
		return false, err
	}

	e.tracer.Stop(common.MsgCompleteRepoLFSEnabled, repoSlug)
	return res.LFSEnabled, nil
}
