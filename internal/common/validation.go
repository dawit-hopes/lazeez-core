package common

import (
	"mime/multipart"
	"slices"
)

const MAX_IMAGE_SIZE = 1024 * 1024 * 15

var allowedImageTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

func validateImageType(image multipart.FileHeader) error {
	if !slices.Contains(allowedImageTypes, image.Header.Get("Content-Type")) {
		return ErrInvalidImageType
	}
	return nil
}

func ValidateImage(image multipart.FileHeader) error {
	if image.Size > MAX_IMAGE_SIZE {
		return ErrImageSizeTooLarge
	}

	if err := validateImageType(image); err != nil {
		return err
	}

	return nil
}
