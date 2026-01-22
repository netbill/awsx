package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Store struct {
	bucketName string
	client     *s3.Client
}

func NewStore(awsCfg aws.Config, BucketName string) *Store {
	return &Store{
		bucketName: BucketName,
		client:     s3.NewFromConfig(awsCfg),
	}
}

type HeadResult struct {
	// Size is the size of the object in bytes.
	Size int64

	// ContentType is the content type of the object.
	ContentType string

	// ETag is the entity tag of the object.
	ETag string
}

// Head retrieves metadata about the object stored at the given key.
func (s *Store) Head(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Copy duplicates the object from srcKey to dstKey.
func (s *Store) Copy(ctx context.Context, srcKey, dstKey string) error {
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		Key:        aws.String(dstKey),
		CopySource: aws.String(s.bucketName + "/" + srcKey),
	})
	return err
}

// Delete removes the object stored at the given key.
func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}
