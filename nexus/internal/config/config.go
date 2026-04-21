package config

// Config holds runtime configuration loaded from env + yaml.
//
// TODO: port relevant fields from /home/user/new-api/common/env.go and
// /home/user/new-api/setting/system_setting/*.
type Config struct {
	ListenAddr string
	DSN        string
	RedisURL   string

	// Web3 / crypto / x402
	Web3JWTSecret     string
	CryptoWalletMnemonic string
	X402FacilitatorURL string

	// Observability
	OTelEndpoint string
}

// Load parses environment and optional yaml file into Config.
func Load() (*Config, error) {
	// TODO: implement env + yaml layered loader.
	return &Config{ListenAddr: ":3000"}, nil
}
