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

const (
	menuPathTable = "tbl"
	menuPathRoom  = "rm"
)

// buildMenuURL builds a guest-menu deep link: {base}/tbl/{ref} or {base}/rm/{ref}.
// The base may already end with /tbl, /rm, or /table; those suffixes are normalized away first.
func buildMenuURL(baseURL, reference, serviceType string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	base = strings.TrimSuffix(base, "/"+menuPathTable)
	base = strings.TrimSuffix(base, "/"+menuPathRoom)
	base = strings.TrimSuffix(base, "/table")

	ref := strings.TrimLeft(strings.TrimSpace(reference), "/")
	pathPrefix := menuPathTable
	switch serviceType {
	case "room":
		pathPrefix = menuPathRoom
	case "table":
		// default
	}
	return fmt.Sprintf("%s/%s/%s", base, pathPrefix, ref)
}

// BuildPaymentReturnURL builds the Chapa browser redirect for table orders.
func BuildPaymentReturnURL(menuBaseURL, tableReference, orderID string) string {
	menuURL := buildMenuURL(menuBaseURL, tableReference, "table")
	return fmt.Sprintf("%s/payment-return?orderId=%s", menuURL, orderID)
}

func menuBaseURLFromEnv(serviceType string) string {
	if serviceType == "table" {
		return os.Getenv("LAZEEZ_TABLE_MENU_BASE_URL")
	}
	return os.Getenv("LAZEEZ_ROOM_MENU_BASE_URL")
}

func GenerateQRCodeHeader(reference string, serviceType string) (*multipart.FileHeader, error) {
	url := buildMenuURL(menuBaseURLFromEnv(serviceType), reference, serviceType)

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
