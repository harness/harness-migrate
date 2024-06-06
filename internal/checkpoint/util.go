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

package checkpoint

import (
	"encoding/json"
	"fmt"
)

// GetCheckpointData retrieves the checkpoint data for a given key and unmarshals it into the provided type
func GetCheckpointData[T any](cm *CheckpointManager, key string) (T, bool, error) {
	var result T
	value, exists := cm.GetCheckpoint(key)
	if !exists {
		return result, false, nil
	}

	// Marshal the intermediate value back to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return result, false, fmt.Errorf("failed to marshal intermediate value: %v", err)
	}

	// Unmarshal the JSON into the result type
	if err := json.Unmarshal(data, &result); err != nil {
		return result, false, fmt.Errorf("failed to unmarshal into target type: %v", err)
	}

	return result, true, nil
}
