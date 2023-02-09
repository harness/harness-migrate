package artifact

import (
	"fmt"

	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/config"
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/converter/circleci/utils"

	harness "github.com/drone/spec/dist/go"
)

const (
	stepType   = "plugin"
	namePrefix = "upload"
)

func Convert(opts commons.Opts, c config.StoreArtifacts) (*harness.Step, error) {
	name := namePrefix
	if c.Name != nil && *c.Name != "" {
		name = *c.Name
	}

	backend := getBackend(opts)
	bucket := getBucket(opts)
	m := make(map[string]interface{})
	m["bucket"] = bucket
	m["source"] = []string{c.Path}
	if backend == "s3" {
		m["region"] = getRegion(opts)
		m["access_key"] = utils.ReplaceSecret(getS3AccessKey(opts), "access-key")
		m["secret_key"] = utils.ReplaceSecret(getS3SecretKey(opts), "secret-key")
	} else {
		m["json_key"] = utils.ReplaceSecret(getGCSJSONKey(opts), "json-key")
		m["target"] = fmt.Sprintf("%s/", bucket)
	}

	return &harness.Step{
		Name: name,
		Type: stepType,
		Spec: harness.StepPlugin{
			Image: getImage(backend),
			With:  m,
		},
	}, nil
}

func getImage(backend string) string {
	if backend == "s3" {
		return "plugins/s3:1.2.0"
	} else {
		return "plugins/gcs:1.3.0"
	}
}

func getBucket(opts commons.Opts) string {
	bucket := ""
	if opts.UploadArtifact.GCS != nil {
		bucket = opts.UploadArtifact.GCS.Bucket
	} else if opts.UploadArtifact.S3 != nil {
		bucket = opts.UploadArtifact.S3.Bucket
	}
	if bucket == "" {
		bucket = "replace-bucket"
	}
	return bucket
}

func getRegion(opts commons.Opts) string {
	backend := getBackend(opts)

	region := ""
	if backend == "s3" {
		region = opts.UploadArtifact.S3.Region
	}
	return utils.ReplaceString(region, "region")
}

func getBackend(opts commons.Opts) string {
	if opts.UploadArtifact.GCS != nil {
		return "gcs"
	} else if opts.UploadArtifact.S3 != nil {
		return "s3"
	}
	return "gcs"
}

func getGCSJSONKey(opts commons.Opts) string {
	if opts.UploadArtifact.GCS != nil {
		return string(opts.UploadArtifact.GCS.JSONKey)
	}
	return ""
}

func getS3AccessKey(opts commons.Opts) string {
	if opts.UploadArtifact.S3 != nil {
		return string(opts.UploadArtifact.S3.AccessKey)
	}
	return ""
}

func getS3SecretKey(opts commons.Opts) string {
	if opts.UploadArtifact.S3 != nil {
		return string(opts.UploadArtifact.S3.SecretKey)
	}
	return ""
}
