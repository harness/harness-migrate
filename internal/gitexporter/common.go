// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitexporter

import (
	"github.com/harness/harness-migrate/internal/types"
	externalTypes "github.com/harness/harness-migrate/types"

	"github.com/drone/go-scm/scm"
)

func mapPRComment(comments []*types.PRComment) []externalTypes.Comment {
	r := make([]externalTypes.Comment, len(comments))
	for i, c := range comments {
		r[i] = externalTypes.Comment{
			ID:          c.ID,
			Body:        c.Body,
			Created:     c.Created,
			Updated:     c.Updated,
			Author:      externalTypes.User(c.Author),
			ParentID:    c.ParentID,
			CodeComment: mapCodeComment(c.CodeComment),
		}
	}
	return r
}

func mapCodeComment(c *types.CodeComment) *externalTypes.CodeComment {
	if c == nil {
		return nil
	}
	return &externalTypes.CodeComment{
		Path:         c.Path,
		CodeSnippet:  externalTypes.Hunk(c.CodeSnippet),
		Side:         c.Side,
		HunkHeader:   c.HunkHeader,
		SourceSHA:    c.SourceSHA,
		MergeBaseSHA: c.MergeBaseSHA,
		Outdated:     c.Outdated,
	}
}

func mapBranchRules(rules []*types.BranchRule) []*externalTypes.BranchRule {
	r := make([]*externalTypes.BranchRule, len(rules))
	for i, b := range rules {
		definition := mapBranchRuleDefinition(b.Definition)
		r[i] = &externalTypes.BranchRule{
			ID:         b.ID,
			Identifier: b.Name,
			State:      string(b.State),
			Definition: definition,
			Pattern: externalTypes.BranchPattern{
				Default: b.Pattern.IncludeDefault,
				Include: b.Pattern.IncludedPatterns,
				Exclude: b.Pattern.ExcludedPatterns,
			},
			Created: b.Created,
			Updated: b.Updated,
		}
	}
	return r
}

func mapBranchRuleDefinition(d types.Definition) externalTypes.Definition {
	return externalTypes.Definition{
		Bypass: externalTypes.Bypass(d.Bypass),
		PullReq: externalTypes.PullReq{
			Approvals:    externalTypes.Approvals(d.Approvals),
			Comments:     externalTypes.Comments(d.Comments),
			Merge:        externalTypes.Merge(d.Merge),
			StatusChecks: externalTypes.StatusChecks(d.StatusChecks),
		},
		Lifecycle: externalTypes.Lifecycle(d.Lifecycle),
	}
}

func mapRepository(repository types.RepoResponse) externalTypes.Repository {
	return externalTypes.Repository{
		Slug:       repository.RepoSlug,
		ID:         repository.ID,
		Namespace:  repository.Namespace,
		Name:       repository.Name,
		Branch:     repository.Branch,
		Archived:   repository.Archived,
		Private:    repository.Private,
		Visibility: mapVisibility(repository.Visibility),
		Clone:      repository.Clone,
		CloneSSH:   repository.CloneSSH,
		Link:       repository.Link,
		Created:    repository.Created,
		Updated:    repository.Updated,
		IsEmpty:    repository.IsEmpty,
	}
}

func mapVisibility(visibility scm.Visibility) externalTypes.Visibility {
	switch visibility {
	case scm.VisibilityPublic:
		return externalTypes.VisibilityPublic
	case scm.VisibilityInternal:
		return externalTypes.VisibilityInternal
	case scm.VisibilityPrivate:
		return externalTypes.VisibilityPrivate
	default:
		return externalTypes.VisibilityUndefined
	}
}

func mapPR(request scm.PullRequest) externalTypes.PullRequest {
	return externalTypes.PullRequest{
		Number:  request.Number,
		Title:   request.Title,
		Body:    request.Body,
		SHA:     request.Sha,
		Ref:     request.Ref,
		Source:  request.Source,
		Target:  request.Target,
		Fork:    request.Fork,
		Link:    request.Link,
		Diff:    request.Diff,
		Draft:   request.Draft,
		Closed:  request.Closed,
		Merged:  request.Merged,
		Merge:   request.Merge,
		Base:    mapReference(request.Base),
		Head:    mapReference(request.Head),
		Author:  externalTypes.User(request.Author),
		Created: request.Created,
		Updated: request.Updated,
		Labels:  mapLabels(request.Labels),
	}
}

func mapReference(reference scm.Reference) externalTypes.Reference {
	return externalTypes.Reference{
		Name: reference.Name,
		Path: reference.Path,
		SHA:  reference.Sha,
	}
}

func mapLabels(labels []scm.Label) []externalTypes.Label {
	l := make([]externalTypes.Label, len(labels))
	for i, label := range labels {
		l[i] = externalTypes.Label{
			Name:  label.Name,
			Color: label.Color,
		}
	}
	return l
}

func mapHooks(hooks []*scm.Hook) []*externalTypes.Hook {
	h := make([]*externalTypes.Hook, len(hooks))
	for i, hook := range hooks {
		h[i] = &externalTypes.Hook{
			ID:         hook.ID,
			Identifier: hook.Name,
			Target:     hook.Target,
			Events:     hook.Events,
			Active:     hook.Active,
			SkipVerify: hook.SkipVerify,
		}
	}
	return h
}
