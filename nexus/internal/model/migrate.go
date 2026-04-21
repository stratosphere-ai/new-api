package model

import "gorm.io/gorm"

// AutoMigrate runs GORM migrations for all nexus-api entities.
//
// TODO: extend with ported User/Token/Channel/Log/Option/... models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Organization{}, &OrgMember{}, &Role{},
		&Web3Credential{},
		&CryptoPayment{}, &X402Payment{}, &SettlementTx{},
		&RoutingPolicy{}, &RoutingStat{},
		&AuditLog{}, &CompliancePolicy{},
		&AgentDefinition{}, &AgentRun{}, &ToolBinding{}, &McpServer{},
		&KnowledgeBase{}, &KBDocument{}, &KBChunk{},
	)
}
