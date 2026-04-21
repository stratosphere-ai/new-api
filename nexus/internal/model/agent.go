package model

import "gorm.io/gorm"

// AgentDefinition is a reusable agent configuration (prompt + tools + KBs).
type AgentDefinition struct {
	gorm.Model
	OrgID        uint   `gorm:"index"`
	Name         string `gorm:"type:varchar(128);index"`
	SystemPrompt string `gorm:"type:text"`
	Tools        string `gorm:"type:text"` // JSON list of ToolBinding IDs + schemas
	KBIDs        string `gorm:"type:text"` // JSON list of KnowledgeBase IDs
	DefaultModel string `gorm:"type:varchar(128)"`
	Version      int    `gorm:"default:1"`
}

// AgentRun is one execution instance.
type AgentRun struct {
	gorm.Model
	OrgID     uint   `gorm:"index"`
	AgentID   uint   `gorm:"index"`
	UserID    uint   `gorm:"index"`
	InputHash string `gorm:"type:varchar(128);index"`
	Status    string `gorm:"type:varchar(32);index"` // queued | running | success | failed
	Steps     string `gorm:"type:text"`              // JSON of Step events
	CostQuota int64
}

// ToolBinding registers one tool attached to an AgentDefinition.
type ToolBinding struct {
	gorm.Model
	AgentID uint   `gorm:"index"`
	Name    string `gorm:"type:varchar(128)"`
	Kind    string `gorm:"type:varchar(32)"` // http | builtin | mcp
	Config  string `gorm:"type:text"`        // JSON (url, headers, schema, MCP server ref)
}

// McpServer is an external MCP endpoint attached to an org.
type McpServer struct {
	gorm.Model
	OrgID       uint   `gorm:"index"`
	Name        string `gorm:"type:varchar(128)"`
	URL         string `gorm:"type:varchar(512)"`
	AuthType    string `gorm:"type:varchar(32)"` // none | bearer | mTLS
	Credentials string `gorm:"type:text"`        // encrypted blob
}
