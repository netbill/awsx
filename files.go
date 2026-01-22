package imgx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/imgx/data/sqlc"
)

type CreateFileIntentsInput struct {
	SessionID   uuid.UUID
	OwnerID     uuid.UUID
	FilesCount  int
	KeyPrefix   string
	ContentType string
}

type FileIntent struct {
	FileID    uuid.UUID
	ObjectKey string
	PutURL    string
}

func (s Service) CreateFileIntents(ctx context.Context, in CreateFileIntentsInput) ([]FileIntent, error) {
	if in.FilesCount <= 0 {
		return nil, errors.New("files_count must be > 0")
	}
	if strings.TrimSpace(in.ContentType) == "" {
		return nil, errors.New("content_type is required")
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

	existing, err := q.ListUploadFilesBySession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if int32(len(existing))+int32(in.FilesCount) > sess.MaxFiles {
		return nil, fmt.Errorf("too many files: max=%d", sess.MaxFiles)
	}

	files := make(map[uuid.UUID]string)

	if err = s.DB.Transaction(ctx, func(txCtx context.Context) error {
		for i := 0; i < in.FilesCount; i++ {
			fileID := uuid.New()
			key := buildObjectKey(in.KeyPrefix, sess.Purpose, sess.OwnerID, in.SessionID, fileID)

			_, err = q.CreateUploadFile(ctx, sqlc.CreateUploadFileParams{
				ID:        fileID,
				SessionID: in.SessionID,
				ObjectKey: key,
			})
			if err != nil {
				return err
			}

			files[fileID] = key
		}

		return nil
	}); err != nil {
		return nil, err
	}

	intents := make([]FileIntent, 0, in.FilesCount)

	for fileID, key := range files {
		putURL, err := s.Signer.PresignPut(ctx, key, in.ContentType)
		if err != nil {
			return nil, err
		}

		intents = append(intents, FileIntent{
			FileID:    fileID,
			ObjectKey: buildObjectKey(in.KeyPrefix, sess.Purpose, sess.OwnerID, in.SessionID, fileID),
			PutURL:    putURL,
		})
	}

	return intents, nil
}
