// Package report handles generation of JSON, HTML, and Markdown reports.
package report

import (
	"github.com/ronniegeraghty/hyoka/internal/build"
	"github.com/ronniegeraghty/hyoka/internal/review"
)

// SessionEventRecord is a serializable representation of a Copilot session event.
type SessionEventRecord struct {
	Type          string  `json:"type"`
	ToolName      string  `json:"tool_name,omitempty"`
	ToolArgs      string  `json:"tool_args,omitempty"`
	Content       string  `json:"content,omitempty"`
	Error         string  `json:"error,omitempty"`
	ToolResult    string  `json:"tool_result,omitempty"`
	ToolSuccess   *bool   `json:"tool_success,omitempty"`
	Duration      float64 `json:"duration_ms,omitempty"`
	MCPServerName string  `json:"mcp_server_name,omitempty"`
	MCPToolName   string  `json:"mcp_tool_name,omitempty"`
	FilePath      string  `json:"file_path,omitempty"`
	// Expanded event fields
	TurnNumber    int    `json:"turnNumber,omitempty"`
	InputTokens   int    `json:"inputTokens,omitempty"`
	OutputTokens  int    `json:"outputTokens,omitempty"`
	SkillName     string `json:"skillName,omitempty"`
	CommandText   string `json:"commandText,omitempty"`
	FileOperation string `json:"fileOperation,omitempty"`
	FileSize      int64  `json:"fileSize,omitempty"`
	SubagentID    string `json:"subagentId,omitempty"`
	IsTruncation  bool   `json:"isTruncation,omitempty"`
	Intent        string `json:"intent,omitempty"`
	WarningText   string `json:"warningText,omitempty"`
}

// VerifyResult holds the outcome of Copilot-based code verification.
type VerifyResult struct {
	Pass      bool   `json:"pass"`
	Reasoning string `json:"reasoning"`
	Summary   string `json:"summary"`
}

// ToolUsageResult holds the comparison of expected vs actual tool usage.
type ToolUsageResult struct {
	ExpectedTools []string `json:"expected_tools"`
	ActualTools   []string `json:"actual_tools"`
	MatchedTools  []string `json:"matched_tools"`
	MissingTools  []string `json:"missing_tools"`
	ExtraTools    []string `json:"extra_tools"`
	Match         bool     `json:"tool_usage_match"`
}

// ReviewedFile holds an annotated code file with inline review comments.
type ReviewedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// EnvironmentInfo captures session environment and configuration metadata.
type EnvironmentInfo struct {
	Model              string   `json:"model"`
	SkillDirectories   []string `json:"skillDirectories,omitempty"`
	SkillsInvoked      []string `json:"skillsInvoked,omitempty"`
	SkillsLoaded       []string `json:"skillsLoaded,omitempty"`
	AvailableTools     []string `json:"availableTools,omitempty"`
	ExcludedTools      []string `json:"excludedTools,omitempty"`
	MCPServers         []string `json:"mcpServers,omitempty"`
	SafetyBoundaries   bool     `json:"safetyBoundaries"`
	AllowCloud         bool     `json:"allowCloud"`
	WorkingDirectory   string   `json:"workingDirectory"`
	TotalInputTokens   int      `json:"totalInputTokens,omitempty"`
	TotalOutputTokens  int      `json:"totalOutputTokens,omitempty"`
	TurnCount          int      `json:"turnCount,omitempty"`
	ContextTruncated   bool     `json:"contextTruncated,omitempty"`
}

// EvalReport contains the results of a single prompt evaluation.
type EvalReport struct {
	PromptID       string               `json:"prompt_id"`
	ConfigName     string               `json:"config_name"`
	Timestamp      string               `json:"timestamp"`
	Duration       float64              `json:"duration_seconds"`
	PromptMeta     map[string]any       `json:"prompt_metadata"`
	ConfigUsed     map[string]any       `json:"config_used"`
	GeneratedFiles []string             `json:"generated_files"`
	StarterFiles   []string             `json:"starter_files,omitempty"`
	ReviewedFiles  []ReviewedFile       `json:"reviewed_files,omitempty"`
	Build          *build.BuildResult   `json:"build,omitempty"`
	Verification   *VerifyResult        `json:"verification,omitempty"`
	Review         *review.ReviewResult   `json:"review,omitempty"`
	ReviewPanel    []review.ReviewResult  `json:"review_panel,omitempty"`
	ToolUsage      *ToolUsageResult     `json:"tool_usage,omitempty"`
	SessionEvents  []SessionEventRecord `json:"session_events,omitempty"`
	EventCount     int                  `json:"event_count"`
	ToolCalls      []string             `json:"tool_calls"`
	Environment    *EnvironmentInfo     `json:"environment,omitempty"`
	Success        bool                 `json:"success"`
	Error          string               `json:"error,omitempty"`
	ErrorDetails   string               `json:"error_details,omitempty"`
	ErrorCategory  string               `json:"error_category,omitempty"`  // timeout, sdk_error, generation_failure, review_failure, no_files
	FailureReason  string               `json:"failure_reason,omitempty"` // human-readable explanation of failure
	IsStub         bool                 `json:"is_stub,omitempty"`
	RerunCommand   string               `json:"rerunCommand,omitempty"`
	// Generator guardrails (#35)
	GuardrailMaxTurns      int    `json:"guardrail_max_turns,omitempty"`
	GuardrailMaxFiles      int    `json:"guardrail_max_files,omitempty"`
	GuardrailMaxOutputSize int64  `json:"guardrail_max_output_size,omitempty"`
	GuardrailAbortReason   string `json:"guardrail_abort_reason,omitempty"`
}

// RunSummary contains aggregate statistics for an evaluation run.
type RunSummary struct {
	RunID        string        `json:"run_id"`
	Timestamp    string        `json:"timestamp"`
	TotalPrompts int           `json:"total_prompts"`
	TotalConfigs int           `json:"total_configs"`
	TotalEvals   int           `json:"total_evaluations"`
	Passed       int           `json:"passed"`
	Failed       int           `json:"failed"`
	Errors       int           `json:"errors"`
	Duration     float64       `json:"duration_seconds"`
	Reports      []string      `json:"report_paths"`
	Results      []*EvalReport `json:"results,omitempty"`
	Analysis     string        `json:"analysis,omitempty"`
}
