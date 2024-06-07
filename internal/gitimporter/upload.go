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
	"fmt"
	"time"

	"github.com/harness/harness-migrate/internal/harness"
)

func (c *Importer) UploadZip() (*harness.RepositoriesImportOutput, error) {
	c.Tracer.Start("starting uploading zip file")
	in := &harness.RepositoriesImportInput{}

	start := time.Now()
	repositoriesImportOutput, err := c.Harness.UploadHarnessCodeZip(c.HarnessSpace, c.ZipFileLocation, c.RequestId, in)
	if err != nil {
		c.Tracer.Stop("error uploading zip")
		return nil, fmt.Errorf("error uploading zip: %w", err)
	}
	c.Tracer.Stop("zip upload complete in %d seconds", time.Since(start).Seconds())
	return repositoriesImportOutput, nil
}
