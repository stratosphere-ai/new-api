package web3

import "context"

// SIWEAuthenticator implements EIP-4361 (Sign-In with Ethereum).
//
// TODO: implement message construction, secp256k1 recovery, address match,
// and nonce lookup/invalidate against Redis-backed NonceStore.
type SIWEAuthenticator struct{}

func (a *SIWEAuthenticator) IssueNonce(ctx context.Context, addr string, chain Chain) (Nonce, error) {
	_ = ctx
	_ = addr
	_ = chain
	return Nonce{}, nil
}

func (a *SIWEAuthenticator) Verify(ctx context.Context, msg SignedMessage) (*Web3Identity, error) {
	_ = ctx
	_ = msg
	return nil, nil
}
