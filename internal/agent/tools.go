package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

// Tool represents a tool that can be called by Claude.
type Tool struct {
	Definition anthropic.ToolUnionParam
	Execute    func(ctx context.Context, input json.RawMessage) (string, error)
}

// allTools returns all available tools.
func allTools() []Tool {
	return []Tool{
		readFileTool(),
		writeFileTool(),
		editFileTool(),
		runBashTool(),
		findFilesTool(),
		searchInFilesTool(),
		listDirectoryTool(),
	}
}

// toolParam is a helper to create a ToolUnionParam wrapping a ToolParam.
func toolParam(name, description string, properties map[string]any, required []string) anthropic.ToolUnionParam {
	return anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        name,
			Description: anthropic.String(description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: properties,
				Required:   required,
			},
		},
	}
}

// prop is a helper to create a JSON schema property map.
func prop(typ, description string) map[string]any {
	return map[string]any{"type": typ, "description": description}
}

// readFileTool reads the contents of a file.
func readFileTool() Tool {
	return Tool{
		Definition: toolParam(
			"read_file",
			"Read the contents of a file at the given path. Returns the file contents as a string.",
			map[string]any{
				"path": prop("string", "The path to the file to read"),
			},
			[]string{"path"},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			data, err := os.ReadFile(args.Path)
			if err != nil {
				return "", fmt.Errorf("failed to read file %q: %w", args.Path, err)
			}
			return string(data), nil
		},
	}
}

// writeFileTool writes content to a file, creating it if it doesn't exist.
func writeFileTool() Tool {
	return Tool{
		Definition: toolParam(
			"write_file",
			"Write content to a file at the given path. Creates the file (and any parent directories) if it doesn't exist, or overwrites it if it does.",
			map[string]any{
				"path":    prop("string", "The path to the file to write"),
				"content": prop("string", "The content to write to the file"),
			},
			[]string{"path", "content"},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			if err := os.MkdirAll(filepath.Dir(args.Path), 0o755); err != nil {
				return "", fmt.Errorf("failed to create directories for %q: %w", args.Path, err)
			}
			if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
				return "", fmt.Errorf("failed to write file %q: %w", args.Path, err)
			}
			return fmt.Sprintf("Successfully wrote %d bytes to %s", len(args.Content), args.Path), nil
		},
	}
}

// editFileTool performs a find-and-replace in a file.
func editFileTool() Tool {
	return Tool{
		Definition: toolParam(
			"edit_file",
			"Edit a file by replacing the first occurrence of old_string with new_string. The old_string must uniquely identify the section to replace. Returns an error if old_string is not found or appears multiple times.",
			map[string]any{
				"path":       prop("string", "The path to the file to edit"),
				"old_string": prop("string", "The exact string to find and replace (must be unique in the file)"),
				"new_string": prop("string", "The replacement string"),
			},
			[]string{"path", "old_string", "new_string"},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Path      string `json:"path"`
				OldString string `json:"old_string"`
				NewString string `json:"new_string"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			data, err := os.ReadFile(args.Path)
			if err != nil {
				return "", fmt.Errorf("failed to read file %q: %w", args.Path, err)
			}
			content := string(data)
			count := strings.Count(content, args.OldString)
			if count == 0 {
				return "", fmt.Errorf("old_string not found in %s", args.Path)
			}
			if count > 1 {
				return "", fmt.Errorf("old_string appears %d times in %s — provide more context to make it unique", count, args.Path)
			}
			newContent := strings.Replace(content, args.OldString, args.NewString, 1)
			if err := os.WriteFile(args.Path, []byte(newContent), 0o644); err != nil {
				return "", fmt.Errorf("failed to write file %q: %w", args.Path, err)
			}
			return fmt.Sprintf("Successfully edited %s", args.Path), nil
		},
	}
}

// runBashTool executes a bash command and returns stdout+stderr.
func runBashTool() Tool {
	return Tool{
		Definition: anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        "run_bash",
				Description: anthropic.String("Execute a bash command and return its combined stdout and stderr. Output is truncated to 10KB. Use this to run tests, build the project, check git status, or any other shell operation."),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"command": prop("string", "The bash command to execute"),
						"timeout_seconds": map[string]any{
							"type":        "number",
							"description": "Timeout in seconds (default: 30, max: 120)",
						},
					},
					Required: []string{"command"},
				},
			},
		},
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Command        string  `json:"command"`
				TimeoutSeconds float64 `json:"timeout_seconds"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			timeout := 30 * time.Second
			if args.TimeoutSeconds > 0 {
				if args.TimeoutSeconds > 120 {
					args.TimeoutSeconds = 120
				}
				timeout = time.Duration(args.TimeoutSeconds * float64(time.Second))
			}
			cmdCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "bash", "-c", args.Command)
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf

			err := cmd.Run()
			output := buf.String()

			// Limit output to 10KB
			const maxOutput = 10 * 1024
			if len(output) > maxOutput {
				output = output[:maxOutput] + "\n... (output truncated)"
			}

			if err != nil {
				// Include exit error info but still return the output
				return fmt.Sprintf("Command failed: %v\n%s", err, output), nil
			}
			if output == "" {
				return "(no output)", nil
			}
			return output, nil
		},
	}
}

// findFilesTool finds files matching a glob pattern.
func findFilesTool() Tool {
	return Tool{
		Definition: toolParam(
			"find_files",
			"Find files matching a glob pattern. Returns a newline-separated list of matching file paths.",
			map[string]any{
				"pattern": prop("string", "Glob pattern to match files (e.g. '**/*.go', 'src/**/*.ts', '*.json')"),
				"base_dir": map[string]any{
					"type":        "string",
					"description": "Base directory to search from (default: current directory)",
				},
			},
			[]string{"pattern"},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Pattern string `json:"pattern"`
				BaseDir string `json:"base_dir"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			baseDir := args.BaseDir
			if baseDir == "" {
				baseDir = "."
			}

			// Use find command for ** glob support
			cmd := exec.CommandContext(ctx, "find", baseDir, "-type", "f", "-name", filepath.Base(args.Pattern), "-not", "-path", "*/\\.*")
			if strings.Contains(args.Pattern, "/") {
				// For patterns with directory components, fall back to shell glob
				cmd = exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("find %s -type f | grep -E '%s' 2>/dev/null || true", baseDir, globToRegex(args.Pattern)))
			}
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			if err := cmd.Run(); err != nil && buf.Len() == 0 {
				return "No files found", nil
			}
			result := strings.TrimSpace(buf.String())
			if result == "" {
				return "No files found", nil
			}
			return result, nil
		},
	}
}

// globToRegex is a simple glob-to-regex converter for basic patterns.
func globToRegex(pattern string) string {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "**", ".*")
	pattern = strings.ReplaceAll(pattern, "*", "[^/]*")
	return pattern
}

// searchInFilesTool searches for a pattern across files.
func searchInFilesTool() Tool {
	return Tool{
		Definition: toolParam(
			"search_in_files",
			"Search for a text pattern across files using grep. Returns matching lines with file paths and line numbers.",
			map[string]any{
				"pattern": prop("string", "The grep pattern to search for (supports basic regex)"),
				"path": map[string]any{
					"type":        "string",
					"description": "Directory or file to search in (default: current directory)",
				},
				"file_pattern": map[string]any{
					"type":        "string",
					"description": "File name pattern to restrict search (e.g. '*.go', '*.py')",
				},
				"context_lines": map[string]any{
					"type":        "number",
					"description": "Number of context lines before and after each match (default: 2)",
				},
			},
			[]string{"pattern"},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Pattern      string  `json:"pattern"`
				Path         string  `json:"path"`
				FilePattern  string  `json:"file_pattern"`
				ContextLines float64 `json:"context_lines"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			searchPath := args.Path
			if searchPath == "" {
				searchPath = "."
			}
			contextLines := 2
			if args.ContextLines > 0 {
				contextLines = int(args.ContextLines)
			}

			grepArgs := []string{"-rn", fmt.Sprintf("-C%d", contextLines), args.Pattern, searchPath}
			if args.FilePattern != "" {
				grepArgs = append([]string{"--include=" + args.FilePattern}, grepArgs...)
				grepArgs = []string{"-rn", fmt.Sprintf("-C%d", contextLines), "--include=" + args.FilePattern, args.Pattern, searchPath}
			}

			cmd := exec.CommandContext(ctx, "grep", grepArgs...)
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			cmd.Run() // grep returns exit code 1 when no matches — that's fine

			result := strings.TrimSpace(buf.String())
			if result == "" {
				return "No matches found", nil
			}

			// Limit output
			const maxOutput = 10 * 1024
			if len(result) > maxOutput {
				result = result[:maxOutput] + "\n... (output truncated)"
			}
			return result, nil
		},
	}
}

// listDirectoryTool lists the contents of a directory.
func listDirectoryTool() Tool {
	return Tool{
		Definition: toolParam(
			"list_directory",
			"List the contents of a directory. Returns file names with type indicators (/ for directories).",
			map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The directory path to list (default: current directory)",
				},
				"recursive": map[string]any{
					"type":        "boolean",
					"description": "Whether to list subdirectories recursively (default: false)",
				},
			},
			[]string{},
		),
		Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
			var args struct {
				Path      string `json:"path"`
				Recursive bool   `json:"recursive"`
			}
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			path := args.Path
			if path == "" {
				path = "."
			}

			var cmd *exec.Cmd
			if args.Recursive {
				cmd = exec.CommandContext(ctx, "find", path, "-not", "-path", "*/\\.*", "-not", "-path", "*/vendor/*")
			} else {
				cmd = exec.CommandContext(ctx, "ls", "-la", path)
			}
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to list directory %q: %s", path, buf.String())
			}
			return buf.String(), nil
		},
	}
}
