package awsx

import (
	"errors"
	"fmt"

	"github.com/aws/smithy-go"
)

var (
	ErrNotFound     = errors.New("awsx: object not found")
	ErrAccessDenied = errors.New("awsx: access denied")
)

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey":
			return fmt.Errorf("%w: %v", ErrNotFound, err)
		case "AccessDenied":
			return fmt.Errorf("%w: %v", ErrAccessDenied, err)
		}
	}

	return err
}
