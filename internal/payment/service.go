package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
	"strings"
	"time"
)

const (
	METHOD_POST          = "POST"
	METHOD_GET           = "GET"
	METHOD_PUT           = "PUT"
	CONTENT_TYPE         = "Content-Type"
	CONTENT_TYPE_JSON    = "application/json"
	AUTHORIZATION        = "Authorization"
	AUTHORIZATION_BEARER = "Bearer"
	SUCCESS_STATUS       = "success"
)

type PaymentService interface {
	InitializePayment(ctx context.Context, payload PaymentPayload) (*common.InitializationResponse, error)
	VerifySignature(body []byte, chapaSignature, xChapaSignature string) bool
	VerifyTransaction(ctx context.Context, transactionID string) (*VerificationResult, error)
	CancelTransaction(ctx context.Context, transactionID string) error
}

type paymentService struct {
	secretKey     string
	initialURL    string
	verifyURL     string
	webhookSecret string
	client        *http.Client
	logger        config.Logger
}

func NewPaymentService(secretKey, initialURL, verifyURL, webhookSecret string, logger config.Logger) PaymentService {
	return &paymentService{
		secretKey:     secretKey,
		initialURL:    initialURL,
		verifyURL:     verifyURL,
		webhookSecret: webhookSecret,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

func (s *paymentService) InitializePayment(ctx context.Context, payload PaymentPayload) (*common.InitializationResponse, error) {
	request, err := s.buildRequest(ctx, payload)
	if err != nil {
		s.logger.Error("failed to build request", "error", err)
		return nil, common.ErrInternalServerError
	}

	response, err := s.client.Do(request)
	if err != nil {
		s.logger.Error("failed to send request", "error", err)
		return nil, common.ErrInternalServerError
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("failed to read response body", "error", err)
		return nil, common.ErrInternalServerError
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		s.logger.Error("chapa initialize request failed", "status", response.StatusCode, "body", string(body))
		return nil, common.ErrInternalServerError
	}

	var responseBody initializeAPIResponse
	if err := json.Unmarshal(body, &responseBody); err != nil {
		s.logger.Error("failed to unmarshal response body", "error", err)
		return nil, common.ErrInternalServerError
	}

	if responseBody.Status != SUCCESS_STATUS || responseBody.Data.CheckoutURL == "" {
		s.logger.Error("chapa initialize returned unsuccessful response", "status", responseBody.Status, "message", responseBody.Message)
		return nil, common.ErrInternalServerError
	}

	return &common.InitializationResponse{
		CheckoutURL: responseBody.Data.CheckoutURL,
	}, nil
}

func (s *paymentService) buildRequest(ctx context.Context, payload PaymentPayload) (*http.Request, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal payload", "error", err)
		return nil, common.ErrInternalServerError
	}
	return s.makeRequest(ctx, METHOD_POST, s.initialURL, payloadBytes)
}

func (s *paymentService) makeRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = strings.NewReader(string(body))
	}

	request, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		s.logger.Error("failed to create request", "error", err)
		return nil, common.ErrInternalServerError
	}

	if method == METHOD_POST {
		request.Header.Set(CONTENT_TYPE, CONTENT_TYPE_JSON)
	}
	request.Header.Set(AUTHORIZATION, AUTHORIZATION_BEARER+" "+s.secretKey)
	return request, nil
}

func (s *paymentService) VerifySignature(body []byte, chapaSignature, xChapaSignature string) bool {
	if xChapaSignature != "" && s.matchesPayloadSignature(body, xChapaSignature) {
		return true
	}
	if chapaSignature != "" && s.matchesSecretSignature(chapaSignature) {
		return true
	}
	return false
}

func (s *paymentService) matchesPayloadSignature(body []byte, signature string) bool {
	hash := hmac.New(sha256.New, []byte(s.webhookSecret))
	hash.Write(body)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func (s *paymentService) matchesSecretSignature(signature string) bool {
	hash := hmac.New(sha256.New, []byte(s.webhookSecret))
	hash.Write([]byte(s.webhookSecret))
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func (s *paymentService) VerifyTransaction(ctx context.Context, transactionID string) (*VerificationResult, error) {
	url := s.verifyURL + "/" + transactionID
	request, err := s.makeRequest(ctx, METHOD_GET, url, nil)
	if err != nil {
		s.logger.Error("failed to create request", "error", err)
		return nil, common.ErrInternalServerError
	}

	response, err := s.client.Do(request)
	if err != nil {
		s.logger.Error("failed to send request", "error", err)
		return nil, common.ErrInternalServerError
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("failed to read response body", "error", err)
		return nil, common.ErrInternalServerError
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		s.logger.Error("chapa verify request failed", "status", response.StatusCode, "body", string(body))
		return nil, common.ErrInternalServerError
	}

	var responseBody VerificationResponse
	if err := json.Unmarshal(body, &responseBody); err != nil {
		s.logger.Error("failed to unmarshal response body", "error", err)
		return nil, common.ErrInternalServerError
	}

	if responseBody.Status != SUCCESS_STATUS || responseBody.Data.Status != SUCCESS_STATUS {
		s.logger.Error("failed to verify transaction", "status", responseBody.Status, "data_status", responseBody.Data.Status)
		return nil, common.ErrInternalServerError
	}

	return &VerificationResult{
		Status:   responseBody.Data.Status,
		Amount:   responseBody.Data.Amount.String(),
		Currency: responseBody.Data.Currency,
		TxRef:    responseBody.Data.TxRef,
	}, nil
}

func (s *paymentService) CancelTransaction(ctx context.Context, transactionID string) error {
	url := strings.Replace(s.verifyURL, "/verify", "/cancel", 1) + "/" + transactionID

	request, err := s.makeRequest(ctx, METHOD_PUT, url, nil)
	if err != nil {
		s.logger.Error("failed to create cancel request", "error", err)
		return common.ErrInternalServerError
	}

	response, err := s.client.Do(request)
	if err != nil {
		s.logger.Error("failed to send cancel request", "error", err)
		return common.ErrInternalServerError
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		s.logger.Error("failed to read cancel response body", "error", err)
		return common.ErrInternalServerError
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		s.logger.Warn("chapa cancel request returned non-success", "status", response.StatusCode, "body", string(body))
		return nil
	}

	return nil
}
