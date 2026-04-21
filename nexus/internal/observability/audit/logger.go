// Package audit captures prompt/response pairs for compliance and debugging.
// Unlike the existing Log table (usage stats), AuditLog stores redacted
// payload hashes and safety verdicts.
package audit

import (
	"context"
	"time"
)

// Entry is a single audit row.
type Entry struct {
	OrgID          uint
	UserID         uint
	TokenID        uint
	RequestID      string
	Endpoint       string
	Model          string
	PromptHash     string
	PromptRedacted string
	ResponseHash   string
	SafetyVerdict  string
	LatencyMs      int64
	At             time.Time
}

// Logger writes audit entries. Implementations may batch to DB or ship to
// an external SIEM sink.
type Logger interface {
	Write(ctx context.Context, e Entry) error
}

// Redactor strips PII from raw prompt/response text before hashing/storing.
type Redactor interface {
	Redact(text string) string
}
