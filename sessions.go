package imgx

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/netbill/imgx/data/sqlc"
)

type CreateSessionInput struct {
	OwnerID  uuid.UUID
	Purpose  string
	TTL      time.Duration
	MaxFiles int32
}

func (s Service) CreateSession(ctx context.Context, in CreateSessionInput) (uuid.UUID, error) {
	if in.MaxFiles <= 0 {
		return uuid.Nil, errors.New("max_files must be > 0")
	}
	if in.TTL <= 0 {
		return uuid.Nil, errors.New("ttl must be > 0")
	}
	if strings.TrimSpace(in.Purpose) == "" {
		return uuid.Nil, errors.New("purpose is required")
	}

	id := uuid.New()
	expiresAt := time.Now().UTC().Add(in.TTL)

	q := s.DB.Queries(ctx)

	_, err := q.CreateUploadSession(ctx, sqlc.CreateUploadSessionParams{
		ID:        id,
		OwnerID:   in.OwnerID,
		Purpose:   in.Purpose,
		ExpiresAt: expiresAt,
		MaxFiles:  in.MaxFiles,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

type PresignGetInput struct {
	ObjectKey string
}

func (s Service) PresignGet(ctx context.Context, in PresignGetInput) (string, error) {
	if strings.TrimSpace(in.ObjectKey) == "" {
		return "", errors.New("object_key is required")
	}
	return s.Signer.PresignGet(ctx, in.ObjectKey)
}

type AcceptSessionInput struct {
	SessionID   uuid.UUID
	OwnerID     uuid.UUID
	FinalPrefix string
}

type AcceptedFile struct {
	FileID     uuid.UUID
	StagingKey string
	FinalKey   string
	Head       *s3.HeadObjectOutput
}

// AcceptSession finalizes the upload session by moving files to their final location.
func (s Service) AcceptSession(ctx context.Context, in AcceptSessionInput) ([]AcceptedFile, error) {
	if strings.TrimSpace(in.FinalPrefix) == "" {
		return nil, errors.New("final_prefix is required")
	}

	q := s.DB.Queries(ctx)

	sess, err := q.GetUploadSessionByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if sess.OwnerID != in.OwnerID {
		return nil, ErrForbidden
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		return nil, ErrExpired
	}

	files, err := q.ListUploadFilesBySession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrNoFiles
	}

	accepted := make([]AcceptedFile, 0, len(files))

	if err = q.DeleteUploadSession(ctx, in.SessionID); err != nil {
		return nil, err
	}

	for _, f := range files {
		head, err := s.S3.Head(ctx, f.ObjectKey)
		if err != nil {
			return nil, ErrFileMissing
		}

		finalKey := buildFinalKey(in.FinalPrefix, sess.Purpose, sess.OwnerID, f.ID)

		if err = s.S3.Copy(ctx, f.ObjectKey, finalKey); err != nil {
			return nil, err
		}
		if err = s.S3.Delete(ctx, f.ObjectKey); err != nil {
			return nil, err
		}

		accepted = append(accepted, AcceptedFile{
			FileID:     f.ID,
			StagingKey: f.ObjectKey,
			FinalKey:   finalKey,
			Head:       head,
		})
	}

	return accepted, nil
}

type CancelSessionInput struct {
	SessionID uuid.UUID
	OwnerID   uuid.UUID
}

// CancelSession deletes the upload session and all associated files from storage.
func (s Service) CancelSession(ctx context.Context, input CancelSessionInput) error {
	q := s.DB.Queries(ctx)

	sess, err := q.GetUploadSessionByID(ctx, input.SessionID)
	if err != nil {
		return err
	}
	if sess.OwnerID != input.OwnerID {
		return ErrForbidden
	}

	files, err := q.ListUploadFilesBySession(ctx, input.SessionID)
	if err != nil {
		return err
	}

	if err = q.DeleteUploadSession(ctx, input.SessionID); err != nil {
		return err
	}

	for _, f := range files {
		if err = s.S3.Delete(ctx, f.ObjectKey); err != nil {
			return err
		}
	}

	return nil
}
