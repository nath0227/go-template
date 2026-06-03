package tasks

import (
	"context"
	"fmt"

	"github.com/nath0227/go-template/internal/agent"
)

// RunFeature runs the feature implementation workflow with the given description.
// The agent will explore the codebase, plan the implementation, write the code,
// and verify it with tests.
func RunFeature(ctx context.Context, description string) error {
	a := agent.New(agent.AgentOptions{
		SystemPrompt: agent.FeaturePrompt,
		MaxTurns:     30,
		Model:        "claude-opus-4-8",
	})

	fmt.Printf("Starting feature implementation agent for: %s\n\n", description)

	result, err := a.Run(ctx, description)
	if err != nil {
		return fmt.Errorf("feature agent failed: %w", err)
	}

	fmt.Println("=== Feature Implementation Complete ===")
	fmt.Println(result)
	return nil
}
