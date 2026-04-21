package web3

import (
	"context"
	"time"
)

// JWTIssuer mints and validates JWTs used by the Web3 sign-in flow.
// Separate from API token auth (which uses opaque sk-* tokens).
type JWTIssuer interface {
	Issue(ctx context.Context, id *Web3Identity, ttl time.Duration) (string, error)
	Verify(ctx context.Context, token string) (*Web3Identity, error)
}
