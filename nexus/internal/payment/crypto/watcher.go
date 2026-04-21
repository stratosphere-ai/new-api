package crypto

import "context"

// Watcher subscribes to on-chain transfer events for the configured deposit
// addresses and produces CryptoPayment rows upon confirmation.
//
// TODO: implement per-chain pollers or websocket subscribers:
//   - Ethereum / BSC: eth_subscribe("logs") for USDT/USDC Transfer topic.
//   - Tron: tronweb event subscribe.
//   - Solana: getSignaturesForAddress polling (SPL token program).
type Watcher interface {
	Run(ctx context.Context) error
}
