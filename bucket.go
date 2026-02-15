package awsx

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Bucket interface {
	PresignPut(ctx context.Context, key string, ttl time.Duration) (uploadURL, getURL string, err error)
	HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	GetObjectRange(ctx context.Context, key string, bytes int64) (io.ReadCloser, error)
	CopyObject(ctx context.Context, fromKey, toKey string) error
	DeleteObject(ctx context.Context, key string) error
}

type bucket struct {
	name    string
	client  *s3.Client
	presign *s3.PresignClient
}

func New(name string, client *s3.Client, presign *s3.PresignClient) Bucket {
	return &bucket{name: name, client: client, presign: presign}
}

func (b *bucket) PresignPut(ctx context.Context, key string, ttl time.Duration) (uploadURL, getURL string, err error) {
	putOut, err := b.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{Bucket: aws.String(b.name), Key: aws.String(key)},
		s3.WithPresignExpires(ttl),
	)
	if err != nil {
		return "", "", err
	}

	getOut, err := b.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{Bucket: aws.String(b.name), Key: aws.String(key)},
		s3.WithPresignExpires(ttl),
	)
	if err != nil {
		return "", "", err
	}

	return putOut.URL, getOut.URL, nil
}

func (b *bucket) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	res, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, Wrap(err)
	}
	return res, nil
}

func (b *bucket) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, Wrap(err)
	}
	if out.Body == nil {
		return nil, fmt.Errorf("awsx: get object: body is nil")
	}
	return out.Body, nil
}

func (b *bucket) GetObjectRange(ctx context.Context, key string, bytes int64) (io.ReadCloser, error) {
	if bytes <= 0 {
		return b.GetObject(ctx, key)
	}

	rng := "bytes=0-" + strconv.FormatInt(bytes-1, 10)
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
		Range:  aws.String(rng),
	})
	if err != nil {
		return nil, Wrap(err)
	}
	if out.Body == nil {
		return nil, fmt.Errorf("awsx: get object range: body is nil")
	}
	return out.Body, nil
}

func (b *bucket) CopyObject(ctx context.Context, fromKey, toKey string) error {
	_, err := b.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(b.name),
		Key:        aws.String(toKey),
		CopySource: aws.String(b.name + "/" + fromKey),
	})
	return Wrap(err)
}

func (b *bucket) DeleteObject(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	return Wrap(err)
}
