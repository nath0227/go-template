// Package tasks provides high-level workflow runners for common SE agent tasks.
package tasks

import (
	"context"
	"fmt"

	"github.com/nath0227/go-template/internal/agent"
)

// RunBugFix runs the bug-fixing workflow with the given description.
// The agent will autonomously explore the codebase, identify the root cause,
// apply a fix, and verify it.
func RunBugFix(ctx context.Context, description string) error {
	a := agent.New(agent.AgentOptions{
		SystemPrompt: agent.BugFixPrompt,
		MaxTurns:     20,
		Model:        "claude-opus-4-8",
	})

	fmt.Printf("Starting bug-fix agent for: %s\n\n", description)

	result, err := a.Run(ctx, description)
	if err != nil {
		return fmt.Errorf("bug-fix agent failed: %w", err)
	}

	fmt.Println("=== Bug Fix Complete ===")
	fmt.Println(result)
	return nil
}
