package model

import "gorm.io/gorm"

// Organization is the tenant root. A User may belong to many Organizations
// via OrgMember. Channels, Tokens, Logs, etc. are scoped to OrgID.
type Organization struct {
	gorm.Model
	Slug        string `gorm:"type:varchar(64);uniqueIndex"`
	Name        string `gorm:"type:varchar(128)"`
	OwnerUserID uint   `gorm:"index"`
	Plan        string `gorm:"type:varchar(32);default:'free'"`
	Status      int    `gorm:"default:1"`
	Settings    string `gorm:"type:text"` // JSON blob for UI + policy prefs
}

// OrgMember is the join table; composite unique (OrgID, UserID).
type OrgMember struct {
	gorm.Model
	OrgID     uint   `gorm:"uniqueIndex:ux_org_user"`
	UserID    uint   `gorm:"uniqueIndex:ux_org_user;index"`
	RoleID    uint   `gorm:"index"`
	InvitedBy uint
	Status    int `gorm:"default:1"`
}

// Role defines the permission set granted by membership.
// If OrgID is zero the role is a system-wide template (owner/admin/member).
type Role struct {
	gorm.Model
	OrgID       uint   `gorm:"index"`
	Name        string `gorm:"type:varchar(64)"`
	Permissions string `gorm:"type:text"` // JSON bitmask or named list
}
