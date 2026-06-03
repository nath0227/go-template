// Command agent is a CLI for the AI software engineering agent.
//
// Usage:
//
//	agent bugfix <description>   — fix a bug described in plain text
//	agent review <path>          — review code at a path
//	agent feature <description>  — implement a feature
//	agent chat                   — interactive multi-turn session
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nath0227/go-template/internal/agent"
	"github.com/nath0227/go-template/internal/tasks"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()
	subcommand := os.Args[1]

	// Allow help without an API key.
	if subcommand != "help" && subcommand != "-h" && subcommand != "--help" {
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
		}
	}

	switch subcommand {
	case "bugfix":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: agent bugfix <description>")
		}
		description := strings.Join(os.Args[2:], " ")
		return tasks.RunBugFix(ctx, description)

	case "review":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: agent review <path>")
		}
		path := os.Args[2]
		return tasks.RunCodeReview(ctx, path)

	case "feature":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: agent feature <description>")
		}
		description := strings.Join(os.Args[2:], " ")
		return tasks.RunFeature(ctx, description)

	case "chat":
		return runChat(ctx)

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		printUsage()
		return fmt.Errorf("unknown subcommand: %q", subcommand)
	}
}

// runChat starts an interactive multi-turn session.
func runChat(ctx context.Context) error {
	a := agent.New(agent.AgentOptions{
		SystemPrompt: agent.ChatPrompt,
		MaxTurns:     50,
		Model:        "claude-opus-4-8",
	})

	fmt.Println("AI Software Engineering Agent — Interactive Chat")
	fmt.Println("Type 'exit' or 'quit' to end the session, 'clear' to reset history.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		switch strings.ToLower(input) {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return nil
		case "clear":
			a.ClearHistory()
			fmt.Println("Conversation history cleared.")
			continue
		}

		result, err := a.RunTurn(ctx, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
			continue
		}
		fmt.Printf("\nAgent: %s\n\n", result)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `AI Software Engineering Agent

Usage:
  agent <subcommand> [args]

Subcommands:
  bugfix <description>   Fix a bug described in plain text
  review <path>          Review code at the given file or directory path
  feature <description>  Implement a feature described in plain text
  chat                   Start an interactive multi-turn session

Environment:
  ANTHROPIC_API_KEY      Required: your Anthropic API key

Examples:
  agent bugfix "The login endpoint returns 500 when the email field is empty"
  agent review ./internal/auth
  agent feature "Add rate limiting to the REST API using a token bucket algorithm"
  agent chat
`)
}
