// Package policy defines the RoutingPolicy abstraction used by
// middleware.Distribute to select a channel for a relay request.
//
// The default implementation "weighted" ports new-api's
// service.CacheGetRandomSatisfiedChannel; additional policies add
// cost-awareness, latency-awareness, quality-awareness, A/B, canary,
// and shadow traffic.
package policy

import (
	"context"
	"errors"
	"time"
)

// Channel is the minimal view of a selectable relay channel. The concrete
// GORM type lives in internal/model; kept here as a narrow interface to
// avoid an import cycle (policy depends on nothing heavy).
type Channel struct {
	ID       uint
	Type     int
	Weight   uint
	Priority int
	BaseURL  string
	Tags     []string
}

// RoutingHint carries request context that policies use to pick a channel.
type RoutingHint struct {
	OrgID       uint
	UserID      uint
	TokenID     uint
	Model       string
	Group       string
	EstPromptTk int
	ClientIP    string
}

// RoutingResult reports the outcome of a relay for RoutingPolicy.Observe.
// Feeds the stats tables that cost/latency/quality policies read from.
type RoutingResult struct {
	ChannelID uint
	Model     string
	Success   bool
	LatencyMs int64
	CostUSD   float64
	At        time.Time
}

// RoutingPolicy picks a channel out of the candidate pool.
type RoutingPolicy interface {
	Name() string
	Pick(ctx context.Context, candidates []*Channel, hint RoutingHint) (*Channel, error)
	Observe(ctx context.Context, result RoutingResult)
}

// ErrNoCandidate is returned when the candidate list is empty or none
// satisfy the policy (e.g. all failing in canary mode).
var ErrNoCandidate = errors.New("policy: no candidate channel")
