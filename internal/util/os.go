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

package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// DeleteDirsExcept deletes all directories in the specified path except for the specified file's path.
func DeleteDirsExcept(rootPath, exceptFilePath string) error {
	// Use a map to keep track of directories to delete
	dirsToDelete := make(map[string]struct{})

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error in deleting empty dirs: %w", err)
		}

		// Skip the excepted file's directory and higher levels
		if path == filepath.Dir(exceptFilePath) {
			return nil
		}

		// Collect directories to delete
		if info.IsDir() {
			dirsToDelete[path] = struct{}{}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error in walking path: %w", err)
	}

	// Delete directories in reverse order to ensure child directories are deleted before parents
	for dir := range dirsToDelete {
		_, err := os.ReadDir(dir)
		if err != nil {
			// Ignore if the directory does not exist
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("error reading directory %s: %w", dir, err)
		}

		err = os.RemoveAll(dir)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing directory %s: %w", dir, err)
		}
	}

	return nil
}
