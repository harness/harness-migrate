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
	"net/url"
	"strconv"
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
		repos, resp, err := e.listGroupProjects(ctx, opts)
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

// glGroupProject mirrors GitLab's project object for GET /groups/:id/projects.
type glGroupProject struct {
	ID            int    `json:"id"`
	Path          string `json:"path"`
	PathNamespace string `json:"path_with_namespace"`
	DefaultBranch string `json:"default_branch"`
	Visibility    string `json:"visibility"`
	Archived      bool   `json:"archived"`
	WebURL        string `json:"web_url"`
	SSHURL        string `json:"ssh_url_to_repo"`
	HTTPURL       string `json:"http_url_to_repo"`
	Namespace     struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		FullPath string `json:"full_path"`
	} `json:"namespace"`
	Permissions struct {
		ProjectAccess glAccess `json:"project_access"`
		GroupAccess   glAccess `json:"group_access"`
	} `json:"permissions"`
}

type glAccess struct {
	AccessLevel int `json:"access_level"`
}

func (e *Export) listGroupProjects(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	q := url.Values{}
	q.Set("membership", "true")
	if e.includeSubgroups {
		q.Set("include_subgroups", "true")
	}
	if opts.Page != 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		q.Set("per_page", strconv.Itoa(opts.Size))
	}
	apiPath := fmt.Sprintf("api/v4/groups/%s/projects?%s", encode(e.group), q.Encode())
	var raw []*glGroupProject
	res, err := e.do(ctx, "GET", apiPath, nil, &raw)
	if err != nil {
		return nil, res, err
	}
	out := make([]*scm.Repository, 0, len(raw))
	for _, p := range raw {
		if p == nil {
			continue
		}
		out = append(out, convertGLGroupProject(p))
	}
	return out, res, nil
}

func convertGLGroupProject(from *glGroupProject) *scm.Repository {
	to := &scm.Repository{
		ID:         strconv.Itoa(from.ID),
		Name:       from.Path,
		Branch:     from.DefaultBranch,
		Archived:   from.Archived,
		Private:    scm.ConvertPrivate(from.Visibility),
		Visibility: scm.ConvertVisibility(from.Visibility),
		Clone:      from.HTTPURL,
		CloneSSH:   from.SSHURL,
		Link:       from.WebURL,
		Perm: &scm.Perm{
			Pull:  true,
			Push:  glCanPush(from),
			Admin: glCanAdmin(from),
		},
	}
	if path := from.Namespace.FullPath; path != "" {
		to.Namespace = path
	}
	if to.Namespace == "" {
		if parts := strings.SplitN(from.PathNamespace, "/", 2); len(parts) == 2 {
			to.Namespace = parts[1]
		}
	}
	return to
}

func glCanPush(proj *glGroupProject) bool {
	switch {
	case proj.Permissions.ProjectAccess.AccessLevel >= 30:
		return true
	case proj.Permissions.GroupAccess.AccessLevel >= 30:
		return true
	default:
		return false
	}
}

func glCanAdmin(proj *glGroupProject) bool {
	switch {
	case proj.Permissions.ProjectAccess.AccessLevel >= 40:
		return true
	case proj.Permissions.GroupAccess.AccessLevel >= 40:
		return true
	default:
		return false
	}
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
