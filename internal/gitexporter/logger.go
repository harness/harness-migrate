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

package gitexporter

import (
	"fmt"
	"path/filepath"

	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

type FileLogger struct {
	Location string
}

type Logger interface {
	Log(format string, args ...any) error
}

// Log writes the exporters' logs at the top level
func (f *FileLogger) Log(format string, args ...any) error {
	data := []byte(fmt.Sprintf(format+"\n", args...))
	err := util.AppendFile(filepath.Join(f.Location, types.ExporterLogsFileName), data)
	if err != nil {
		return fmt.Errorf("error writing log: %w", err)
	}
	return nil
}
