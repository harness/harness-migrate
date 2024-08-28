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

package github

import "github.com/harness/harness-migrate/internal/types"

func convertLabels(labels []*label) []*types.Label {
	allLabels := make([]*types.Label, len(labels))
	for i, label := range labels {
		allLabels[i] = &types.Label{
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
		}
	}

	return allLabels
}
