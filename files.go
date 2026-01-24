package imgx

import (
	"fmt"
	"strings"
)

func GetFileExtension(link string) (string, error) {
	parts := strings.Split(link, ".")
	if len(parts) == 0 {
		return "", fmt.Errorf("unexcepted file link")
	}

	return parts[len(parts)-1], nil
}

func CheckExtension(link string, extension ...string) (bool, error) {
	fileExt, err := GetFileExtension(link)
	if err != nil {
		return false, err
	}

	for _, ext := range extension {
		if ext == fileExt {
			return true, nil
		}
	}

	return false, nil
}
