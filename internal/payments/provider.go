package payments

import (
	"context"
	"errors"
)

var (
	ErrProviderNotConfigured = errors.New("payment provider not configured")
	ErrVerificationFailed    = errors.New("payment verification failed")
	ErrInvalidWebhook        = errors.New("invalid webhook signature")
	ErrProviderUnavailable   = errors.New("payment provider unavailable")
)

// InitializeRequest is the input to a payment init.
type InitializeRequest struct {
	AmountMinor int64  // in smallest currency unit (kobo, cents, etc.)
	Currency    string // ISO 4217 (NGN, USD, etc.)
	Email       string
	Reference   string // our internal reference — must be unique
	CallbackURL string
	Metadata    map[string]interface{}
}

// InitializeResult is the payload returned to the client.
type InitializeResult struct {
	AuthorizationURL string
	AccessCode       string
	Reference        string
}

// VerifyResult is the outcome of checking a transaction's status.
type VerifyResult struct {
	Status      string // success | failed | pending | reversed
	AmountMinor int64
	Currency    string
	Reference   string
	ProviderRef string
	PaidAt      int64 // unix
	Channel     string
	RawResponse []byte // raw JSON for audit
}

// RefundRequest initiates a refund for a transaction.
type RefundRequest struct {
	TransactionReference string
	AmountMinor          int64
	Currency             string
	Reason               string
	// If empty, a refund reference is generated.
	RefundReference string
}

// RefundResult is the outcome of a refund attempt.
type RefundResult struct {
	Status      string // processed | pending | failed
	ProviderRef string
	AmountMinor int64
	RawResponse []byte
}

// WebhookEvent is a normalized representation of a provider webhook.
type WebhookEvent struct {
	EventType   string // charge.success, refund.processed, etc.
	Reference   string
	ProviderRef string
	AmountMinor int64
	Currency    string
	Status      string
	RawPayload  []byte
}

// PaymentProvider is the contract every processor implements.
type PaymentProvider interface {
	Name() string

	Initialize(ctx context.Context, req InitializeRequest) (*InitializeResult, error)
	Verify(ctx context.Context, reference string) (*VerifyResult, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)

	// ParseWebhook validates the signature and normalizes the event.
	// rawBody must be the exact bytes received (before any JSON decoding).
	ParseWebhook(rawBody []byte, signatureHeader string) (*WebhookEvent, error)
}
