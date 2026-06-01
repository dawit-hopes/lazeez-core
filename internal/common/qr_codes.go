package common

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/skip2/go-qrcode"
)

func buildMenuURL(baseURL, reference, serviceType string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch serviceType {
	case "table":
		base = strings.TrimSuffix(base, "/tbl")
	case "room":
		base = strings.TrimSuffix(base, "/rm")
	}
	ref := strings.TrimLeft(strings.TrimSpace(reference), "/")
	return fmt.Sprintf("%s/%s", base, ref)
}

func menuBaseURLFromEnv() string {
	if base := strings.TrimSpace(os.Getenv("LAZEEZ_MENU_BASE_URL")); base != "" {
		return base
	}
	return os.Getenv("LAZEEZ_TABLE_BASE_URL")
}

func GenerateQRCodeHeader(reference string, serviceType string) (*multipart.FileHeader, error) {
	url := buildMenuURL(menuBaseURLFromEnv(), reference, serviceType)

	// Generate QR in memory (no disk needed 🚀)
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create fake file field
	fileName := GenerateUUID()
	part, err := writer.CreateFormFile("file", fmt.Sprintf("%s.png", fileName))
	if err != nil {
		return nil, err
	}

	_, err = part.Write(png)
	if err != nil {
		return nil, err
	}

	writer.Close()

	req := &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(body),
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(int64(body.Len()))
	if err != nil {
		return nil, err
	}

	fileHeader := req.MultipartForm.File["file"][0]

	return fileHeader, nil
}
