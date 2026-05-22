package table

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"lazeez-core/internal/common"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/skip2/go-qrcode"
)

func GenerateTableRef() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// buildTableMenuURL encodes the guest menu route (/:reference) into QR codes.
func buildTableMenuURL(baseURL, reference string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	base = strings.TrimSuffix(base, "/table")
	ref := strings.TrimLeft(strings.TrimSpace(reference), "/")
	return fmt.Sprintf("%s/%s", base, ref)
}

func menuBaseURLFromEnv() string {
	if base := strings.TrimSpace(os.Getenv("LAZEEZ_MENU_BASE_URL")); base != "" {
		return base
	}
	return os.Getenv("LAZEEZ_TABLE_BASE_URL")
}

func GenerateQRCodeHeader(reference string) (*multipart.FileHeader, error) {
	url := buildTableMenuURL(menuBaseURLFromEnv(), reference)

	// Generate QR in memory (no disk needed 🚀)
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create fake file field
	fileName := common.GenerateUUID()
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
