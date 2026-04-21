package main

import "log"

// main runs GORM AutoMigrate + seed for all nexus-api models.
//
// TODO: wire up model.AutoMigrate(db) for every entity declared in
// internal/model, including new multi-tenant tables (Organization,
// OrgMember, Role, Web3Credential, CryptoPayment, X402Payment, ...).
func main() {
	log.Println("nexus-api migrate: TODO")
}
