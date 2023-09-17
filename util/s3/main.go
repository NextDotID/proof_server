package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nextdotid/proof_server/util"
)

var (
	s3Client *s3.Client
	bucketName string
)

func Init(awsConfig aws.Config) error {
	if s3Client == nil {
		s3Client = s3.NewFromConfig(awsConfig)
	}
	bucketName = util.GetE("S3_BUCKET", "")
	return nil
}

// ReadFromS3 should be called before Init()
func ReadFromS3(ctx context.Context, key string) (content []byte, err error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return []byte{}, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

// PutToS3 should be called before Init()
func PutToS3(ctx context.Context, key string, content []byte) (err error) {
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(content),
	})
	return err
}
