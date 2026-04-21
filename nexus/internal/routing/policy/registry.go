package policy

import "sync"

// Registry holds the set of RoutingPolicy implementations by name.
// Orgs reference a policy by Name() in their RoutingPolicy DB rows.
type Registry struct {
	mu sync.RWMutex
	m  map[string]RoutingPolicy
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{m: make(map[string]RoutingPolicy)} }

// Register installs p keyed by p.Name().
func (r *Registry) Register(p RoutingPolicy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[p.Name()] = p
}

// Get returns the policy by name (or nil).
func (r *Registry) Get(name string) RoutingPolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.m[name]
}

// TODO: register default policies (weighted, cost, latency, quality, ab,
// canary, shadow) during server bootstrap.
