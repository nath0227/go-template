# Go AI Software Engineering Agent

This repo contains a Go implementation of an AI software engineering agent powered by the Anthropic Claude API. The agent can autonomously read, write, and edit code, run shell commands, search files, and iterate in a tool-use loop to complete software engineering tasks.

## Project Layout

```
cmd/agent/main.go          CLI entry point (subcommands: bugfix, review, feature, chat)
internal/agent/
  agent.go                 Core agentic loop + message history management
  tools.go                 Tool definitions and executor functions
  prompts.go               System prompt constants for each task role
internal/tasks/
  bugfix.go                Bug-fixing workflow
  review.go                Code review workflow
  feature.go               Feature implementation workflow
go.mod                     Module: github.com/nath0227/go-template (Go 1.22+)
```

## Development Commands

```bash
# Build the CLI binary
go build ./cmd/agent

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Check for issues
go vet ./...

# Tidy module dependencies
go mod tidy
```

## Required Environment Variables

- `ANTHROPIC_API_KEY` — your Anthropic API key (required at runtime)

## Key Patterns

### Agentic Loop
The `Agent.RunTurn` method implements the standard tool-use loop:
1. Send messages + tools to Claude
2. Execute any tool calls (run_bash, read_file, edit_file, etc.)
3. Add results back to history as user messages
4. Repeat until Claude returns a response with no tool calls

### Message History
The `Agent` struct maintains `[]anthropic.MessageParam` history. `Run()` resets history for a fresh one-shot task; `RunTurn()` appends to existing history for multi-turn sessions.

### Tool Structure
Each `Tool` has:
- `Definition` — `anthropic.ToolUnionParam` with the JSON schema Claude uses to call the tool
- `Execute` — a Go function that receives `json.RawMessage` and returns `(string, error)`

### Adding New Tools
1. Write a function `myTool() Tool` in `internal/agent/tools.go`
2. Add it to the `allTools()` slice
3. The tool will automatically be available in all agent sessions

### Adding New Task Types
1. Add a new system prompt constant in `internal/agent/prompts.go`
2. Create `internal/tasks/mytask.go` following the pattern of `bugfix.go`
3. Wire it up as a new subcommand in `cmd/agent/main.go`

## Code Conventions

- Error handling: wrap errors with `fmt.Errorf("context: %w", err)`
- Contexts: always pass `context.Context` as the first argument
- Exported types have doc comments
- Tool executors are safe to call concurrently (they are stateless)
- Use `go test ./...` before committing any changes
