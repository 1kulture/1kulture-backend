package factory

import (
	"fmt"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/payments"
	"github.com/1kulture/1kulture-backend/internal/payments/paystack"
)

// NewProvider returns a PaymentProvider based on config.
// Additional providers (flutterwave, stripe) plug in here.
func NewProvider(cfg *config.Config) (payments.PaymentProvider, error) {
	switch cfg.Payments.DefaultProvider {
	case "", "paystack":
		if cfg.Payments.PaystackSecretKey == "" {
			return nil, fmt.Errorf("paystack secret key not configured")
		}
		return paystack.NewClient(
			cfg.Payments.PaystackSecretKey,
			cfg.Payments.PaystackWebhookSecret,
		), nil
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", cfg.Payments.DefaultProvider)
	}
}
