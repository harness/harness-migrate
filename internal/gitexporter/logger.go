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
