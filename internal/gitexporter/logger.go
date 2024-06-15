package gitexporter

import (
	"path/filepath"

	"github.com/harness/harness-migrate/internal/util"
	"github.com/harness/harness-migrate/types"
)

type Logger interface {
	Log(data []byte) error
}

// log writes the exporters' logs at the top level
func (e *Exporter) Log(data []byte) error {
	err := util.AppendFile(filepath.Join(e.zipLocation, types.ExporterLogsFileName), data)
	if err != nil {
		return err
	}
	return nil
}
