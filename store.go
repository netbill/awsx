package awsx

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service struct {
	bucketName string
	client     *s3.Client
	presign    *s3.PresignClient
}

type Config struct {
	BucketName string
	AwsCfg     aws.Config
	PutTTL     time.Duration
	GetTTL     time.Duration
}

func New(
	bucketName string,
	client *s3.Client,
	presign *s3.PresignClient,
) *Service {
	return &Service{
		bucketName: bucketName,
		client:     client,
		presign:    presign,
	}
}

func (s *Service) PresignPut(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (uploadURL, getURL string, err error) {
	putOut, err := s.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(ttl),
	)
	if err != nil {
		return "", "", err
	}

	getOut, err := s.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(ttl),
		func(opts *s3.PresignOptions) {
			opts.Expires = ttl
		},
	)
	if err != nil {
		return "", "", err
	}

	return putOut.URL, getOut.URL, nil
}

func (s *Service) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &key,
	})
	if err != nil {
		return io.ReadCloser(nil), err
	}

	return obj.Body, nil
}

func (s *Service) GetObjectRange(ctx context.Context, key string, maxBytes int64) (io.ReadCloser, error) {
	if maxBytes <= 0 {
		return s.GetObject(ctx, key)
	}

	rng := "bytes=0-" + strconv.FormatInt(maxBytes-1, 10)

	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Range:  aws.String(rng),
	})
	if err != nil {
		return nil, err
	}

	return obj.Body, nil
}

func (s *Service) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	return s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
}

func (s *Service) CopyObject(ctx context.Context, tmplKey, finalKey string) (string, error) {
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		Key:        aws.String(finalKey),
		CopySource: aws.String(s.bucketName + "/" + tmplKey),
	})
	if err != nil {
		return "", err
	}

	return finalKey, nil
}

func (s *Service) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}
