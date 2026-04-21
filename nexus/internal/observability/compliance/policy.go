package compliance

import "context"

// Policy defines org-specific compliance rules evaluated by middleware.
type Policy struct {
	OrgID         uint
	Region        string
	PIIRules      []string
	BannedModels  []string
	RetentionDays int
}

// Checker loads the policy for an org and verifies a request against it.
type Checker interface {
	Load(ctx context.Context, orgID uint) (*Policy, error)
	Check(ctx context.Context, p *Policy, model string) error
}
