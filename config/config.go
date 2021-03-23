package config

import (
	"os"
)

type Config struct {
	root   string
	bucket string
	region string
}

func ReadConfig() Config {
	return Config{
		root:   os.Getenv("ROOT_FOLDER"),
		bucket: os.Getenv("S3_BUCKET_NAME"),
		region: os.Getenv("AWS_REGION"),
	}
}

func (c Config) GetRoot() string {
	return c.root
}

func (c Config) GetBucketName() string {
	return c.bucket
}

func (c Config) GetRegion() string {
	return c.region
}
