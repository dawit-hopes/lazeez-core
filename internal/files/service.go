package files

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type FileService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error)
	UploadFileToFolder(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	GetFile(ctx context.Context, id string) (string, error)
	DeleteFile(ctx context.Context, id string) error
}

type fileService struct {
	logger config.Logger
	cld    *cloudinary.Cloudinary
}

func NewFileService(logger config.Logger, cld *cloudinary.Cloudinary) FileService {
	return &fileService{logger: logger, cld: cld}
}

func (s *fileService) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Open the multipart file
	s.logger.Info("Uploading file", "file", file.Filename)

	fileName := s.generateFileName()
	src, err := file.Open()
	if err != nil {
		s.logger.Error("Failed to open file", err)
		return "", common.ErrInternalServerError
	}
	defer src.Close()

	// Upload to Cloudinary
	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:   "lazeez_menu",
		PublicID: fileName,
	})

	if err != nil {
		s.logger.Error("Cloudinary upload failed", err)
		return "", common.ErrInternalServerError
	}

	if uploadResult == nil || uploadResult.SecureURL == "" {
		s.logger.Error("Upload result is nil or secure URL is empty")
		return "", common.ErrInternalServerError
	}

	s.logger.Info("Uploaded file", "file", file.Filename, "url", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

func (s *fileService) UploadFileToFolder(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	s.logger.Info("Uploading file", "file", file.Filename, "folder", folder)

	fileName := s.generateFileName()
	src, err := file.Open()
	if err != nil {
		s.logger.Error("Failed to open file", err)
		return "", common.ErrInternalServerError
	}
	defer src.Close()

	uploadFolder := folder
	if uploadFolder == "" {
		uploadFolder = "lazeez_menu"
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:   uploadFolder,
		PublicID: fileName,
	})

	if err != nil {
		s.logger.Error("Cloudinary upload failed", err)
		return "", common.ErrInternalServerError
	}

	if uploadResult == nil || uploadResult.SecureURL == "" {
		s.logger.Error("Upload result is nil or secure URL is empty")
		return "", common.ErrInternalServerError
	}

	s.logger.Info("Uploaded file", "file", file.Filename, "url", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

func (s *fileService) GetFile(ctx context.Context, id string) (string, error) {
	// Generate a URL for the given Public ID
	img, err := s.cld.Image(id)
	if err != nil {
		s.logger.Error("Failed to generate image URL", err)
		return "", common.ErrInternalServerError
	}

	url, err := img.String()
	if err != nil {
		s.logger.Error("Failed to generate image URL", err)
		return "", common.ErrInternalServerError
	}
	return url, nil
}

func (s *fileService) DeleteFile(ctx context.Context, id string) error {
	// Delete the file from Cloudinary
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: id,
	})
	if err != nil {
		s.logger.Error("Failed to delete file", err)
		return common.ErrInternalServerError
	}
	return err
}

func (s *fileService) generateFileName() string {
	return uuid.New().String()
}
