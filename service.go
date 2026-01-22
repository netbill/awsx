package imgx

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/netbill/imgx/data/sqlc"
)

type SignerS3 interface {
	PresignPut(
		ctx context.Context,
		key string,
		contentType string,
	) (string, error)

	PresignGet(
		ctx context.Context,
		key string,
	) (string, error)
}

type StoreS3 interface {
	Head(ctx context.Context, key string) (*s3.HeadObjectOutput, error)
	Copy(ctx context.Context, srcKey, dstKey string) error
	Delete(ctx context.Context, key string) error
}

type Database interface {
	Queries(ctx context.Context) *sqlc.Queries
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	Signer SignerS3
	S3     StoreS3
	DB     Database
}

func NewService(signer SignerS3, store StoreS3, db Database) Service {
	return Service{
		Signer: signer,
		S3:     store,
		DB:     db,
	}
}

func buildObjectKey(prefix, purpose string, ownerID, sessionID, fileID uuid.UUID) string {
	p := strings.Trim(prefix, "/")
	if p != "" {
		p += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s/%s", p, purpose, ownerID.String(), sessionID.String(), fileID.String())
}

func buildFinalKey(finalPrefix, purpose string, ownerID, fileID uuid.UUID) string {
	p := strings.Trim(finalPrefix, "/")
	if p != "" {
		p += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s", p, purpose, ownerID.String(), fileID.String())
}
