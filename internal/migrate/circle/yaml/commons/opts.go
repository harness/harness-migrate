package commons

type SecretID string

type Opts struct {
	Cache
	UploadArtifact
}

type UploadArtifact struct {
	GCS *GCS
	S3  *S3
}

type Cache struct {
	GCS *GCS
	S3  *S3
}

type GCS struct {
	Bucket  string
	JSONKey SecretID
}

type S3 struct {
	Bucket    string
	Region    string
	AccessKey SecretID
	SecretKey SecretID
}
