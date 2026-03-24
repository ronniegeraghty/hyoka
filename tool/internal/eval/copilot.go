package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/progress"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/prompt"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
)

// CopilotSDKEvaluator uses the Copilot SDK to run real evaluations.
type CopilotSDKEvaluator struct {
	clientOpts *copilot.ClientOptions
	debug      bool
	progressFn progress.ProgressFunc
}

// SetProgressFunc registers a callback for live progress updates.
func (e *CopilotSDKEvaluator) SetProgressFunc(fn progress.ProgressFunc) {
	e.progressFn = fn
}

// CopilotEvalOptions configures the CopilotSDKEvaluator.
type CopilotEvalOptions struct {
	// GitHubToken for SDK authentication (optional; falls back to logged-in user).
	GitHubToken string
	// CLIPath overrides the Copilot CLI executable path.
	CLIPath string
	// Debug enables verbose logging.
	Debug bool
}

// NewCopilotSDKEvaluator creates a new evaluator backed by the Copilot SDK.
func NewCopilotSDKEvaluator(opts CopilotEvalOptions) *CopilotSDKEvaluator {
	clientOpts := &copilot.ClientOptions{}
	if opts.GitHubToken != "" {
		clientOpts.GitHubToken = opts.GitHubToken
	}
	if opts.CLIPath != "" {
		clientOpts.CLIPath = opts.CLIPath
	}
	if opts.Debug {
		clientOpts.LogLevel = "debug"
	}
	return &CopilotSDKEvaluator{
		clientOpts: clientOpts,
		debug:      opts.Debug,
	}
}

// Evaluate runs a prompt through a real Copilot session and returns generated files and events.
func (e *CopilotSDKEvaluator) Evaluate(ctx context.Context, p *prompt.Prompt, cfg *config.ToolConfig, workDir string) (*EvalResult, error) {
	// Copy starter project if configured
	var starterFiles []string
	if p.StarterProject != "" {
		starterDir := p.StarterProject
		if !filepath.IsAbs(starterDir) && p.FilePath != "" {
			starterDir = filepath.Join(filepath.Dir(p.FilePath), starterDir)
		}
		if err := copyDir(starterDir, workDir); err != nil {
			return nil, fmt.Errorf("copying starter project: %w", err)
		}
		starterFiles, _ = listFiles(workDir)
	}

	// Create Copilot client
	opts := *e.clientOpts
	opts.Cwd = workDir
	client := copilot.NewClient(&opts)

	if err := client.Start(ctx); err != nil {
		return &EvalResult{
			Error:        fmt.Sprintf("copilot client start failed: %v", err),
			ErrorDetails: err.Error(),
		}, fmt.Errorf("starting copilot client: %w", err)
	}
	// Defer client cleanup. We use client.Stop() which sends session.destroy
	// for a graceful shutdown. The generated files are in the report tree
	// (not a temp dir), so session.destroy won't delete them.
	defer func() {
		done := make(chan struct{})
		go func() { client.Stop(); close(done) }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			client.ForceStop()
		}
	}()

	// Build session config from tool config
	sessionCfg := e.buildSessionConfig(cfg, workDir)

	session, err := client.CreateSession(ctx, sessionCfg)
	if err != nil {
		return &EvalResult{
			Error:        fmt.Sprintf("session creation failed: %v", err),
			ErrorDetails: err.Error(),
		}, fmt.Errorf("creating session: %w", err)
	}
	// No session.Disconnect() — that sends session.destroy which causes the
	// CLI to delete generated files from WorkingDirectory. ForceStop above
	// handles process cleanup without file deletion.

	// Subscribe to events with detailed capture and debug logging
	var events []copilot.SessionEvent
	var sessionRecords []report.SessionEventRecord
	var mu sync.Mutex
	debugPrefix := p.ID + "/" + cfg.Name
	unsub := session.On(func(event copilot.SessionEvent) {
		mu.Lock()
		events = append(events, event)

		// Build serializable event record
		rec := report.SessionEventRecord{
			Type: string(event.Type),
		}
		if event.Data.ToolName != nil {
			rec.ToolName = *event.Data.ToolName
		}
		if event.Data.Content != nil {
			rec.Content = *event.Data.Content
		}
		if event.Data.Arguments != nil {
			if argsBytes, err := json.Marshal(event.Data.Arguments); err == nil {
				rec.ToolArgs = string(argsBytes)
			}
		}
		if event.Data.Result != nil {
			if event.Data.Result.DetailedContent != nil {
				rec.ToolResult = *event.Data.Result.DetailedContent
			} else if event.Data.Result.Content != nil {
				rec.ToolResult = *event.Data.Result.Content
			}
		}
		if event.Data.Error != nil {
			if event.Data.Error.ErrorClass != nil {
				rec.Error = event.Data.Error.ErrorClass.Message
			} else if event.Data.Error.String != nil {
				rec.Error = *event.Data.Error.String
			}
		}
		if event.Data.Success != nil {
			rec.ToolSuccess = event.Data.Success
		}
		if event.Data.Duration != nil {
			rec.Duration = *event.Data.Duration
		}
		if event.Data.MCPServerName != nil {
			rec.MCPServerName = *event.Data.MCPServerName
		}
		if event.Data.MCPToolName != nil {
			rec.MCPToolName = *event.Data.MCPToolName
		}
		if event.Data.Path != nil {
			rec.FilePath = *event.Data.Path
		}
		sessionRecords = append(sessionRecords, rec)
		mu.Unlock()

		// Forward progress events to display
		if e.progressFn != nil {
			evalID := debugPrefix
			switch event.Type {
			case copilot.SessionEventTypeToolExecutionStart:
				toolName := ""
				if event.Data.ToolName != nil {
					toolName = *event.Data.ToolName
				}
				if isFileWriteTool(toolName) {
					arg := toolArgSummary(event)
					e.progressFn(progress.ProgressEvent{
						EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
						Type:    progress.EventWritingFile,
						Message: toolName + " → " + arg,
					})
				} else {
					arg := toolArgSummary(event)
					msg := toolName
					if arg != "" {
						msg = toolName + " → " + arg
					}
					e.progressFn(progress.ProgressEvent{
						EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
						Type:    progress.EventToolStart,
						Message: msg,
					})
				}
			case copilot.SessionEventTypeToolExecutionComplete:
				toolName := ""
				if event.Data.ToolName != nil {
					toolName = *event.Data.ToolName
				}
				result := ""
				if event.Data.Result != nil && event.Data.Result.Content != nil {
					result = truncateStr(*event.Data.Result.Content, 60)
				}
				msg := toolName
				if result != "" {
					msg = toolName + " → " + result
				}
				e.progressFn(progress.ProgressEvent{
					EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
					Type:    progress.EventToolComplete,
					Message: msg,
				})
			case copilot.SessionEventTypeAssistantMessage:
				content := ""
				if event.Data.Content != nil {
					content = *event.Data.Content
				}
				if content != "" {
					summary := truncateStr(content, 80)
					e.progressFn(progress.ProgressEvent{
						EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
						Type:    progress.EventReasoning,
						Message: summary,
					})
				}
			}
		}

		// Debug logging to stderr — all event types
		if e.debug {
			switch event.Type {
			case copilot.SessionEventTypeToolExecutionStart:
				toolName := ""
				if event.Data.ToolName != nil {
					toolName = *event.Data.ToolName
				}
				log.Printf("[DEBUG] %s: ⚙ Tool start: %s", debugPrefix, toolName)
			case copilot.SessionEventTypeToolExecutionComplete:
				toolName := ""
				if event.Data.ToolName != nil {
					toolName = *event.Data.ToolName
				}
				content := ""
				if event.Data.Content != nil {
					content = truncateStr(*event.Data.Content, 200)
				}
				log.Printf("[DEBUG] %s: ✓ Tool done: %s → %s", debugPrefix, toolName, content)
			case copilot.SessionEventTypeAssistantMessage:
				content := ""
				if event.Data.Content != nil {
					content = *event.Data.Content
				}
				if content != "" {
					// Detect file creation patterns
					if summary := detectFileCreation(content); summary != "" {
						log.Printf("[DEBUG] %s: ← Assistant creating file: %s", debugPrefix, summary)
					} else {
						log.Printf("[DEBUG] %s: ← Assistant: %s", debugPrefix, truncateStr(content, 200))
					}
				}
			case copilot.SessionEventTypeSessionError:
				content := ""
				if event.Data.Content != nil {
					content = *event.Data.Content
				}
				log.Printf("[DEBUG] %s: ✗ Session error: %s", debugPrefix, content)
			default:
				// Log all other events (session.start, session.idle, user.message, assistant.turn_end, etc.)
				content := ""
				if event.Data.Content != nil {
					content = truncateStr(*event.Data.Content, 100)
				}
				if content != "" {
					log.Printf("[DEBUG] %s:   Event %s: %s", debugPrefix, event.Type, content)
				} else {
					log.Printf("[DEBUG] %s:   Event %s", debugPrefix, event.Type)
				}
			}
		}
	})
	defer unsub()

	// Send the prompt
	if e.progressFn != nil {
		e.progressFn(progress.ProgressEvent{
			EvalID: debugPrefix, PromptID: p.ID, ConfigName: cfg.Name,
			Type:    progress.EventSendingPrompt,
			Message: fmt.Sprintf("Sending prompt (%d chars)...", len(p.PromptText)),
		})
	}
	if e.debug {
		log.Printf("[DEBUG] %s: → Sending prompt (%d chars)...", debugPrefix, len(p.PromptText))
	}
	_, err = session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: p.PromptText,
	})
	if err != nil {
		mu.Lock()
		captured := make([]report.SessionEventRecord, len(sessionRecords))
		copy(captured, sessionRecords)
		capturedEvts := make([]copilot.SessionEvent, len(events))
		copy(capturedEvts, events)
		mu.Unlock()
		return &EvalResult{
			SessionEvents: captured,
			EventCount:    len(captured),
			ToolCalls:     extractToolCalls(capturedEvts),
			Error:         fmt.Sprintf("prompt send failed: %v", err),
			ErrorDetails:  err.Error(),
		}, fmt.Errorf("sending prompt: %w", err)
	}

	// Collect results
	mu.Lock()
	capturedEvents := make([]copilot.SessionEvent, len(events))
	copy(capturedEvents, events)
	capturedRecords := make([]report.SessionEventRecord, len(sessionRecords))
	copy(capturedRecords, sessionRecords)
	mu.Unlock()

	generatedFiles, _ := listFiles(workDir)
	toolCalls := extractToolCalls(capturedEvents)
	hasError := hasSessionError(capturedEvents)

	if e.debug {
		log.Printf("[DEBUG] %s: %d events, %d tool calls, %d files",
			debugPrefix, len(capturedEvents), len(toolCalls), len(generatedFiles))
	}

	return &EvalResult{
		GeneratedFiles: generatedFiles,
		EventCount:     len(capturedEvents),
		ToolCalls:      toolCalls,
		SessionEvents:  capturedRecords,
		Success:        !hasError,
		Error:          "",
		StarterFiles:   starterFiles,
	}, nil
}

// truncateStr truncates a string to maxLen characters, appending "..." if truncated.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// detectFileCreation checks if assistant content looks like file creation
// and returns a summary like "key_vault_crud.py (89 lines)".
func detectFileCreation(content string) string {
	lines := strings.Split(content, "\n")
	// Look for patterns like "```python", file path references, or create_file tool patterns
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Detect "Creating file: filename" or "Writing filename" patterns
		for _, prefix := range []string{"creating file:", "writing file:", "creating ", "writing "} {
			lower := strings.ToLower(trimmed)
			if strings.HasPrefix(lower, prefix) {
				filename := strings.TrimSpace(trimmed[len(prefix):])
				if filename != "" && !strings.Contains(filename, " ") {
					lineCount := len(lines)
					return fmt.Sprintf("%s (%d lines)", filename, lineCount)
				}
			}
		}
	}
	// If content is very long (likely code), summarize by line count
	if len(lines) > 20 {
		// Try to find a filename from a markdown code fence
		for _, line := range lines[:min(5, len(lines))] {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") && len(trimmed) > 3 {
				return fmt.Sprintf("code block (%d lines)", len(lines))
			}
		}
	}
	return ""
}



// Client returns a new Copilot client for the given working directory.
// Exported for use by the review package.
func (e *CopilotSDKEvaluator) Client(ctx context.Context, workDir string) (*copilot.Client, error) {
	opts := *e.clientOpts
	opts.Cwd = workDir
	client := copilot.NewClient(&opts)
	if err := client.Start(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (e *CopilotSDKEvaluator) buildSessionConfig(cfg *config.ToolConfig, workDir string) *copilot.SessionConfig {
	// Use generator-specific skills if configured, otherwise fall back to shared
	skillDirs := cfg.GeneratorSkillDirectories
	if len(skillDirs) == 0 {
		skillDirs = cfg.SkillDirectories
	}

	// System message ensures the agent creates actual code files in the workspace
	// rather than responding with inline text or writing files to the wrong directory.
	// The create tool requires an explicit 'path' parameter — without it, files land in ~.
	// Also constrains bash usage, discourages web_fetch research, and standardizes on python3.
	systemMsg := fmt.Sprintf(
		"You are a code generation agent. Your working directory is: %s\n"+
			"CRITICAL FILE CREATION RULES:\n"+
			"1. Always write code to files using the create tool — never just explain code in text.\n"+
			"2. Always create files using the create tool. Never provide code inline in your response.\n"+
			"3. Every create call MUST include the 'path' parameter with a FULL ABSOLUTE PATH.\n"+
			"4. Every file path MUST start with: %s/\n"+
			"5. Example: create with path=\"%s/main.py\"\n"+
			"6. NEVER omit the path parameter. NEVER use relative paths.\n"+
			"7. NEVER create files outside your working directory.\n"+
			"BASH RULES:\n"+
			"8. When using bash, always cd to %s first. Never run commands from ~ or any directory outside your workspace.\n"+
			"RESEARCH RULES:\n"+
			"9. Do not use web_fetch to research documentation. Focus on generating code files based on the prompt.\n"+
			"PYTHON RULES:\n"+
			"10. Use python3 (not python) for all Python scripts and commands.",
		workDir, workDir, workDir, workDir,
	)

	sc := &copilot.SessionConfig{
		Model: cfg.Model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: systemMsg,
		},
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    skillDirs,
	}

	// Only set AvailableTools/ExcludedTools when non-empty.
	// An empty slice serializes as JSON [] which tells the CLI "zero tools" —
	// nil serializes as null which means "all default tools available."
	if len(cfg.AvailableTools) > 0 {
		sc.AvailableTools = cfg.AvailableTools
	}
	if len(cfg.ExcludedTools) > 0 {
		sc.ExcludedTools = cfg.ExcludedTools
	}

	// Map MCP servers
	if len(cfg.MCPServers) > 0 {
		sc.MCPServers = make(map[string]copilot.MCPServerConfig, len(cfg.MCPServers))
		for name, srv := range cfg.MCPServers {
			sc.MCPServers[name] = copilot.MCPServerConfig{
				"type":    srv.Type,
				"command": srv.Command,
				"args":    srv.Args,
			}
		}
	}

	return sc
}

// extractToolCalls returns unique tool names from session events.
func extractToolCalls(events []copilot.SessionEvent) []string {
	seen := make(map[string]bool)
	var tools []string
	for _, e := range events {
		if e.Type == copilot.SessionEventTypeToolExecutionStart ||
			e.Type == copilot.SessionEventTypeToolExecutionComplete {
			name := ""
			if e.Data.ToolName != nil {
				name = *e.Data.ToolName
			}
			if name != "" && !seen[name] {
				seen[name] = true
				tools = append(tools, name)
			}
		}
	}
	return tools
}

// hasSessionError checks for error events.
func hasSessionError(events []copilot.SessionEvent) bool {
	for _, e := range events {
		if e.Type == copilot.SessionEventTypeSessionError {
			return true
		}
	}
	return false
}

// isFileWriteTool returns true for tools that create or modify files.
func isFileWriteTool(name string) bool {
	switch name {
	case "create", "edit", "write_file", "create_file",
		"insert_edit_into_file", "write_to_file":
		return true
	}
	return false
}

// toolArgSummary extracts a short summary of the tool's primary argument.
func toolArgSummary(event copilot.SessionEvent) string {
	if event.Data.Path != nil && *event.Data.Path != "" {
		return filepath.Base(*event.Data.Path)
	}
	if event.Data.Arguments != nil {
		if args, ok := event.Data.Arguments.(map[string]interface{}); ok {
			for _, key := range []string{"path", "file", "command"} {
				if v, ok := args[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						if key == "path" || key == "file" {
							return filepath.Base(s)
						}
						return truncateStr(s, 40)
					}
				}
			}
		}
	}
	return ""
}
