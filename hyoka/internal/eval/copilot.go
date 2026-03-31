package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/hyoka/internal/config"
	"github.com/ronniegeraghty/hyoka/internal/logging"
	"github.com/ronniegeraghty/hyoka/internal/progress"
	"github.com/ronniegeraghty/hyoka/internal/prompt"
	"github.com/ronniegeraghty/hyoka/internal/report"
)

// CopilotSDKEvaluator uses the Copilot SDK to run real evaluations.
type CopilotSDKEvaluator struct {
	clientOpts *copilot.ClientOptions
	allowCloud bool
	maxTurns   int
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
	// AllowCloud permits generated code to provision real cloud resources (#36).
	AllowCloud bool
	// MaxTurns limits assistant turns during generation. When reached, the
	// session context is cancelled to stop the run immediately (#69).
	MaxTurns int
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
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		clientOpts.LogLevel = "debug"
	}
	// Tag SDK-spawned processes with HYOKA_SESSION env var (#70).
	clientOpts.Env = HyokaBaseEnv()
	return &CopilotSDKEvaluator{
		clientOpts: clientOpts,
		allowCloud: opts.AllowCloud,
		maxTurns:   opts.MaxTurns,
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
	// Enrich env with prompt/config metadata for this specific eval (#70).
	opts.Env = HyokaEvalEnv(p.ID, cfg.Name)
	client := copilot.NewClient(&opts)

	if err := client.Start(ctx); err != nil {
		return &EvalResult{
			Error:        fmt.Sprintf("copilot client start failed: %v", err),
			ErrorDetails: err.Error(),
		}, fmt.Errorf("starting copilot client: %w", err)
	}
	// NOTE: ProcessTracker.Register/Deregister cannot be wired here because the
	// Copilot SDK's Client struct does not expose the underlying process PID (the
	// osProcess field is unexported). The SDK manages its own process lifecycle
	// via client.Stop()/ForceStop(). DefaultTracker.TerminateAll is still called
	// from the signal handler in engine.go for any future tracked processes.

	// Track session ID for cleanup — set after CreateSession.
	var sessionID string
	// Defer client cleanup (#62). Delete session state first (requires
	// connected client), then stop the client. DeleteSession sends
	// session.delete RPC which removes session-state dir AND the SQLite
	// session-store.db entry — unlike os.RemoveAll which misses the DB.
	defer func() {
		if sessionID != "" {
			deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer deleteCancel()
			if err := client.DeleteSession(deleteCtx, sessionID); err != nil {
				slog.Debug("session delete failed, session-state may remain",
					"sessionID", sessionID, "error", err)
			}
		}
		done := make(chan struct{})
		go func() { client.Stop(); close(done) }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			client.ForceStop()
		}
	}()

	// Build session config from tool config
	// Create isolated config directory to prevent user-level skills from
	// leaking into the eval session (#21). Only skills explicitly listed
	// in the eval config's SkillDirectories are loaded.
	configDir, err := NewIsolatedConfigDir()
	if err != nil {
		return nil, fmt.Errorf("creating isolated config dir: %w", err)
	}
	defer os.RemoveAll(configDir)

	sessionCfg := e.buildSessionConfig(cfg, workDir, configDir)

	// Subscribe to events with detailed capture and debug logging.
	// This MUST be set before CreateSession — the SDK reads OnEvent during
	// session creation and won't pick up a callback assigned afterwards.
	var events []copilot.SessionEvent
	var sessionRecords []report.SessionEventRecord
	var mu sync.Mutex
	debugPrefix := p.ID + "/" + cfg.Name
	// Structured logger for this eval session (#42)
	lg := logging.EvalLogger(p.ID, cfg.Name, "generation", 0)

	// Capture turn counter for expanded events
	var turnCounter int

	// Mid-generation turn limit (#69). Create a cancellable child context
	// so the OnEvent callback can stop runaway sessions in real time.
	genCtx, genCancel := context.WithCancel(ctx)
	defer genCancel()
	var turnLimitHit bool

	sessionCfg.OnEvent = func(event copilot.SessionEvent) {
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

		// Expanded event fields
		switch event.Type {
		case copilot.SessionEventTypeAssistantTurnStart:
			turnCounter++
			rec.TurnNumber = turnCounter
			lg.Info("Turn started", "turn", turnCounter)
			// Mid-generation turn limit (#69): cancel context to stop runaway sessions
			if e.maxTurns > 0 && turnCounter > e.maxTurns && !turnLimitHit {
				turnLimitHit = true
				lg.Warn("Turn limit reached mid-generation, cancelling session",
					"turn", turnCounter, "max_turns", e.maxTurns)
				genCancel()
			}
		case copilot.SessionEventTypeAssistantTurnEnd:
			rec.TurnNumber = turnCounter
			if event.Data.Duration != nil {
				lg.Info("Turn ended", "turn", turnCounter, "duration_ms", *event.Data.Duration)
			}
		case copilot.SessionEventTypeAssistantReasoning:
			// Content already captured above
		case copilot.SessionEventTypeAssistantIntent:
			if event.Data.Intent != nil {
				rec.Intent = *event.Data.Intent
			}
		case copilot.SessionEventTypeAssistantUsage:
			if event.Data.InputTokens != nil {
				rec.InputTokens = int(*event.Data.InputTokens)
			}
			if event.Data.OutputTokens != nil {
				rec.OutputTokens = int(*event.Data.OutputTokens)
			}
		case copilot.SessionEventTypeSessionWorkspaceFileChanged:
			if event.Data.Operation != nil {
				rec.FileOperation = string(*event.Data.Operation)
			}
		case copilot.SessionEventTypeCommandExecute:
			if event.Data.Command != nil {
				rec.CommandText = *event.Data.Command
			}
		case copilot.SessionEventTypeCommandCompleted:
			if event.Data.Command != nil {
				rec.CommandText = *event.Data.Command
			}
		case copilot.SessionEventTypeSkillInvoked:
			if event.Data.Name != nil {
				rec.SkillName = *event.Data.Name
			}
		case copilot.SessionEventTypeExternalToolRequested, copilot.SessionEventTypeExternalToolCompleted:
			if event.Data.ToolName != nil {
				rec.ToolName = *event.Data.ToolName
			}
		case copilot.SessionEventTypeSessionTruncation:
			rec.IsTruncation = true
			lg.Warn("Context truncated")
		case copilot.SessionEventTypeSessionCompactionStart:
			lg.Info("Context compaction started")
		case copilot.SessionEventTypeSessionCompactionComplete:
			lg.Info("Context compaction complete")
		case copilot.SessionEventTypeSessionWarning:
			if event.Data.Message != nil {
				rec.WarningText = *event.Data.Message
				lg.Warn("Session warning", "message", *event.Data.Message)
			}
		case copilot.SessionEventTypeAbort:
			lg.Error("Session aborted")
		case copilot.SessionEventTypePermissionRequested:
			// Audit trail
		case copilot.SessionEventTypePermissionCompleted:
			// Audit trail
		case copilot.SessionEventTypeSessionSkillsLoaded:
			if len(event.Data.Skills) > 0 {
				names := make([]string, 0, len(event.Data.Skills))
				for _, s := range event.Data.Skills {
					names = append(names, s.Name)
				}
				rec.Content = strings.Join(names, ", ")
				lg.Info("Skills loaded", "skills", rec.Content)
			}
		case copilot.SessionEventTypeSessionMcpServersLoaded:
			if len(event.Data.Servers) > 0 {
				names := make([]string, 0, len(event.Data.Servers))
				for _, s := range event.Data.Servers {
					names = append(names, s.Name)
				}
				rec.Content = strings.Join(names, ", ")
				lg.Info("MCP servers loaded", "servers", rec.Content)
			}
		case copilot.SessionEventTypeSessionToolsUpdated:
			lg.Info("Tools updated")
		case copilot.SessionEventTypeSubagentCompleted:
			if event.Data.ToolCallID != nil {
				rec.SubagentID = *event.Data.ToolCallID
			}
		case copilot.SessionEventTypeSubagentFailed:
			if event.Data.ToolCallID != nil {
				rec.SubagentID = *event.Data.ToolCallID
			}
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
			case copilot.SessionEventTypeAssistantTurnStart:
				e.progressFn(progress.ProgressEvent{
					EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
					Type:    progress.EventReasoning,
					Message: fmt.Sprintf("Turn %d started", turnCounter),
				})
			case copilot.SessionEventTypeSessionTruncation:
				e.progressFn(progress.ProgressEvent{
					EvalID: evalID, PromptID: p.ID, ConfigName: cfg.Name,
					Type:    progress.EventReasoning,
					Message: "⚠ Context truncated",
				})
			}
		}

		// Debug logging — all event types (slog.Debug is a no-op at higher levels)
		switch event.Type {
		case copilot.SessionEventTypeToolExecutionStart:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			lg.Debug("Tool start", "tool", toolName)
		case copilot.SessionEventTypeToolExecutionComplete:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			content := ""
			if event.Data.Content != nil {
				content = truncateStr(*event.Data.Content, 200)
			}
			lg.Debug("Tool done", "tool", toolName, "result", content)
		case copilot.SessionEventTypeAssistantMessage:
			content := ""
			if event.Data.Content != nil {
				content = *event.Data.Content
			}
			if content != "" {
				if summary := detectFileCreation(content); summary != "" {
					lg.Debug("Assistant creating file", "summary", summary)
				} else {
					lg.Debug("Assistant message", "content", truncateStr(content, 200))
				}
			}
		case copilot.SessionEventTypeSessionError:
			content := ""
			if event.Data.Content != nil {
				content = *event.Data.Content
			}
			lg.Debug("Session error", "content", content)
		case copilot.SessionEventTypeAssistantTurnStart:
			lg.Debug("Turn started", "turn", turnCounter)
		case copilot.SessionEventTypeAssistantTurnEnd:
			lg.Debug("Turn ended", "turn", turnCounter)
		case copilot.SessionEventTypeAssistantUsage:
			in, out := 0, 0
			if event.Data.InputTokens != nil {
				in = int(*event.Data.InputTokens)
			}
			if event.Data.OutputTokens != nil {
				out = int(*event.Data.OutputTokens)
			}
			lg.Debug("Token usage", "input_tokens", in, "output_tokens", out)
		case copilot.SessionEventTypeSessionTruncation:
			lg.Debug("Context truncated")
		case copilot.SessionEventTypeSkillInvoked:
			name := ""
			if event.Data.Name != nil {
				name = *event.Data.Name
			}
			lg.Debug("Skill invoked", "skill", name)
		case copilot.SessionEventTypeSubagentCompleted, copilot.SessionEventTypeSubagentFailed:
			lg.Debug("Subagent event", "type", string(event.Type))
		default:
			content := ""
			if event.Data.Content != nil {
				content = truncateStr(*event.Data.Content, 100)
			}
			lg.Debug("SDK event", "type", string(event.Type), "content", content)
		}
	}

	session, err := client.CreateSession(genCtx, sessionCfg)
	if err != nil {
		return &EvalResult{
			Error:        fmt.Sprintf("session creation failed: %v", err),
			ErrorDetails: err.Error(),
		}, fmt.Errorf("creating session: %w", err)
	}
	sessionID = session.SessionID

	// Send the prompt
	if e.progressFn != nil {
		e.progressFn(progress.ProgressEvent{
			EvalID: debugPrefix, PromptID: p.ID, ConfigName: cfg.Name,
			Type:    progress.EventSendingPrompt,
			Message: fmt.Sprintf("Sending prompt (%d chars)...", len(p.PromptText)),
		})
	}
	lg.Debug("Sending prompt", "chars", len(p.PromptText))
	_, err = session.SendAndWait(genCtx, copilot.MessageOptions{
		Prompt: p.PromptText,
	})
	if err != nil {
		mu.Lock()
		captured := make([]report.SessionEventRecord, len(sessionRecords))
		copy(captured, sessionRecords)
		capturedEvts := make([]copilot.SessionEvent, len(events))
		copy(capturedEvts, events)
		mu.Unlock()

		// Mid-generation turn limit (#69): return partial results so the
		// post-generation guardrail in engine.go can mark the eval as failed
		// with a proper reason instead of treating it as an SDK error.
		if turnLimitHit {
			generatedFiles, _ := listFiles(workDir)
			lg.Warn("Returning partial results after turn-limit cancellation",
				"turns", turnCounter, "files", len(generatedFiles))
			return &EvalResult{
				GeneratedFiles: generatedFiles,
				EventCount:     len(captured),
				ToolCalls:      extractToolCalls(capturedEvts),
				SessionEvents:  captured,
				Success:        true, // Let engine.go guardrail set the proper failure
				StarterFiles:   starterFiles,
			}, nil
		}

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

	lg.Debug("Session results",
		"events", len(capturedEvents),
		"tool_calls", len(toolCalls),
		"files", len(generatedFiles))

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

func (e *CopilotSDKEvaluator) buildSessionConfig(cfg *config.ToolConfig, workDir string, configDir string) *copilot.SessionConfig {
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

	// Safety boundaries (#36): prevent real Azure resource provisioning unless --allow-cloud is set.
	if !e.allowCloud {
		systemMsg += "\n\nSAFETY BOUNDARIES:\n" +
			"11. Do NOT provision real Azure resources. Do NOT run `az` CLI commands that create, update, or delete resources.\n" +
			"12. Do NOT use `az group create`, `az storage account create`, `az webapp create`, or any destructive Azure CLI commands.\n" +
			"13. Use mock data, environment variables, or local emulators (e.g., Azurite for Storage, CosmosDB emulator) for connection strings.\n" +
			"14. Generate code that can run locally without cloud dependencies. Use placeholder values like `os.Getenv(\"AZURE_STORAGE_CONNECTION_STRING\")` for configuration.\n" +
			"15. If the prompt asks for infrastructure provisioning, generate Bicep/ARM templates or Terraform files instead of running live commands."
	}

	sc := &copilot.SessionConfig{
		Model: cfg.EffectiveModel(),
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: systemMsg,
		},
		ConfigDir:           configDir,
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		Hooks: &copilot.SessionHooks{
			OnPreToolUse: func(input copilot.PreToolUseHookInput, invocation copilot.HookInvocation) (*copilot.PreToolUseHookOutput, error) {
				toolName := input.ToolName
				// Validate file paths for file-write tools
				if isFileWriteTool(toolName) {
					if args, ok := input.ToolArgs.(map[string]interface{}); ok {
						if p, ok := args["path"].(string); ok {
							if !strings.HasPrefix(p, workDir) {
								slog.Warn("File path outside workspace",
									"tool", toolName, "path", p, "workspace", workDir)
								return &copilot.PreToolUseHookOutput{
									PermissionDecision:       "deny",
									PermissionDecisionReason: fmt.Sprintf("path %q is outside workspace %q", p, workDir),
								}, nil
							}
						}
					}
				}
				// Log bash/command tools
				if toolName == "bash" || toolName == "shell" || toolName == "run_command" {
					if args, ok := input.ToolArgs.(map[string]interface{}); ok {
						if cmd, ok := args["command"].(string); ok {
							slog.Debug("Command execution", "tool", toolName, "command", truncateStr(cmd, 120))
						}
					}
				}
				return &copilot.PreToolUseHookOutput{}, nil
			},
			OnPostToolUse: func(input copilot.PostToolUseHookInput, invocation copilot.HookInvocation) (*copilot.PostToolUseHookOutput, error) {
				slog.Debug("Tool complete", "tool", input.ToolName)
				// Check file sizes for file operations
				if isFileWriteTool(input.ToolName) {
					if args, ok := input.ToolArgs.(map[string]interface{}); ok {
						if p, ok := args["path"].(string); ok {
							if info, err := os.Stat(p); err == nil && info.Size() > 100*1024 {
								slog.Warn("Large file created", "path", p, "bytes", info.Size())
							}
						}
					}
				}
				return &copilot.PostToolUseHookOutput{}, nil
			},
		},
		SkillDirectories:    skillDirs,
	}

	// Only set AvailableTools/ExcludedTools when non-empty.
	// An empty slice serializes as JSON [] which tells the CLI "zero tools" —
	// nil serializes as null which means "all default tools available."
	availableTools := cfg.EffectiveAvailableTools()
	excludedTools := cfg.EffectiveExcludedTools()
	if len(availableTools) > 0 {
		sc.AvailableTools = availableTools
	}
	if len(excludedTools) > 0 {
		sc.ExcludedTools = excludedTools
	}

	// Map MCP servers
	mcpServers := cfg.EffectiveMCPServers()
	if len(mcpServers) > 0 {
		sc.MCPServers = make(map[string]copilot.MCPServerConfig, len(mcpServers))
		for name, srv := range mcpServers {
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
