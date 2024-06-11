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
	"context"
	"fmt"
	"time"

	"github.com/harness/harness-migrate/types"
)

const (
	pollInterval = 5 * time.Second
	pollTimeout  = 5 * time.Minute
)

func (c *Importer) IsComplete() error {
	c.Tracer.Start("Performing import on harness code")
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()

	// todo: handle context cancelled due to timeout.
	err := c.pollOperationStatus(ctx)
	if err != nil {
		c.Tracer.Stop("Import error: %s", err)
		return err
	}
	c.Tracer.Stop("Import complete in %s seconds", time.Since(start).Seconds())

	return nil
}

func (c *Importer) pollOperationStatus(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		complete, err := c.checkOperationStatus()
		if err != nil {
			return err
		}
		if complete {
			fmt.Println("Operation is complete")
			return nil
		}
		fmt.Println("Operation is not complete, waiting to poll again...")
		time.Sleep(pollInterval)
	}

	return fmt.Errorf("operation did not complete within the expected time")
}

func (c *Importer) checkOperationStatus() (bool, error) {
	checkImport, err := c.Harness.HarnessCodeCheckImport(c.HarnessSpace, c.RequestId)
	if err != nil {
		return false, fmt.Errorf("error checking status: %w", err)
	}
	return checkImport.Status == types.RepoImportStatusComplete, nil
}
