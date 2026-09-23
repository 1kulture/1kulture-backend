package idgen

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// OrderReference generates a short, unique order reference.
// Example: ORD_20260923_ABC4X9KZ
func OrderReference() string {
	return fmt.Sprintf("ORD_%s_%s", time.Now().UTC().Format("20060102"), randomStr(8))
}

// PaymentReference generates a unique payment reference for the provider.
// Example: PAY_20260923_XY7K3M2Q
func PaymentReference() string {
	return fmt.Sprintf("PAY_%s_%s", time.Now().UTC().Format("20060102"), randomStr(8))
}

// RefundReference generates a unique reference for a refund.
func RefundReference() string {
	return fmt.Sprintf("REF_%s_%s", time.Now().UTC().Format("20060102"), randomStr(8))
}

// TicketCode generates a unique ticket code (used for QR display/scanning).
// Example: TKT_4KZ9XJ7P2M6Q
func TicketCode() string {
	return "TKT_" + randomStr(12)
}

// TransferToken generates a secure token for ticket transfer links.
func TransferToken() string {
	return randomStr(32)
}

func randomStr(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	// base32 output is longer than n; trim to n chars and remove ambiguous letters
	enc = strings.NewReplacer("0", "A", "O", "B", "I", "C", "L", "D").Replace(enc)
	if len(enc) >= n {
		return enc[:n]
	}
	return enc
}
