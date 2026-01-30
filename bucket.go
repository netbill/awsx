package awsx

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Bucket struct {
	name    string
	client  *s3.Client
	presign *s3.PresignClient
}

func New(
	name string,
	client *s3.Client,
	presign *s3.PresignClient,
) *Bucket {
	return &Bucket{
		name:    name,
		client:  client,
		presign: presign,
	}
}

func (b *Bucket) PresignPut(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (uploadURL, getURL string, err error) {
	putOut, err := b.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(b.name),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(ttl),
	)
	if err != nil {
		return "", "", err
	}

	getOut, err := b.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(b.name),
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

func (b *Bucket) HeadObject(
	ctx context.Context,
	key string,
) (*s3.HeadObjectOutput, error) {
	output, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})

	return output, err
}

func (b *Bucket) GetObject(
	ctx context.Context,
	key string,
) (body io.ReadCloser, size int64, err error) {
	output, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &b.name,
		Key:    &key,
	})
	if err != nil {
		return nil, 0, err
	}

	if output.Body == nil {
		return nil, 0, fmt.Errorf("s3 get object: body is nil")
	}

	if output.ContentLength == nil {
		_ = output.Body.Close()
		return nil, 0, fmt.Errorf("s3 get object: content length is nil")
	}

	return output.Body, *output.ContentLength, nil
}

func (b *Bucket) GetObjectRange(
	ctx context.Context,
	key string,
	bytes int64,
) (body io.ReadCloser, size int64, err error) {
	if bytes <= 0 {
		return b.GetObject(ctx, key)
	}

	rng := "bytes=0-" + strconv.FormatInt(bytes-1, 10)

	output, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
		Range:  aws.String(rng),
	})
	if err != nil {
		return nil, 0, err
	}

	if output.Body == nil {
		return nil, 0, fmt.Errorf("s3 get object range: body is nil")
	}

	if output.ContentRange == nil {
		_ = output.Body.Close()
		return nil, 0, fmt.Errorf("s3 get object range: content-range is nil")
	}

	cr := strings.TrimSpace(*output.ContentRange)
	slash := strings.LastIndexByte(cr, '/')
	if slash < 0 || slash == len(cr)-1 {
		_ = output.Body.Close()
		return nil, 0, fmt.Errorf("s3 get object range: invalid content-range %q", cr)
	}

	totalStr := strings.TrimSpace(cr[slash+1:])
	total, err := strconv.ParseInt(totalStr, 10, 64)
	if err != nil {
		_ = output.Body.Close()
		return nil, 0, fmt.Errorf("s3 get object range: invalid content-range total %q: %w", totalStr, err)
	}

	return output.Body, total, nil
}

func (b *Bucket) CopyObject(
	ctx context.Context,
	fromKey, toKey string,
) (string, error) {
	_, err := b.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(b.name),
		Key:        aws.String(toKey),
		CopySource: aws.String(b.name + "/" + fromKey),
	})
	if err != nil {
		return "", err
	}

	return toKey, nil
}

func (b *Bucket) DeleteObject(
	ctx context.Context,
	key string,
) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	return err
}
