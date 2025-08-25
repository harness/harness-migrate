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

package gitimporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/harness/harness-migrate/internal/common"
	"github.com/harness/harness-migrate/internal/report"
	"github.com/harness/harness-migrate/types"
)

func (m *Importer) ImportPullRequests(
	repoRef string,
	repoFolder string,
) error {
	m.Tracer.Start(common.MsgStartImportPRs, repoRef)
	prDir := filepath.Join(repoFolder, types.PullRequestDir)
	in, err := m.readPRs(prDir)
	if err != nil {
		m.Tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to read pull requests and comments from %q: %w", prDir, err)
	}

	if len(in) == 0 {
		m.Report[repoRef].ReportMetric(report.ReportTypePRs, len(in))
		m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)
		return nil
	}

	// from := //61358   // 60646, 59645, 58744, 57743, 56742, 55741, 54740, 53739, 50738, 47737, 44736,41735,38734,35733,32232,28731,25230,21729,17728,12727,7726
	// tryFor := //60647 // 59646, 58745, 57744, 56743, 55742, 54741, 53740, 50739, 47738, 44737, 41736,38735,35734,32233,28732,25231,21730, 17729,12728, 7727,14
	//listpr := []int{30484, 30448, 30444, 30329, 30158, 30104, 30103, 30075, 30074, 29994}
	//MAX for core-ui is
	// from := 3294 // 23632 //21968 , 21967, 19966, 17965, 15964, 14963, 13962, 13161, 12160, 11159, 9158, 7157 , 5156
	// tryFor := 1  //22969 //20968 , 19967, 17966, 15965, 14964, 13963, 13162, 12161, 11160, 9159, 7158, 5157 , 3156

	from := 999 //2099   //4884   //11364  //8999 // 28748   //33749   // 8999 //33745, 28744
	tryFor := 1 //1000 //2100 //4885 // 7886 // 28745, 18000,
	skipPRs := []int{11293, 11269, 11232, 11262, 11252, 10904, 10453, 11181, 11123, 10190}
	var subPRs []*types.PullRequestData

	for _, pr := range in {
		// Skip forked open PRs (PRs where fork field has a value and PR is not merged or closed)
		isForkedOpenPR := pr.PullRequest.Fork != "" && !pr.PullRequest.Closed && !pr.PullRequest.Merged

		if pr.PullRequest.Number <= from && pr.PullRequest.Number >= tryFor {
			if !slices.Contains(skipPRs, pr.PullRequest.Number) && !isForkedOpenPR {
				subPRs = append(subPRs, &types.PullRequestData{
					PullRequest: pr.PullRequest,
					Comments:    pr.Comments,
				})
			} else if isForkedOpenPR {
				// Log that we're skipping a forked open PR using the tracer
				m.Tracer.Debug().Log("Skipping forked open PR #%d from %s", pr.PullRequest.Number, pr.PullRequest.Fork)
			}
		}
	}

	if err := m.Harness.ImportPRs(repoRef, &types.PRsImportInput{PullRequestData: subPRs}); err != nil {
		m.Tracer.Stop(common.ErrImportPRs, repoRef, err)
		return fmt.Errorf("failed to import pull requests and comments for repo '%s' : %w",
			repoRef, err)
	}
	m.Report[repoRef].ReportMetric(report.ReportTypePRs, len(in))
	m.Tracer.Stop(common.MsgCompleteImportPRs, len(in), repoRef)

	return nil
}

func (m *Importer) readPRs(prFolder string) ([]*types.PullRequestData, error) {
	pattern := regexp.MustCompile(`^pr\d+\.json$`)
	prOut := make([]*types.PullRequestData, 0)

	if _, err := os.Stat(prFolder); os.IsNotExist(err) {
		return prOut, nil
	}

	fileEntries, err := os.ReadDir(prFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", types.PullRequestDir, err)
	}

	for _, entry := range fileEntries {
		if entry.IsDir() || !pattern.MatchString(entry.Name()) {
			continue
		}

		prFile := entry.Name()
		data, err := ioutil.ReadFile(filepath.Join(prFolder, prFile))
		if err != nil {
			return nil, fmt.Errorf("failed to read %q content: %w", prFile, err)
		}

		var prs []*types.PullRequestData
		if err := json.Unmarshal(data, &prs); err != nil {
			return nil, fmt.Errorf("error parsing repo pull request json: %w", err)
		}

		prOut = append(prOut, prs...)
	}

	return prOut, nil
}
