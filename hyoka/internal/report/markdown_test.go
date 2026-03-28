package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronniegeraghty/hyoka/internal/build"
	"github.com/ronniegeraghty/hyoka/internal/review"
)

func TestWriteMarkdownReport(t *testing.T) {
	dir := t.TempDir()

	boolTrue := true
	r := &EvalReport{
		PromptID:   "test-prompt",
		ConfigName: "baseline",
		Timestamp:  "2024-01-15T10:00:00Z",
		Duration:   12.5,
		PromptMeta: map[string]any{"service": "storage", "language": "dotnet"},
		ConfigUsed: map[string]any{"name": "baseline", "model": "gpt-4"},
		GeneratedFiles: []string{"Program.cs"},
		Build: &build.BuildResult{
			Language: "dotnet",
			Command:  "dotnet build",
			ExitCode: 0,
			Success:  true,
			Stdout:   "Build succeeded.",
		},
		Verification: &VerifyResult{
			Pass:      true,
			Reasoning: "Code correctly implements storage auth",
			Summary:   "All requirements met",
		},
		Review: &review.ReviewResult{
			Scores: review.ReviewScores{
				Criteria: []review.CriterionResult{
					{Name: "Code Builds", Passed: true, Reason: "Compiles successfully"},
					{Name: "Latest Packages", Passed: true, Reason: "Using latest versions"},
					{Name: "Best Practices", Passed: true, Reason: "Uses DefaultAzureCredential"},
					{Name: "Error Handling", Passed: false, Reason: "Missing retry logic"},
					{Name: "Code Quality", Passed: true, Reason: "Clean structure"},
				},
			},
			OverallScore: 4,
			MaxScore:     5,
			Summary:      "Good implementation",
			Issues:       []string{"Missing retry logic"},
			Strengths:    []string{"Clean code structure"},
		},
		SessionEvents: []SessionEventRecord{
			{Type: "user.message", Content: "Write a dotnet storage auth sample"},
			{Type: "assistant.reasoning", Content: "I need to create an auth sample"},
			{Type: "tool.execution_start", ToolName: "create", ToolArgs: `{"path":"Program.cs"}`},
			{Type: "tool.execution_complete", ToolName: "create", ToolResult: "File created", ToolSuccess: &boolTrue, Duration: 150.5},
			{Type: "assistant.message", Content: "Here is your sample"},
		},
		EventCount: 15,
		ToolCalls:  []string{"create_file", "edit_file"},
		Success:    true,
	}

	reportPath, err := WriteMarkdownReport(r, dir, "20240115-100000", "storage", "data-plane", "dotnet", "authentication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	content := string(data)
	checks := []string{
		"# Evaluation Report: test-prompt",
		"baseline",
		"✅ PASSED",
		"4/5",
		"Code Builds",
		"Good implementation",
		"Program.cs",
		"dotnet build",
		"Write a dotnet storage auth sample",
		"I need to create an auth sample",
		"Code Review",
		"Clean code structure",
		"Missing retry logic",
		"Tool Calls",
		"Back to Summary",
		"File created",
		"150ms",
		"Verification",
		"All requirements met",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Markdown report missing %q", check)
		}
	}

	expectedDir := filepath.Join(dir, "20240115-100000", "results", "storage", "data-plane", "dotnet", "authentication", "test-prompt", "baseline")
	if _, err := os.Stat(expectedDir); err != nil {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestWriteMarkdownReportFailed(t *testing.T) {
	dir := t.TempDir()

	r := &EvalReport{
		PromptID:       "test-prompt",
		ConfigName:     "baseline",
		Timestamp:      "2024-01-15T10:00:00Z",
		Duration:       5.0,
		PromptMeta:     map[string]any{},
		ConfigUsed:     map[string]any{},
		GeneratedFiles: []string{},
		Success:        false,
		Error:          "timeout exceeded",
	}

	reportPath, err := WriteMarkdownReport(r, dir, "run1", "svc", "plane", "lang", "cat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "FAILED") {
		t.Error("expected FAILED in report")
	}
	if !strings.Contains(content, "timeout exceeded") {
		t.Error("expected error message in report")
	}
}

func TestWriteSummaryMarkdown(t *testing.T) {
	dir := t.TempDir()

	s := &RunSummary{
		RunID:        "20240115-100000",
		Timestamp:    "2024-01-15T10:00:00Z",
		TotalPrompts: 2,
		TotalConfigs: 2,
		TotalEvals:   4,
		Passed:       3,
		Failed:       1,
		Errors:       0,
		Duration:     120.5,
		Results: []*EvalReport{
			{
				PromptID:   "prompt-a",
				ConfigName: "baseline",
				Success:    true,
				Duration:   10.0,
				PromptMeta: map[string]any{"service": "storage", "plane": "data-plane", "language": "dotnet", "category": "auth"},
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 4, MaxScore: 5},
			},
			{
				PromptID:   "prompt-a",
				ConfigName: "azure-mcp",
				Success:    true,
				Duration:   15.0,
				PromptMeta: map[string]any{"service": "storage", "plane": "data-plane", "language": "dotnet", "category": "auth"},
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 5, MaxScore: 5},
			},
			{
				PromptID:   "prompt-b",
				ConfigName: "baseline",
				Success:    false,
				Duration:   5.0,
				PromptMeta: map[string]any{},
				Build:      &build.BuildResult{Success: false},
			},
			{
				PromptID:   "prompt-b",
				ConfigName: "azure-mcp",
				Success:    true,
				Duration:   12.0,
				PromptMeta: map[string]any{},
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 3, MaxScore: 5},
			},
		},
	}

	summaryPath, err := WriteSummaryMarkdown(s, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("failed to read summary: %v", err)
	}

	content := string(data)
	checks := []string{
		"Evaluation Summary",
		"20240115-100000",
		"prompt-a",
		"prompt-b",
		"baseline",
		"azure-mcp",
		"4/5",
		"5/5",
		"3/5",
		"Comparison Matrix",
		"Detailed Results",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("summary markdown missing %q", check)
		}
	}

	// Verify pass/fail icons
	passCount := strings.Count(content, "✅")
	failCount := strings.Count(content, "❌")
	// Matrix: 3 pass + 1 fail, Detail table: 3 pass + 1 fail = 6 pass, 2 fail
	if passCount != 6 {
		t.Errorf("expected 6 ✅ icons (3 matrix + 3 detail), got %d", passCount)
	}
	if failCount != 2 {
		t.Errorf("expected 2 ❌ icons (1 matrix + 1 detail), got %d", failCount)
	}
}

func TestWriteMarkdownReportStub(t *testing.T) {
	dir := t.TempDir()

	r := &EvalReport{
		PromptID:       "test-prompt",
		ConfigName:     "baseline",
		Timestamp:      "2024-01-15T10:00:00Z",
		Duration:       1.0,
		PromptMeta:     map[string]any{},
		ConfigUsed:     map[string]any{},
		GeneratedFiles: []string{},
		Success:        true,
		IsStub:         true,
	}

	reportPath, err := WriteMarkdownReport(r, dir, "run1", "svc", "plane", "lang", "cat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Stub") {
		t.Error("expected Stub indicator in report")
	}
}

func TestTruncateStr(t *testing.T) {
	short := "hello"
	if truncateStr(short, 10) != "hello" {
		t.Error("short string should not be truncated")
	}

	long := strings.Repeat("a", 100)
	result := truncateStr(long, 50)
	if len(result) <= 50 {
		t.Error("truncated string should have suffix")
	}
	if !strings.Contains(result, "truncated") {
		t.Error("truncated string should contain truncation marker")
	}
}
