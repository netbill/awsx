package awsx

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"strings"
)

type ImgObjectValidator struct {
	AllowedContentTypes []string
	AllowedFormats      []string
	MaxWidth            uint
	MaxHeight           uint
	ContentLengthMax    uint
}

func (v *ImgObjectValidator) ValidateImageSize(size uint) (bool, error) {
	if v.ContentLengthMax > 0 && size > v.ContentLengthMax {
		return false, nil
	}

	return true, nil
}

func (v *ImgObjectValidator) ValidateImageResolution(data []byte) (bool, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("decode config: %w", err)
	}

	if v.MaxWidth > 0 && uint(config.Width) > v.MaxWidth {
		return false, nil
	}
	if v.MaxHeight > 0 && uint(config.Height) > v.MaxHeight {
		return false, nil
	}

	return true, nil
}

func (v *ImgObjectValidator) ValidateImageFormat(data []byte) (bool, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("unknown format: %w", err)
	}

	for _, f := range v.AllowedFormats {
		if f == format {
			return true, nil
		}
	}

	return false, nil
}

func (v *ImgObjectValidator) ValidateImageContentType(data []byte) (bool, error) {
	contentType := http.DetectContentType(data)

	for _, t := range v.AllowedContentTypes {
		if strings.Contains(contentType, t) {
			return true, nil
		}
	}

	return false, nil
}
