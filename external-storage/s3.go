package externalstorage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.temporal.io/sdk/contrib/aws/s3driver"
	"go.temporal.io/sdk/contrib/aws/s3driver/awssdkv2"
	"go.temporal.io/sdk/converter"
)

// S3 endpoint and credentials used by every process in this sample (worker,
// starter, codec server, s3-mock). They point at the mock S3 server started by
// s3-mock/main.go.
const (
	S3Endpoint  = "http://localhost:5000"
	S3Bucket    = "temporal-payloads"
	S3AccessKey = "test"
	S3SecretKey = "test"
	S3Region    = "us-east-1"
)

// NewS3Client returns an aws-sdk-go-v2 S3 client configured for the local mock
// server. Path-style addressing (http://host/bucket/key) is used in place of
// the SDK's default virtual-hosted style (http://bucket.host/key) so the
// client doesn't have to resolve bucket.localhost — that resolution depends on
// the OS's handling of *.localhost (RFC 6761) and isn't guaranteed everywhere.
func NewS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(S3AccessKey, S3SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(S3Endpoint)
		o.UsePathStyle = true
	}), nil
}

// NewS3Driver builds the SDK's S3 StorageDriver, wired to the mock S3 client
// via the SDK's aws-sdk-go-v2 client adapter.
func NewS3Driver(ctx context.Context) (converter.StorageDriver, error) {
	s3Client, err := NewS3Client(ctx)
	if err != nil {
		return nil, err
	}
	return s3driver.NewDriver(s3driver.Options{
		Client: awssdkv2.NewClient(s3Client),
		Bucket: s3driver.StaticBucket(S3Bucket),
	})
}
