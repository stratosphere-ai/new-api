package web3

import (
	"context"
	"time"
)

// NonceStore persists issued nonces keyed by (chain, address) for a short TTL.
// Expected implementation: Redis with EXPIRE.
type NonceStore interface {
	Put(ctx context.Context, chain Chain, address string, value string, ttl time.Duration) error
	Take(ctx context.Context, chain Chain, address string) (string, error)
}
