// Package x402 implements the Coinbase x402 HTTP 402 Payment Required protocol
// for per-request micro-settlement of AI API calls.
//
// Spec (draft): https://www.x402.org/
//
// Flow summary:
//  1. Client -> GET/POST resource without X-PAYMENT.
//  2. Server -> 402 Payment Required + X-PAYMENT-REQUIRED header whose value
//     is a base64url JSON describing asset, amount, recipient, chain, and a
//     facilitator URL.
//  3. Client constructs a signed payment intent (EIP-712 or similar) and
//     retries with X-PAYMENT: <base64url intent>.
//  4. Server Verifier decodes + verifies with on-chain or facilitator check.
//  5. On response, Settler finalizes (may batch on-chain).
package x402

import (
	"context"
	"errors"
	"net/http"

	"github.com/stratosphere-ai/nexus-api/internal/payment"
)

// PaymentRequirements encodes the X-PAYMENT-REQUIRED payload.
type PaymentRequirements struct {
	Version     string   `json:"version"`
	Recipient   string   `json:"recipient"`
	Amount      string   `json:"amount"` // decimal string
	AmountUSD   float64  `json:"amountUSD"`
	Asset       string   `json:"asset"` // contract address or symbol
	Chain       string   `json:"chain"`
	Nonce       string   `json:"nonce"`
	ExpiresAt   int64    `json:"expiresAt"`
	Facilitator string   `json:"facilitator"`
	Accepts     []string `json:"accepts"` // e.g. ["eip712","eip3009"]
}

// PaymentIntent is the decoded X-PAYMENT payload.
type PaymentIntent struct {
	Version   string `json:"version"`
	Asset     string `json:"asset"`
	Chain     string `json:"chain"`
	From      string `json:"from"`
	Amount    string `json:"amount"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Raw       []byte `json:"-"`
}

// Verifier decodes and verifies X-PAYMENT headers.
type Verifier interface {
	Decode(h http.Header) (*PaymentIntent, error)
	Verify(ctx context.Context, intent *PaymentIntent, requirements *PaymentRequirements) error
}

// RequirementsBuilder produces the 402 response headers + JSON body for a
// given (org, model, estUSD). Organizations may override recipient wallet,
// chain preference, or facilitator URL in their settings.
type RequirementsBuilder interface {
	Build(ctx context.Context, orgID uint, model string, estUSD float64) (header http.Header, body []byte, err error)
}

// Settler finalizes a verified PaymentIntent into a Receipt. Implementations
// may batch multiple intents into one on-chain transaction.
type Settler interface {
	payment.PaymentSettler
}

// ErrMissingPayment is returned when X-PAYMENT is absent.
var ErrMissingPayment = errors.New("x402: missing X-PAYMENT header")

// ErrInvalidPayment covers any verification failure (bad sig, wrong amount,
// expired nonce, replayed intent).
var ErrInvalidPayment = errors.New("x402: invalid payment intent")
