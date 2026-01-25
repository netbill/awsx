package imgx

import (
	"context"
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
	contentLength int64,
	ttl time.Duration,
) (uploadURL, getURL string, err error) {
	putOut, err := s.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:        aws.String(s.bucketName),
			Key:           aws.String(key),
			ContentLength: aws.Int64(contentLength),
			//Expires:       aws.Time(time.Now().UTC().Add(ttl)),
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
