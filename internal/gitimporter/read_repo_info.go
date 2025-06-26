package gitimporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harness/harness-migrate/types"
)

func (m *Importer) ReadRepoInfo(dir string) (types.Repository, error) {
	dirInfo, err := os.Stat(dir)
	if err != nil || !dirInfo.IsDir() {
		return types.Repository{}, ErrInvalidRepoDir
	}

	infoFile := filepath.Join(dir, types.InfoFileName)
	_, err = os.Stat(infoFile)
	if os.IsNotExist(err) {
		return types.Repository{}, ErrInvalidRepoDir
	}

	data, err := os.ReadFile(infoFile)
	if err != nil {
		return types.Repository{}, fmt.Errorf("failed to read %q content from %q: %w", types.InfoFileName, infoFile, err)
	}

	var repoInfo types.Repository
	if err := json.Unmarshal(data, &repoInfo); err != nil {
		return types.Repository{}, fmt.Errorf("error parsing repo info json: %w", err)
	}

	return repoInfo, nil
}
