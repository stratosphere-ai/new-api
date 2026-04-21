package web3

import "context"

// SolanaAuthenticator verifies ed25519 signatures over a canonical message
// (base58-encoded address + nonce).
//
// TODO: implement using ed25519.Verify.
type SolanaAuthenticator struct{}

func (a *SolanaAuthenticator) IssueNonce(ctx context.Context, addr string, chain Chain) (Nonce, error) {
	_ = ctx
	_ = addr
	_ = chain
	return Nonce{}, nil
}

func (a *SolanaAuthenticator) Verify(ctx context.Context, msg SignedMessage) (*Web3Identity, error) {
	_ = ctx
	_ = msg
	return nil, nil
}
