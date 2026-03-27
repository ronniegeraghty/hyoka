package eval

import (
"context"
"fmt"
"os"
"path/filepath"
"strings"
"testing"
"time"

"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/config"
"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/prompt"
)

// slowEvaluator blocks until context cancellation, simulating a timeout.
type slowEvaluator struct{}

func (s *slowEvaluator) Evaluate(ctx context.Context, _ *prompt.Prompt, _ *config.ToolConfig, _ string) (*EvalResult, error) {
	<-ctx.Done()
	return nil, fmt.Errorf("prompt send failed: %w", ctx.Err())
}

func TestStubEvaluator(t *testing.T) {
stub := &StubEvaluator{}
p := &prompt.Prompt{ID: "test-prompt", Language: "go"}
cfg := &config.ToolConfig{Name: "test-config", Model: "gpt-4"}

result, err := stub.Evaluate(context.Background(), p, cfg, t.TempDir())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !result.Success {
t.Error("expected stub to succeed")
}
if len(result.GeneratedFiles) == 0 {
t.Error("expected stub to return generated files")
}
if !result.IsStub {
t.Error("expected IsStub to be true for stub evaluator")
}
}

func TestStubVerifier(t *testing.T) {
stub := &StubVerifier{}
result, err := stub.Verify(context.Background(), "test prompt", "/tmp/test", "")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !result.Pass {
t.Error("expected stub verifier to pass")
}
}

func TestEngineDryRun(t *testing.T) {
engine := NewEngine(&StubEvaluator{}, EngineOptions{
Workers: 2,
DryRun:  true,
})

prompts := []*prompt.Prompt{
{ID: "p1", Service: "storage", Language: "dotnet"},
{ID: "p2", Service: "keyvault", Language: "python"},
}
configs := []config.ToolConfig{
{Name: "baseline", Model: "gpt-4"},
{Name: "azure-mcp", Model: "claude-sonnet-4.5"},
}

summary, err := engine.Run(context.Background(), prompts, configs)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if summary.RunID != "dry-run" {
t.Errorf("expected run ID 'dry-run', got %q", summary.RunID)
}
if summary.TotalEvals != 4 {
t.Errorf("expected 4 evaluations (2 prompts x 2 configs), got %d", summary.TotalEvals)
}
if summary.TotalPrompts != 2 {
t.Errorf("expected 2 prompts, got %d", summary.TotalPrompts)
}
if summary.TotalConfigs != 2 {
t.Errorf("expected 2 configs, got %d", summary.TotalConfigs)
}
}

func TestEngineRun(t *testing.T) {
outputDir := t.TempDir()
engine := NewEngine(&StubEvaluator{}, EngineOptions{
Workers:   1,
Timeout:   30 * time.Second,
OutputDir: outputDir,
})

prompts := []*prompt.Prompt{
{ID: "test-prompt", Service: "storage", Plane: "data-plane", Language: "go", Category: "auth"},
}
configs := []config.ToolConfig{
{Name: "test-config", Model: "gpt-4"},
}

summary, err := engine.Run(context.Background(), prompts, configs)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if summary.TotalEvals != 1 {
t.Errorf("expected 1 evaluation, got %d", summary.TotalEvals)
}
}

func TestEngineRunCapturesGeneratedFiles(t *testing.T) {
	// The evaluator returns GeneratedFiles in its result, but may not leave
	// files on disk (e.g., SDK cleanup removes them). The engine must use the
	// evaluator's captured list rather than relying solely on ws.ListFiles().
	outputDir := t.TempDir()
	engine := NewEngine(&StubEvaluator{}, EngineOptions{
		Workers:   1,
		Timeout:   30 * time.Second,
		OutputDir: outputDir,
	})

	prompts := []*prompt.Prompt{
		{ID: "filelist-test", Service: "storage", Plane: "data-plane", Language: "python", Category: "crud"},
	}
	configs := []config.ToolConfig{
		{Name: "baseline", Model: "gpt-4"},
	}

	summary, err := engine.Run(context.Background(), prompts, configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(summary.Results))
	}
	r := summary.Results[0]
	if len(r.GeneratedFiles) == 0 {
		t.Error("expected GeneratedFiles to be populated from evaluator result, got 0 files")
	}
}

func TestEngineRunTimeoutError(t *testing.T) {
	// An evaluator that blocks until the context is cancelled.
	slowEval := &slowEvaluator{}
	outputDir := t.TempDir()
	engine := NewEngine(slowEval, EngineOptions{
		Workers:   1,
		Timeout:   100 * time.Millisecond,
		OutputDir: outputDir,
	})

	prompts := []*prompt.Prompt{
		{ID: "timeout-test", Service: "storage", Plane: "data-plane", Language: "go", Category: "auth"},
	}
	configs := []config.ToolConfig{
		{Name: "baseline", Model: "gpt-4"},
	}

	summary, err := engine.Run(context.Background(), prompts, configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(summary.Results))
	}
	r := summary.Results[0]
	if r.Error == "" {
		t.Fatal("expected error in report for timed-out eval")
	}
	if !strings.Contains(r.Error, "timed out") {
		t.Errorf("expected timeout message in error, got %q", r.Error)
	}
}

func TestNewWorkspace(t *testing.T) {
ws, err := NewWorkspace("test-prompt", "test-config")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if ws.Dir == "" {
t.Error("expected non-empty workspace dir")
}
// Verify directory exists
info, err := os.Stat(ws.Dir)
if err != nil {
t.Fatalf("workspace dir does not exist: %v", err)
}
if !info.IsDir() {
t.Error("expected workspace to be a directory")
}

// Verify it's in temp dir, not in reports
if !strings.HasPrefix(ws.Dir, os.TempDir()) {
t.Errorf("expected workspace in temp dir, got %s", ws.Dir)
}

// Test ListFiles on empty workspace
files, err := ws.ListFiles()
if err != nil {
t.Fatalf("ListFiles failed: %v", err)
}
if len(files) != 0 {
t.Errorf("expected 0 files in empty workspace, got %d", len(files))
}

// Create a test file and verify ListFiles
testFile := filepath.Join(ws.Dir, "test.py")
if err := os.WriteFile(testFile, []byte("print('hello')"), 0644); err != nil {
t.Fatalf("failed to write test file: %v", err)
}
files, err = ws.ListFiles()
if err != nil {
t.Fatalf("ListFiles failed: %v", err)
}
if len(files) != 1 || files[0] != "test.py" {
t.Errorf("expected [test.py], got %v", files)
}

// Test CopyFilesTo
destDir := t.TempDir()
copied, err := ws.CopyFilesTo(destDir)
if err != nil {
t.Fatalf("CopyFilesTo failed: %v", err)
}
if len(copied) != 1 || copied[0] != "test.py" {
t.Errorf("expected [test.py] copied, got %v", copied)
}
destFile := filepath.Join(destDir, "test.py")
data, err := os.ReadFile(destFile)
if err != nil {
t.Fatalf("failed to read copied file: %v", err)
}
if string(data) != "print('hello')" {
t.Errorf("unexpected file content: %s", data)
}

// Cleanup
if err := ws.Cleanup(); err != nil {
t.Fatalf("cleanup failed: %v", err)
}
if _, err := os.Stat(ws.Dir); !os.IsNotExist(err) {
t.Error("expected workspace to be removed after cleanup")
}
}
