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
