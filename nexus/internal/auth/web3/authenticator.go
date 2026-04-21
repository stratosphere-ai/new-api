package web3

import (
	"context"
	"errors"
	"time"
)

// Chain identifies a supported blockchain for wallet sign-in.
type Chain string

const (
	ChainEthereum Chain = "ethereum"
	ChainSolana   Chain = "solana"
)

// Nonce is a single-use challenge bound to (chain, address).
type Nonce struct {
	Value     string
	Message   string // full SIWE message for ETH; canonical message for SOL
	ExpiresAt time.Time
}

// SignedMessage carries a client signature over Nonce.Message.
type SignedMessage struct {
	Chain     Chain
	Address   string
	Signature string
	Message   string
}

// Web3Identity is the verified wallet identity returned upon successful
// signature verification. Callers issue a JWT from this.
type Web3Identity struct {
	Chain   Chain
	Address string
	UserID  uint // resolved or newly created
}

// Web3Authenticator is the high-level interface consumed by controllers.
// Implementations: siwe.Authenticator (ETH EIP-4361), solana.Authenticator
// (ed25519 over a canonical message).
type Web3Authenticator interface {
	IssueNonce(ctx context.Context, addr string, chain Chain) (Nonce, error)
	Verify(ctx context.Context, msg SignedMessage) (*Web3Identity, error)
}

// ErrInvalidSignature is returned when signature verification fails.
var ErrInvalidSignature = errors.New("web3: invalid signature")

// ErrNonceExpired is returned when the nonce has expired or was never issued.
var ErrNonceExpired = errors.New("web3: nonce expired")
