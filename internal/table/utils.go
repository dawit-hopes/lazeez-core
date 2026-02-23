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

	"github.com/skip2/go-qrcode"
)

func GenerateTableRef() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateQRCodeHeader(reference string) (*multipart.FileHeader, error) {
	BASE_URL := os.Getenv("LAZEEZ_TABLE_BASE_URL")
	url := fmt.Sprintf("%s%s", BASE_URL, reference)

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
