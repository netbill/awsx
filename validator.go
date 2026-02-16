package awsx

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrorNoContentUploaded   = errors.New("no content uploaded for resource")
	ErrorSizeExceedsMax      = errors.New("uploaded resource size exceeds the maximum allowed size")
	ErrorResolutionIsInvalid = errors.New("uploaded resource resolution is invalid")
	ErrorFormatNotAllowed    = errors.New("uploaded resource format is not allowed")
)

type ImageValidator struct {
	AllowedFormats []string `mapstructure:"allowed_formats" required:"true"`
	MaxWidth       int      `mapstructure:"max_width" required:"true"`
	MinWidth       int      `mapstructure:"min_width" required:"true"`
	MaxHeight      int      `mapstructure:"max_height" required:"true"`
	MinHeight      int      `mapstructure:"min_height" required:"true"`
	ContentSizeMax int64    `mapstructure:"content_size_max" required:"true"`
}

func (v ImageValidator) String() string {
	return fmt.Sprintf(
		"allowed=%v width=%d..%d height=%d..%d size_max=%d",
		v.AllowedFormats,
		v.MinWidth, v.MaxWidth,
		v.MinHeight, v.MaxHeight,
		v.ContentSizeMax,
	)
}

func (v ImageValidator) Validate(out *s3.GetObjectOutput) error {
	totalSize, err := TotalSizeFromGetObjectOutput(out)
	if err != nil {
		return fmt.Errorf("determine resource size: %w", err)
	}

	if totalSize == 0 {
		return fmt.Errorf("%w", ErrorNoContentUploaded)
	}

	if totalSize > v.ContentSizeMax {
		return fmt.Errorf("%w: got=%d max=%d", ErrorSizeExceedsMax, totalSize, v.ContentSizeMax)
	}

	probe, err := io.ReadAll(out.Body)
	if err != nil {
		return fmt.Errorf("read resource bytes: %w", err)
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(probe))
	if err != nil {
		return fmt.Errorf("decode image config: %w", err)
	}

	if cfg.Width < v.MinWidth || cfg.Width > v.MaxWidth || cfg.Height < v.MinHeight || cfg.Height > v.MaxHeight {
		return fmt.Errorf(
			"%w: got=%dx%d allowed=%dx%d..%dx%d",
			ErrorResolutionIsInvalid,
			cfg.Width, cfg.Height,
			v.MinWidth, v.MinHeight,
			v.MaxWidth, v.MaxHeight,
		)
	}

	if !contains(v.AllowedFormats, format) {
		return fmt.Errorf("%w: got=%s allowed=%v", ErrorFormatNotAllowed, format, v.AllowedFormats)
	}

	return nil
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func TotalSizeFromGetObjectOutput(out *s3.GetObjectOutput) (int64, error) {
	if out == nil {
		return 0, fmt.Errorf("nil GetObjectOutput")
	}
	if out.ContentRange == nil {
		return 0, fmt.Errorf("content-range is nil")
	}
	return TotalSizeFromContentRange(*out.ContentRange)
}

func TotalSizeFromContentRange(cr string) (int64, error) {
	slash := strings.LastIndex(cr, "/")
	if slash < 0 || slash+1 >= len(cr) {
		return 0, fmt.Errorf("invalid content-range: %s", cr)
	}
	return strconv.ParseInt(cr[slash+1:], 10, 64)
}
