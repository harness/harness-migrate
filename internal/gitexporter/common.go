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
)

func MapPRComment(comments []*types.PRComment) []externalTypes.Comment {
	r := make([]externalTypes.Comment, len(comments))
	for i, c := range comments {
		r[i] = externalTypes.Comment{
			Comment:     c.Comment,
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
		SourceSha:    c.SourceSha,
		MergeBaseSha: c.MergeBaseSha,
	}
}

func MapBranchRules(rules []*types.BranchRule) []externalTypes.BranchRule {
	r := make([]externalTypes.BranchRule, len(rules))
	for i, b := range rules {
		r[i] = externalTypes.BranchRule{
			ID:               b.ID,
			Name:             b.Name,
			RuleDefinition:   mapRuleDefinition(b.RuleDefinition),
			IncludeDefault:   b.IncludeDefault,
			IncludedPatterns: b.IncludedPatterns,
			ExcludedPatterns: b.ExcludedPatterns,
			BypassGroups:     b.BypassGroups,
			BypassKeys:       b.BypassKeys,
		}
	}
	return r
}

func mapRuleDefinition(d types.RuleDefinition) externalTypes.RuleDefinition {
	return externalTypes.RuleDefinition{
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
