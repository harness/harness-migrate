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

package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/harness/harness-migrate/internal/util"
)

const filePath = "checkpoint.ckpt"

// CheckpointManager manages the checkpoints for different types of data
type CheckpointManager struct {
	mu                 sync.Mutex
	data               map[string]any
	checkpointLocation string
}

// NewCheckpointManager creates a new CheckpointManager
func NewCheckpointManager(checkpointLocation string) *CheckpointManager {
	return &CheckpointManager{
		checkpointLocation: checkpointLocation,
		data:               make(map[string]any),
	}
}

// SaveCheckpoint saves a checkpoint for a given key
func (cm *CheckpointManager) SaveCheckpoint(key string, value any) error {
	data, err := func() ([]byte, error) {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		cm.data[key] = value
		return json.Marshal(cm.data)
	}()

	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %v", err)
	}

	if err := util.WriteFile(filepath.Join(cm.checkpointLocation, filePath), data); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %v", err)
	}
	return nil
}

// LoadCheckpoint loads the checkpoint data from a file
func (cm *CheckpointManager) LoadCheckpoint() error {
	if _, err := os.Stat(filepath.Join(cm.checkpointLocation, filePath)); os.IsNotExist(err) {
		return nil // No checkpoint file exists
	}

	data, err := os.ReadFile(filepath.Join(cm.checkpointLocation, filePath))
	if err != nil {
		return fmt.Errorf("failed to read checkpoint file: %v", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	if err := json.Unmarshal(data, &cm.data); err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint: %v", err)
	}

	return nil
}

// GetCheckpoint gets the checkpoint for a given key
func (cm *CheckpointManager) GetCheckpoint(key string) (any, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	value, exists := cm.data[key]
	return value, exists
}

func CleanupCheckpoint(path string) error {
	err := os.Remove(filepath.Join(path, filePath))
	if err != nil {
		return fmt.Errorf("error deleting checkpoint file: %w", err)
	}
	return nil
}
