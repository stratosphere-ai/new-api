package model

import (
	"time"

	"gorm.io/gorm"
)

// RoutingPolicy is a named, org-owned routing configuration. A zero OrgID
// means the policy is a global default template.
type RoutingPolicy struct {
	gorm.Model
	OrgID  uint   `gorm:"index"`
	Name   string `gorm:"type:varchar(64);index"`
	Kind   string `gorm:"type:varchar(32)"` // weighted | cost | latency | quality | ab | canary | shadow
	Rules  string `gorm:"type:text"`        // JSON
	Active bool   `gorm:"default:true"`
}

// RoutingStat is a rolling window aggregation used by cost/latency/quality
// policies. Keyed by (OrgID, ChannelID, Model, WindowStart).
type RoutingStat struct {
	ID           uint `gorm:"primaryKey"`
	OrgID        uint `gorm:"index:ix_stat,priority:1"`
	ChannelID    uint `gorm:"index:ix_stat,priority:2"`
	ModelName    string `gorm:"column:model;type:varchar(64);index:ix_stat,priority:3"`
	WindowStart  time.Time `gorm:"index:ix_stat,priority:4"`
	Requests     int64
	Errors       int64
	P95LatencyMs int64
	AvgCostCents int64
}
