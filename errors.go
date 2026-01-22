package imgx

import "errors"

var (
	ErrForbidden   = errors.New("forbidden")
	ErrExpired     = errors.New("expired")
	ErrNoFiles     = errors.New("no files")
	ErrFileMissing = errors.New("file missing")
)
