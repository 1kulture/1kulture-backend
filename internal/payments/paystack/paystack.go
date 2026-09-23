package paystack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/1kulture/1kulture-backend/internal/payments"
)

const baseURL = "https://api.paystack.co"

type Client struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
}

func NewClient(secretKey, webhookSecret string) *Client {
	if webhookSecret == "" {
		// Paystack uses the secret key as the webhook signing key.
		webhookSecret = secretKey
	}
	return &Client{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) Name() string { return "paystack" }

// ---------------- Initialize ----------------

type paystackInitRequest struct {
	Email       string                 `json:"email"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency,omitempty"`
	Reference   string                 `json:"reference"`
	CallbackURL string                 `json:"callback_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type paystackInitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (c *Client) Initialize(ctx context.Context, req payments.InitializeRequest) (*payments.InitializeResult, error) {
	body := paystackInitRequest{
		Email:       req.Email,
		Amount:      req.AmountMinor,
		Currency:    req.Currency,
		Reference:   req.Reference,
		CallbackURL: req.CallbackURL,
		Metadata:    req.Metadata,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("paystack: marshal init: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/transaction/initialize", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("paystack: build init request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", payments.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paystack: init failed (%d): %s", resp.StatusCode, string(raw))
	}

	var out paystackInitResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paystack: decode init: %w", err)
	}
	if !out.Status {
		return nil, fmt.Errorf("paystack: init rejected: %s", out.Message)
	}

	return &payments.InitializeResult{
		AuthorizationURL: out.Data.AuthorizationURL,
		AccessCode:       out.Data.AccessCode,
		Reference:        out.Data.Reference,
	}, nil
}

// ---------------- Verify ----------------

type paystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Channel   string `json:"channel"`
		PaidAt    string `json:"paid_at"`
		ID        int64  `json:"id"`
	} `json:"data"`
}

func (c *Client) Verify(ctx context.Context, reference string) (*payments.VerifyResult, error) {
	url := fmt.Sprintf("%s/transaction/verify/%s", baseURL, reference)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("paystack: build verify request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", payments.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paystack: verify failed (%d): %s", resp.StatusCode, string(raw))
	}

	var out paystackVerifyResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paystack: decode verify: %w", err)
	}
	if !out.Status {
		return nil, fmt.Errorf("%w: %s", payments.ErrVerificationFailed, out.Message)
	}

	var paidAt int64
	if out.Data.PaidAt != "" {
		if t, err := time.Parse(time.RFC3339, out.Data.PaidAt); err == nil {
			paidAt = t.Unix()
		}
	}

	return &payments.VerifyResult{
		Status:      out.Data.Status,
		AmountMinor: out.Data.Amount,
		Currency:    out.Data.Currency,
		Reference:   out.Data.Reference,
		ProviderRef: fmt.Sprintf("%d", out.Data.ID),
		PaidAt:      paidAt,
		Channel:     out.Data.Channel,
		RawResponse: raw,
	}, nil
}

// ---------------- Refund ----------------

type paystackRefundRequest struct {
	Transaction string `json:"transaction"`
	Amount      int64  `json:"amount,omitempty"`
	Currency    string `json:"currency,omitempty"`
	Reason      string `json:"customer_note,omitempty"`
	Reference   string `json:"merchant_note,omitempty"`
}

type paystackRefundResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"`
	} `json:"data"`
}

func (c *Client) Refund(ctx context.Context, req payments.RefundRequest) (*payments.RefundResult, error) {
	body := paystackRefundRequest{
		Transaction: req.TransactionReference,
		Amount:      req.AmountMinor,
		Currency:    req.Currency,
		Reason:      req.Reason,
		Reference:   req.RefundReference,
	}
	b, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/refund", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("paystack: build refund request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", payments.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paystack: refund failed (%d): %s", resp.StatusCode, string(raw))
	}

	var out paystackRefundResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paystack: decode refund: %w", err)
	}

	return &payments.RefundResult{
		Status:      out.Data.Status,
		ProviderRef: out.Data.Reference,
		AmountMinor: out.Data.Amount,
		RawResponse: raw,
	}, nil
}

// ---------------- Webhook ----------------

type paystackWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID        int64  `json:"id"`
		Reference string `json:"reference"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
	} `json:"data"`
}

func (c *Client) ParseWebhook(rawBody []byte, signatureHeader string) (*payments.WebhookEvent, error) {
	if signatureHeader == "" {
		return nil, payments.ErrInvalidWebhook
	}

	mac := hmac.New(sha512.New, []byte(c.webhookSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signatureHeader)) {
		return nil, payments.ErrInvalidWebhook
	}

	var p paystackWebhookPayload
	if err := json.Unmarshal(rawBody, &p); err != nil {
		return nil, fmt.Errorf("paystack: decode webhook: %w", err)
	}

	return &payments.WebhookEvent{
		EventType:   p.Event,
		Reference:   p.Data.Reference,
		ProviderRef: fmt.Sprintf("%d", p.Data.ID),
		AmountMinor: p.Data.Amount,
		Currency:    p.Data.Currency,
		Status:      p.Data.Status,
		RawPayload:  rawBody,
	}, nil
}
