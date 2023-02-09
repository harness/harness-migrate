package cache

import (
	harness "github.com/drone/spec/dist/go"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/utils"
)

const (
	saveStepType   = "plugin"
	saveNamePrefix = "save cache"
)

func ConvertSave(opts commons.Opts, c config.SaveCache) (*harness.Step, error) {
	name := saveNamePrefix
	if c.Name != nil && *c.Name != "" {
		name = *c.Name
	}

	backend := getBackend(opts)

	m := make(map[string]interface{})
	m["bucket"] = getBucket(opts)
	m["cache_key"] = getKey(&c.Key, []string{})
	m["rebuild"] = "true"
	m["mount"] = c.Paths
	m["exit_code"] = "true"
	m["archive_format"] = "tar"
	m["backend"] = backend
	m["backend_operation_timeout"] = "1800s"
	m["fail_restore_if_key_not_present"] = "false"
	if backend == "s3" {
		m["region"] = getRegion(opts)
		m["access_key"] = utils.ReplaceSecret(getS3AccessKey(opts), "access-key")
		m["secret_key"] = utils.ReplaceSecret(getS3SecretKey(opts), "secret-key")
	} else {
		m["json_key"] = utils.ReplaceSecret(getGCSJSONKey(opts), "json-key")
	}

	return &harness.Step{
		Name: name,
		Type: saveStepType,
		Spec: harness.StepPlugin{
			Image: "plugins/cache:1.4.6",
			With:  m,
		},
	}, nil
}
