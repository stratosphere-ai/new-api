// Package agent orchestrates multi-step LLM + tool-calling runs. It is a
// thin layer on top of the relay: the orchestrator decides the next step
// and calls the relay internally, so billing/audit still flow through the
// standard middleware stack.
package agent

import "context"

// Input initiates an agent run.
type Input struct {
	OrgID   uint
	UserID  uint
	AgentID uint
	Prompt  string
	Vars    map[string]any
}

// Step describes one step of the run (model call, tool call, or KB query).
type Step struct {
	Kind      string // "llm" | "tool" | "kb"
	StartedAt int64
	EndedAt   int64
	Raw       map[string]any
}

// Output is the final result of a run.
type Output struct {
	Answer string
	Steps  []Step
}

// Agent runs a single request to completion, streaming intermediate steps
// via the returned channel when applicable.
type Agent interface {
	Run(ctx context.Context, in Input) (Output, error)
}

// ToolCall is what the orchestrator passes to ToolProxy.
type ToolCall struct {
	Name string
	Args map[string]any
}

// ToolResult is returned by ToolProxy.Invoke.
type ToolResult struct {
	OK  bool
	Raw map[string]any
}

// ToolProxy executes a tool (HTTP, built-in, or MCP-bridged).
type ToolProxy interface {
	Invoke(ctx context.Context, call ToolCall) (ToolResult, error)
}
