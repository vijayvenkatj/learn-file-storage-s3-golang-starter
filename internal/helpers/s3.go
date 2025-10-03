package helpers

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)


func GeneratePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {

	preSignClient := s3.NewPresignClient(s3Client)

	resp, err := preSignClient.PresignGetObject(context.TODO(),&s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	},s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return resp.URL, nil
}