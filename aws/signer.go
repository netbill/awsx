package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Signer struct {
	//BucketName is the S3 bucketName name.
	bucketName string

	//putTTL is the duration a presigned PUT URL is valid for.
	putTTL time.Duration

	//getTTL is the duration a presigned GET URL is valid for.
	getTTL time.Duration

	//presign is the S3 presign client.
	presign *s3.PresignClient
}

// SignerConfig is the configuration for the Signer.
type SignerConfig struct {
	BucketName string
	PutTTL     time.Duration
	GetTTL     time.Duration
}

func NewSigner(awsCfg aws.Config, cfg SignerConfig) *Signer {
	return &Signer{
		bucketName: cfg.BucketName,
		putTTL:     cfg.PutTTL,
		getTTL:     cfg.GetTTL,
		presign:    s3.NewPresignClient(s3.NewFromConfig(awsCfg)),
	}
}

// PresignPut generates a presigned PUT URL for the given key and content type.
func (s *Signer) PresignPut(
	ctx context.Context,
	key string,
	contentType string,
) (string, error) {
	out, err := s.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucketName),
			Key:         aws.String(key),
			ContentType: aws.String(contentType),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = s.putTTL
		},
	)
	if err != nil {
		return "", err
	}

	return out.URL, nil
}

// PresignGet generates a presigned GET URL for the given key.
func (s *Signer) PresignGet(
	ctx context.Context,
	key string,
) (string, error) {

	out, err := s.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = s.getTTL
		},
	)
	if err != nil {
		return "", err
	}

	return out.URL, nil
}
