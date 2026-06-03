# go-template

A production-quality Go AI software engineering agent template built on the [Anthropic Claude API](https://docs.anthropic.com/). This template demonstrates the key patterns for building autonomous AI agents that can read, write, and edit code; run shell commands; and iterate in a tool-use loop to complete real software engineering tasks.

## Features

- **Agentic loop** — Claude autonomously reads files, runs commands, edits code, and verifies its work
- **Seven built-in tools** — read/write/edit files, bash execution, file search, text search, directory listing
- **Specialized roles** — distinct system prompts for bug fixing, code review, and feature implementation
- **Multi-turn chat** — interactive sessions with preserved conversation history
- **Extensible** — add new tools or task types in a few lines of Go

## Prerequisites

- Go 1.22 or newer
- An [Anthropic API key](https://console.anthropic.com/)

## Installation

```bash
git clone https://github.com/nath0227/go-template
cd go-template
go build -o agent ./cmd/agent
```

## Setup

Export your Anthropic API key before running:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

## Usage

### Bug Fixing

Describe a bug in plain English and let the agent explore the codebase, identify the root cause, and apply a fix:

```bash
./agent bugfix "The login endpoint returns HTTP 500 when the email field is empty"
./agent bugfix "Unit tests fail with index out of range in pkg/parser/parser.go"
```

### Code Review

Point the agent at a file or directory and receive a structured review covering correctness, security, performance, and maintainability:

```bash
./agent review ./internal/auth
./agent review ./cmd/server/main.go
```

### Feature Implementation

Describe a feature in plain text and the agent will explore the codebase, plan the implementation, write the code, and run tests:

```bash
./agent feature "Add rate limiting to the REST API using a token bucket algorithm"
./agent feature "Implement JWT refresh token rotation"
```

### Interactive Chat

Start an interactive multi-turn session for general software engineering questions or ad-hoc tasks:

```bash
./agent chat
```

In chat mode:
- Type `clear` to reset the conversation history
- Type `exit` or `quit` (or press Ctrl+D) to end the session

## Architecture

```
cmd/agent/main.go          CLI entry point — subcommand routing
internal/agent/
  agent.go                 Core agentic loop + conversation history
  tools.go                 Tool definitions and executor functions
  prompts.go               System prompt templates per role
internal/tasks/
  bugfix.go                Bug-fixing workflow (wraps the agent)
  review.go                Code review workflow
  feature.go               Feature implementation workflow
```

### Agentic Loop

The core loop in `internal/agent/agent.go`:

1. User prompt → added to message history
2. Call `client.Messages.New(...)` with tools and history
3. If Claude calls tools: execute them, add results to history, go to step 2
4. If Claude returns text with no tool calls: return the response

This continues for up to `MaxTurns` iterations (default varies per task type).

### Available Tools

| Tool | Description |
|------|-------------|
| `read_file` | Read the contents of a file |
| `write_file` | Write or create a file (creates parent dirs) |
| `edit_file` | Find-and-replace within a file (requires unique match) |
| `run_bash` | Execute a bash command with configurable timeout |
| `find_files` | Find files matching a glob pattern |
| `search_in_files` | Grep for text across files with context lines |
| `list_directory` | List directory contents, optionally recursive |

### System Prompts

Each task type uses a role-specific system prompt in `internal/agent/prompts.go`:

- `BugFixPrompt` — focuses on minimal, verified fixes with root-cause analysis
- `CodeReviewPrompt` — structured review across correctness, security, performance, and maintainability
- `FeaturePrompt` — codebase-aware implementation that follows existing conventions
- `ChatPrompt` — general-purpose assistant with file and command access

## Extending

### Adding a New Tool

Edit `internal/agent/tools.go`:

```go
func myTool() Tool {
    return Tool{
        Definition: toolParam(
            "my_tool",
            "Description of what this tool does",
            map[string]any{
                "arg": prop("string", "Description of arg"),
            },
            []string{"arg"}, // required fields
        ),
        Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
            var args struct {
                Arg string `json:"arg"`
            }
            if err := json.Unmarshal(input, &args); err != nil {
                return "", fmt.Errorf("invalid input: %w", err)
            }
            // ... do work ...
            return result, nil
        },
    }
}
```

Then add `myTool()` to the `allTools()` slice.

### Adding a New Task Type

1. Add a system prompt constant in `internal/agent/prompts.go`
2. Create `internal/tasks/mytask.go`:

```go
func RunMyTask(ctx context.Context, description string) error {
    a := agent.New(agent.AgentOptions{
        SystemPrompt: agent.MyTaskPrompt,
        MaxTurns:     20,
        Model:        "claude-opus-4-8",
    })
    result, err := a.Run(ctx, description)
    if err != nil {
        return err
    }
    fmt.Println(result)
    return nil
}
```

3. Wire it up in `cmd/agent/main.go` as a new `case "mytask":` branch.

## Development

```bash
# Run tests
go test ./...

# Build
go build ./cmd/agent

# Vet
go vet ./...

# Update dependencies
go mod tidy
```

## License

MIT
