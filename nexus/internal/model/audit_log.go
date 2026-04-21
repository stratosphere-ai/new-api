package model

import "gorm.io/gorm"

// AuditLog is a per-request record of prompt/response hashes and safety
// verdicts. Heavier than the existing Log table; retention is governed by
// CompliancePolicy.RetentionDays.
type AuditLog struct {
	gorm.Model
	OrgID          uint   `gorm:"index"`
	UserID         uint   `gorm:"index"`
	TokenID        uint   `gorm:"index"`
	RequestID      string `gorm:"type:varchar(64);index"`
	Endpoint       string `gorm:"type:varchar(128)"`
	ModelName      string `gorm:"column:model;type:varchar(128);index"`
	PromptHash     string `gorm:"type:varchar(128);index"`
	PromptRedacted string `gorm:"type:text"`
	ResponseHash   string `gorm:"type:varchar(128);index"`
	SafetyVerdict  string `gorm:"type:varchar(32)"`
	LatencyMs      int64
}

// CompliancePolicy is the per-org compliance configuration.
type CompliancePolicy struct {
	gorm.Model
	OrgID         uint   `gorm:"uniqueIndex"`
	Region        string `gorm:"type:varchar(32)"`
	PIIRules      string `gorm:"type:text"` // JSON
	BannedModels  string `gorm:"type:text"` // JSON list
	RetentionDays int    `gorm:"default:90"`
}
