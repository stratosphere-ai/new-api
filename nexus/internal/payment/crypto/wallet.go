package crypto

import "context"

// WalletService allocates deposit addresses for users/orgs. Strategies:
//   - Shared omnibus wallet with per-user memo (Tron "message").
//   - HD wallet derivation: m/44'/60'/account'/0/index for unique EVM addr.
//
// TODO: implement HD derivation for ETH/BSC and omnibus + memo for Tron/Solana.
type WalletService interface {
	DepositAddress(ctx context.Context, userID uint, chain string, asset string) (address string, memo string, err error)
}
