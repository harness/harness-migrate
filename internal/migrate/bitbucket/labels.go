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

package bitbucket

import (
	"context"

	"github.com/harness/harness-migrate/internal/types"
	externalTypes "github.com/harness/harness-migrate/types"
)

func (e *Export) ListLabels(
	ctx context.Context,
	repoSlug string,
	opts types.ListOptions,
) (map[string]externalTypes.Label, error) {
	return nil, nil
}
