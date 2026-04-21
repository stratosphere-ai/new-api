package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/stratosphere-ai/nexus-api/internal/config"
	"github.com/stratosphere-ai/nexus-api/internal/router"
)

// main bootstraps the nexus-api server.
//
// Mirrors /home/user/new-api/main.go:142-182 but adds pre-router init for:
//   - OpenTelemetry tracer/meter
//   - Crypto wallet clients (eth/sol/tron/bsc)
//   - MCP server registry
//
// TODO: port env/flag parsing, DB init (GORM), Redis, i18n, session store.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// TODO: initialize OTel, DB, Redis, wallet clients, MCP registry here.

	engine := gin.New()
	engine.Use(gin.Recovery())
	router.SetRouter(engine)

	addr := cfg.ListenAddr
	log.Printf("nexus-api listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
