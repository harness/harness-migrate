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
	"strings"

	"github.com/harness/harness-migrate/internal/types"
	externalTypes "github.com/harness/harness-migrate/types"
)

func convertLabels(labels []*types.LabelResponse) []externalTypes.Label {
	allLabels := make([]externalTypes.Label, len(labels))
	for i, label := range labels {
		// Gitlab scoped labels are in a form of key::value or key1::key2::value
		name := label.Name
		value := ""
		keyIndex := strings.LastIndex(label.Name, "::")
		if keyIndex != -1 {
			name = label.Name[:keyIndex]
			if keyIndex+2 < len(label.Name) {
				value = label.Name[keyIndex+2:]
			}
		}
		allLabels[i] = externalTypes.Label{
			Name:        name,
			Description: label.Description,
			Color:       label.Color,
			Value:       value,
		}
	}

	return allLabels
}
