// Package mcp bridges the Model Context Protocol so external MCP servers
// can be registered by an org and invoked as tools by Agent runs.
package mcp

import "context"

// Client connects to one MCP server (stdio/http) and lists/invokes tools.
type Client interface {
	ListTools(ctx context.Context) ([]Tool, error)
	Call(ctx context.Context, name string, args map[string]any) (map[string]any, error)
	Close() error
}

// Tool describes an MCP-exposed tool schema.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
}
