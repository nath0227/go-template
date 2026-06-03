// Package agent implements the core agentic loop for the AI software engineering agent.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

const (
	defaultModel    = string(anthropic.ModelClaudeOpus4_8)
	defaultMaxTurns = 20
	defaultMaxTokens = 8192
)

// AgentOptions configures an Agent.
type AgentOptions struct {
	// SystemPrompt is the system prompt sent with every request.
	SystemPrompt string
	// MaxTurns is the maximum number of agentic loop iterations (default: 20).
	MaxTurns int
	// Model is the Claude model to use (default: claude-opus-4-8).
	Model string
	// MaxTokens is the maximum number of tokens per response (default: 8192).
	MaxTokens int64
}

// Agent is a stateful Claude-powered agent that can use tools.
type Agent struct {
	client   anthropic.Client
	model    string
	tools    []Tool
	history  []anthropic.MessageParam
	maxTurns int
	maxTokens int64
	system   string
}

// New creates a new Agent with the given options.
func New(opts AgentOptions) *Agent {
	model := opts.Model
	if model == "" {
		model = defaultModel
	}
	maxTurns := opts.MaxTurns
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	return &Agent{
		client:    anthropic.NewClient(),
		model:     model,
		tools:     allTools(),
		maxTurns:  maxTurns,
		maxTokens: maxTokens,
		system:    opts.SystemPrompt,
	}
}

// Run runs the agentic loop with the given prompt, starting fresh (no history).
// It loops until Claude produces a final text response with no tool calls,
// or until maxTurns is reached.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	// Start fresh for a one-shot task.
	a.history = nil
	return a.RunTurn(ctx, prompt)
}

// RunTurn adds the prompt to history and runs the agentic loop.
// Unlike Run, it preserves existing history for multi-turn sessions.
func (a *Agent) RunTurn(ctx context.Context, prompt string) (string, error) {
	// Add the user message to history.
	a.history = append(a.history, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))

	// Build tool definitions for the API.
	toolDefs := make([]anthropic.ToolUnionParam, len(a.tools))
	for i, t := range a.tools {
		toolDefs[i] = t.Definition
	}

	// Build a map for fast tool lookup.
	toolMap := make(map[string]func(context.Context, json.RawMessage) (string, error), len(a.tools))
	for _, t := range a.tools {
		name := toolName(t.Definition)
		toolMap[name] = t.Execute
	}

	for turn := 0; turn < a.maxTurns; turn++ {
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(a.model),
			MaxTokens: a.maxTokens,
			Messages:  a.history,
			Tools:     toolDefs,
		}
		if a.system != "" {
			params.System = []anthropic.TextBlockParam{{Text: a.system}}
		}

		resp, err := a.client.Messages.New(ctx, params)
		if err != nil {
			return "", fmt.Errorf("API call failed on turn %d: %w", turn+1, err)
		}

		// Collect any tool calls from the response.
		var toolCalls []anthropic.ToolUseBlock
		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				toolCalls = append(toolCalls, block.AsToolUse())
			}
		}

		// Add the assistant message to history.
		a.history = append(a.history, resp.ToParam())

		// If no tool calls, Claude is done — return the text response.
		if len(toolCalls) == 0 {
			return extractText(resp), nil
		}

		// Execute each tool call and collect results.
		toolResults := make([]anthropic.ContentBlockParamUnion, 0, len(toolCalls))
		for _, toolUse := range toolCalls {
			result, toolErr := a.executeTool(ctx, toolMap, toolUse)
			isError := toolErr != nil
			content := result
			if toolErr != nil {
				content = toolErr.Error()
			}
			toolResults = append(toolResults, anthropic.NewToolResultBlock(toolUse.ID, content, isError))
		}

		// Add the tool results as a user message.
		a.history = append(a.history, anthropic.NewUserMessage(toolResults...))
	}

	return "", fmt.Errorf("reached maximum turns (%d) without a final response", a.maxTurns)
}

// executeTool dispatches a tool call to the appropriate executor.
func (a *Agent) executeTool(ctx context.Context, toolMap map[string]func(context.Context, json.RawMessage) (string, error), toolUse anthropic.ToolUseBlock) (string, error) {
	executor, ok := toolMap[toolUse.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %q", toolUse.Name)
	}
	result, err := executor(ctx, toolUse.Input)
	if err != nil {
		return "", fmt.Errorf("tool %q failed: %w", toolUse.Name, err)
	}
	return result, nil
}

// History returns a copy of the current conversation history.
func (a *Agent) History() []anthropic.MessageParam {
	cp := make([]anthropic.MessageParam, len(a.history))
	copy(cp, a.history)
	return cp
}

// ClearHistory resets the conversation history.
func (a *Agent) ClearHistory() {
	a.history = nil
}

// extractText returns the concatenated text content from a Message.
func extractText(msg *anthropic.Message) string {
	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			tb := block.AsText()
			sb.WriteString(tb.Text)
		}
	}
	return sb.String()
}

// toolName extracts the tool name from a ToolUnionParam.
func toolName(t anthropic.ToolUnionParam) string {
	if t.OfTool != nil {
		return t.OfTool.Name
	}
	return ""
}
