package tasks

import (
	"context"
	"fmt"

	"github.com/nath0227/go-template/internal/agent"
)

// RunCodeReview runs the code review workflow on the given path.
// The agent will read the specified files and produce a structured review.
func RunCodeReview(ctx context.Context, path string) error {
	a := agent.New(agent.AgentOptions{
		SystemPrompt: agent.CodeReviewPrompt,
		MaxTurns:     15,
		Model:        "claude-opus-4-8",
	})

	prompt := fmt.Sprintf("Please review the code at path: %s\n\nProvide a thorough code review covering correctness, security, performance, and maintainability.", path)

	fmt.Printf("Starting code review agent for: %s\n\n", path)

	result, err := a.Run(ctx, prompt)
	if err != nil {
		return fmt.Errorf("code review agent failed: %w", err)
	}

	fmt.Println("=== Code Review Complete ===")
	fmt.Println(result)
	return nil
}
