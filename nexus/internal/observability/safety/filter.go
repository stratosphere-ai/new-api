// Package safety provides a content-safety hook interface executed by the
// ContentSafetyPre/Post middlewares. Plug in any moderation service
// (OpenAI moderation, self-hosted classifier, regex rules).
package safety

import "context"

// Verdict is the outcome of a safety check.
type Verdict struct {
	Action  string // "allow" | "flag" | "reject"
	Reasons []string
	Score   float64
}

// Filter is invoked pre- and post-response by middleware.
type Filter interface {
	PreCheck(ctx context.Context, prompt string) (Verdict, error)
	PostCheck(ctx context.Context, response string) (Verdict, error)
}
