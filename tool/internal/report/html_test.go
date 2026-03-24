package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/build"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/review"
)

func TestWriteHTMLReport(t *testing.T) {
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

	reportPath, err := WriteHTMLReport(r, dir, "20240115-100000", "storage", "data-plane", "dotnet", "authentication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}

	content := string(data)
	checks := []string{
		"test-prompt",
		"baseline",
		"PASSED",
		"4/5",
		"Code Builds",
		"Good implementation",
		"Program.cs",
		"dotnet build",
		"Generation Timeline",
		"Write a dotnet storage auth sample",
		"I need to create an auth sample",
		"Code Review",
		"Clean code structure",
		"Missing retry logic",
		"Tool call: create",
		"Back to Summary",
		"File created",
		"150ms",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("HTML report missing %q", check)
		}
	}

	expectedDir := filepath.Join(dir, "20240115-100000", "results", "storage", "data-plane", "dotnet", "authentication", "baseline")
	if _, err := os.Stat(expectedDir); err != nil {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestWriteHTMLReportNoReview(t *testing.T) {
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

	reportPath, err := WriteHTMLReport(r, dir, "run1", "svc", "plane", "lang", "cat")
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

func TestWriteSummaryHTML(t *testing.T) {
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
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 4, MaxScore: 5},
			},
			{
				PromptID:   "prompt-a",
				ConfigName: "azure-mcp",
				Success:    true,
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 5, MaxScore: 5},
			},
			{
				PromptID:   "prompt-b",
				ConfigName: "baseline",
				Success:    false,
				Build:      &build.BuildResult{Success: false},
			},
			{
				PromptID:   "prompt-b",
				ConfigName: "azure-mcp",
				Success:    true,
				Build:      &build.BuildResult{Success: true},
				Review:     &review.ReviewResult{OverallScore: 3, MaxScore: 5},
			},
		},
	}

	summaryPath, err := WriteSummaryHTML(s, dir)
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
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("summary HTML missing %q", check)
		}
	}

	// Verify the summary uses Success field: 3 passed should show ✅, 1 failed should show ❌
	// Matrix has 3 pass + 1 fail, detailed results table also has 3 pass + 1 fail
	passCount := strings.Count(content, "✅")
	failCount := strings.Count(content, "❌")
	if passCount != 6 {
		t.Errorf("expected 6 ✅ icons (3 matrix + 3 detail), got %d", passCount)
	}
	if failCount != 2 {
		t.Errorf("expected 2 ❌ icons (1 matrix + 1 detail), got %d", failCount)
	}
}

func TestWriteSummaryHTMLNoBuild(t *testing.T) {
	dir := t.TempDir()

	// Simulate Copilot-verified results (no Build, only Verification)
	s := &RunSummary{
		RunID:      "20240201-090000",
		Timestamp:  "2024-02-01T09:00:00Z",
		TotalEvals: 3,
		Passed:     2,
		Failed:     1,
		Duration:   60.0,
		Results: []*EvalReport{
			{PromptID: "p1", ConfigName: "baseline", Success: true, Verification: &VerifyResult{Pass: true}},
			{PromptID: "p1", ConfigName: "mcp", Success: true, Verification: &VerifyResult{Pass: true}, Review: &review.ReviewResult{OverallScore: 3, MaxScore: 5}},
			{PromptID: "p2", ConfigName: "baseline", Success: false, Verification: &VerifyResult{Pass: false}},
		},
	}

	summaryPath, err := WriteSummaryHTML(s, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("failed to read summary: %v", err)
	}

	content := string(data)
	// Matrix has 2 pass + 1 fail, detailed results table also has 2 pass + 1 fail
	passCount := strings.Count(content, "✅")
	failCount := strings.Count(content, "❌")
	if passCount != 4 {
		t.Errorf("expected 4 ✅ (2 matrix + 2 detail), got %d", passCount)
	}
	if failCount != 2 {
		t.Errorf("expected 2 ❌ (1 matrix + 1 detail), got %d", failCount)
	}
}

func TestBuildMatrix(t *testing.T) {
	s := &RunSummary{
		Results: []*EvalReport{
			{PromptID: "p1", ConfigName: "c1", Success: true, Build: &build.BuildResult{Success: true}, Review: &review.ReviewResult{OverallScore: 4, MaxScore: 5}},
			{PromptID: "p1", ConfigName: "c2", Success: false, Build: &build.BuildResult{Success: false}},
			{PromptID: "p2", ConfigName: "c1", Error: "timeout"},
		},
	}

	m := buildMatrix(s)

	if len(m.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(m.Prompts))
	}
	if len(m.Configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(m.Configs))
	}

	cell := m.Cells["p1"]["c1"]
	if cell == nil {
		t.Fatal("expected cell for p1/c1")
	}
	if cell.Score != 4 {
		t.Errorf("expected score 4, got %d", cell.Score)
	}
	if !cell.BuildPass {
		t.Error("expected build pass")
	}
	if !cell.Success {
		t.Error("expected Success=true for p1/c1")
	}

	failCell := m.Cells["p1"]["c2"]
	if failCell == nil {
		t.Fatal("expected cell for p1/c2")
	}
	if failCell.Success {
		t.Error("expected Success=false for p1/c2")
	}

	errCell := m.Cells["p2"]["c1"]
	if errCell == nil {
		t.Fatal("expected cell for p2/c1")
	}
	if errCell.Error != "timeout" {
		t.Errorf("expected timeout error, got %q", errCell.Error)
	}
}

func TestBuildReportData(t *testing.T) {
	boolTrue := true
	r := &EvalReport{
		PromptID:       "test-prompt",
		GeneratedFiles: []string{"main.py", "requirements.txt"},
		SessionEvents: []SessionEventRecord{
			{Type: "session.start"},
			{Type: "user.message", Content: "Write a Python script"},
			{Type: "assistant.reasoning", Content: "I should create a script"},
			{Type: "tool.execution_start", ToolName: "create", ToolArgs: `{"path":"main.py"}`, MCPServerName: "fs-server"},
			{Type: "tool.execution_complete", ToolName: "create", ToolResult: "File created successfully", ToolSuccess: &boolTrue, Duration: 42.5},
			{Type: "tool.execution_start", ToolName: "create", ToolArgs: `{"path":"requirements.txt"}`},
			{Type: "tool.execution_complete", ToolName: "create", ToolResult: "File created successfully", ToolSuccess: &boolTrue, Duration: 10.0},
			{Type: "assistant.message", Content: "Done! Here are your files."},
		},
	}

	d := buildReportData(r)

	if d.Prompt != "Write a Python script" {
		t.Errorf("expected prompt from user.message, got %q", d.Prompt)
	}
	if d.Reasoning != "I should create a script" {
		t.Errorf("expected reasoning, got %q", d.Reasoning)
	}
	if d.FinalReply != "Done! Here are your files." {
		t.Errorf("expected final reply, got %q", d.FinalReply)
	}
	if len(d.ToolActions) != 2 {
		t.Errorf("expected 2 tool actions, got %d", len(d.ToolActions))
	}
	if d.ToolActions[0].Index != 1 || d.ToolActions[0].ToolName != "create" {
		t.Errorf("unexpected first tool action: %+v", d.ToolActions[0])
	}
	if d.ToolActions[0].Args != `{"path":"main.py"}` {
		t.Errorf("expected tool args, got %q", d.ToolActions[0].Args)
	}
	if d.ToolActions[0].MCPServer != "fs-server" {
		t.Errorf("expected MCP server 'fs-server', got %q", d.ToolActions[0].MCPServer)
	}
	if d.ToolActions[0].Result != "File created successfully" {
		t.Errorf("expected tool result from completion, got %q", d.ToolActions[0].Result)
	}
	if d.ToolActions[0].Success == nil || !*d.ToolActions[0].Success {
		t.Error("expected tool success=true")
	}
	if d.ToolActions[0].Duration != 42.5 {
		t.Errorf("expected duration 42.5, got %f", d.ToolActions[0].Duration)
	}
	if d.ToolActions[1].Result != "File created successfully" {
		t.Errorf("expected second tool result, got %q", d.ToolActions[1].Result)
	}
	if d.FileCount != 2 {
		t.Errorf("expected file count 2, got %d", d.FileCount)
	}

	// Verify timeline steps are built correctly
	// Expected: prompt, reasoning, tool_call, tool_call, message, complete = 6 steps
	if len(d.TimelineSteps) != 6 {
		t.Errorf("expected 6 timeline steps, got %d", len(d.TimelineSteps))
	}
	if len(d.TimelineSteps) >= 1 && d.TimelineSteps[0].StepType != "prompt" {
		t.Errorf("expected first step to be prompt, got %q", d.TimelineSteps[0].StepType)
	}
	if len(d.TimelineSteps) >= 2 && d.TimelineSteps[1].StepType != "reasoning" {
		t.Errorf("expected second step to be reasoning, got %q", d.TimelineSteps[1].StepType)
	}
	if len(d.TimelineSteps) >= 3 && d.TimelineSteps[2].StepType != "tool_call" {
		t.Errorf("expected third step to be tool_call, got %q", d.TimelineSteps[2].StepType)
	}
	if len(d.TimelineSteps) >= 3 && d.TimelineSteps[2].Duration != 42.5 {
		t.Errorf("expected tool_call duration 42.5, got %f", d.TimelineSteps[2].Duration)
	}
	if len(d.TimelineSteps) >= 6 && d.TimelineSteps[5].StepType != "complete" {
		t.Errorf("expected last step to be complete, got %q", d.TimelineSteps[5].StepType)
	}
}
