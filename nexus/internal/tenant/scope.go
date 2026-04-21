package tenant

import "gorm.io/gorm"

// WithOrgScope returns a GORM scope function that narrows any query to a
// given org_id. Prefer this over hand-written `.Where("org_id = ?")` so
// the scope rule stays in one place.
//
//	db.Scopes(tenant.WithOrgScope(orgID)).Find(&channels)
//
// TODO: validate that the target table actually has an org_id column (some
// global settings tables should skip scoping).
func WithOrgScope(orgID uint) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if orgID == 0 {
			return tx
		}
		return tx.Where("org_id = ?", orgID)
	}
}
