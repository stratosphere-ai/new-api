package x402

import "context"

// FacilitatorClient talks to an x402 facilitator service that can verify
// payment intents and optionally execute settlement on behalf of the server.
//
// TODO: implement HTTP client against the chosen facilitator (e.g. Coinbase's
// public facilitator or a self-hosted one).
type FacilitatorClient interface {
	Verify(ctx context.Context, intent *PaymentIntent, requirements *PaymentRequirements) error
	Settle(ctx context.Context, intent *PaymentIntent) (txHash string, err error)
}
