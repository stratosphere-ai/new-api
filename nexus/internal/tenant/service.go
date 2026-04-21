// Package tenant implements the multi-tenant (Organization) layer.
// Every tenant-scoped row carries an OrgID. Use WithOrgScope(db, orgID) to
// apply a GORM scope that adds "AND org_id = ?" to every query.
package tenant

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Role constants for OrgMember.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Organization is the resolved tenant context attached to requests.
type Organization struct {
	ID   uint
	Slug string
	Name string
}

// Member is the acting user's membership record within the resolved org.
type Member struct {
	OrgID  uint
	UserID uint
	Role   string
}

// OrgResolver is used by middleware.TenantResolver to identify the acting
// organization for a request (from token, header, subdomain, or default).
type OrgResolver interface {
	Resolve(c *gin.Context) (*Organization, *Member, error)
}

// Service exposes CRUD and invite operations for organizations. Backed by
// the GORM models in internal/model.
//
// TODO: implement with DB queries.
type Service interface {
	Create(ctx context.Context, ownerUserID uint, name, slug string) (*Organization, error)
	ListFor(ctx context.Context, userID uint) ([]*Organization, error)
	Invite(ctx context.Context, orgID, inviterUserID, inviteeUserID uint, role string) error
	SetRole(ctx context.Context, orgID, userID uint, role string) error
	Remove(ctx context.Context, orgID, userID uint) error
}
