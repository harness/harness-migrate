package converter

import (
	"encoding/json"

	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci"

	"github.com/ghodss/yaml"
)

func Convert(opts commons.Opts, input []byte) ([]byte, error) {

	p, err := circleci.Convert(opts, input)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	output, err := yaml.JSONToYAML(b)
	if err != nil {
		return nil, err
	}

	return output, nil
}
