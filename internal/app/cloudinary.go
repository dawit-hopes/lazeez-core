package app

import (
	"fmt"
	"lazeez-core/config"
	"lazeez-core/internal/common"

	"github.com/cloudinary/cloudinary-go/v2"
)

func initCloudinary(logger config.Logger) (*cloudinary.Cloudinary, error) {
	cloudName := getEnv("CLOUDINARY_CLOUD_NAME", "")
	apiKey := getEnv("CLOUDINARY_API_KEY", "")
	apiSecret := getEnv("CLOUDINARY_API_SECRET", "")
	fmt.Println("Cloudinary Configuration", cloudName, apiKey, apiSecret)

	url := fmt.Sprintf("cloudinary://%s:%s@%s", apiKey, apiSecret, cloudName)

	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		logger.Error("Failed to initialize cloudinary", "error", err)
		return nil, common.ErrInternalServerError
	}
	return cld, nil
}
