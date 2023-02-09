package cache

import (
	"github.com/harness/harness-migrate/internal/migrate/circle/yaml/commons"
)

func getKey(key *string, keys []string) string {
	if key != nil && *key != "" {
		return *key
	}

	if len(keys) > 0 {
		return keys[0]
	}
	return "replace-cache-key"
}

func getBucket(opts commons.Opts) string {
	bucket := ""
	if opts.Cache.GCS != nil {
		bucket = opts.Cache.GCS.Bucket
	} else if opts.Cache.S3 != nil {
		bucket = opts.Cache.S3.Bucket
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
		region = opts.Cache.S3.Region
	}
	if region == "" {
		region = "replace-region"
	}
	return region
}

func getBackend(opts commons.Opts) string {
	if opts.Cache.GCS != nil {
		return "gcs"
	} else if opts.Cache.S3 != nil {
		return "s3"
	}
	return "gcs"
}

func getGCSJSONKey(opts commons.Opts) string {
	if opts.Cache.GCS != nil {
		return string(opts.Cache.GCS.JSONKey)
	}
	return ""
}

func getS3AccessKey(opts commons.Opts) string {
	if opts.Cache.S3 != nil {
		return string(opts.Cache.S3.AccessKey)
	}
	return ""
}

func getS3SecretKey(opts commons.Opts) string {
	if opts.Cache.S3 != nil {
		return string(opts.Cache.S3.SecretKey)
	}
	return ""
}
