package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewDisplay_Disabled(t *testing.T) {
	var buf bytes.Buffer
	d := NewDisplay(DisplayConfig{
		Total:    5,
		Workers:  2,
		Writer:   &buf,
		Disabled: true,
	})
	d.HandleEvent(ProgressEvent{
		EvalID:     "test/config",
		PromptID:   "test",
		ConfigName: "config",
		Type:       EventStarting,
	})
	d.Done()
	if buf.Len() != 0 {
		t.Errorf("expected no output when disabled, got %d bytes", buf.Len())
	}
}

func TestDisplay_LineManagement(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines:     make([]evalLine, 0),
		lineIndex: make(map[string]int),
		total:     4,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
	}

	// Start two evals — should create lines 0 and 1
	d.HandleEvent(ProgressEvent{
		EvalID: "p1/c1", PromptID: "p1", ConfigName: "c1",
		Type: EventStarting,
	})
	d.HandleEvent(ProgressEvent{
		EvalID: "p2/c2", PromptID: "p2", ConfigName: "c2",
		Type: EventStarting,
	})

	if len(d.lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(d.lines))
	}
	if len(d.lineIndex) != 2 {
		t.Errorf("expected 2 active line entries, got %d", len(d.lineIndex))
	}

	// Complete first eval — line stays in list, marked completed
	d.HandleEvent(ProgressEvent{
		EvalID: "p1/c1", Type: EventPassed, FileCount: 3,
	})
	if d.completed != 1 || d.passed != 1 {
		t.Errorf("expected 1 completed/passed, got %d/%d", d.completed, d.passed)
	}
	if !d.lines[0].completed {
		t.Error("expected line 0 to be completed")
	}
	if len(d.lines) != 2 {
		t.Errorf("expected 2 lines (completed stays), got %d", len(d.lines))
	}

	// New eval should append as line 2 (not reuse old position)
	d.HandleEvent(ProgressEvent{
		EvalID: "p3/c3", PromptID: "p3", ConfigName: "c3",
		Type: EventStarting,
	})
	if len(d.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(d.lines))
	}
	if d.lineIndex["p3/c3"] != 2 {
		t.Errorf("expected new eval at line 2, got line %d", d.lineIndex["p3/c3"])
	}
}

func TestDisplay_EventIcons_PhaseAware(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines:     make([]evalLine, 0),
		lineIndex: make(map[string]int),
		total:     1,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
	}

	// Start eval
	d.HandleEvent(ProgressEvent{EvalID: "p/c", PromptID: "p", ConfigName: "c", Type: EventStarting})
	if d.lines[0].icon != "⏳" {
		t.Errorf("expected ⏳ after starting, got %q", d.lines[0].icon)
	}
	if d.lines[0].phase != PhaseGenerating {
		t.Errorf("expected phase generating after start, got %q", d.lines[0].phase)
	}

	// Tool start — should show ⚙
	d.HandleEvent(ProgressEvent{EvalID: "p/c", Type: EventToolStart, Message: "bash → ls"})
	if d.lines[0].icon != "⚙" {
		t.Errorf("expected ⚙ after tool start, got %q", d.lines[0].icon)
	}

	// Tool complete — Issue 1: should show ✓ and keep last action visible
	d.HandleEvent(ProgressEvent{EvalID: "p/c", Type: EventToolComplete, Message: "bash"})
	if d.lines[0].icon != "✓" {
		t.Errorf("expected ✓ after tool complete (Issue 1), got %q", d.lines[0].icon)
	}

	// Phase change to verifying
	d.HandleEvent(ProgressEvent{EvalID: "p/c", Type: EventPhaseChange, Phase: PhaseVerifying})
	if d.lines[0].phase != PhaseVerifying {
		t.Errorf("expected phase verifying, got %q", d.lines[0].phase)
	}
	if d.lines[0].icon != "🔍" {
		t.Errorf("expected 🔍 after verify phase change, got %q", d.lines[0].icon)
	}

	// Phase change to reviewing
	d.HandleEvent(ProgressEvent{EvalID: "p/c", Type: EventPhaseChange, Phase: PhaseReviewing})
	if d.lines[0].phase != PhaseReviewing {
		t.Errorf("expected phase reviewing, got %q", d.lines[0].phase)
	}
	if d.lines[0].icon != "📝" {
		t.Errorf("expected 📝 after review phase change, got %q", d.lines[0].icon)
	}

	// Other activity icons still work within a phase
	d.HandleEvent(ProgressEvent{EvalID: "p/c", Type: EventReasoning, Message: "Reasoning..."})
	if d.lines[0].icon != "💭" {
		t.Errorf("expected 💭 after reasoning, got %q", d.lines[0].icon)
	}
}

func TestDisplay_CompletedEvalsTracked(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines:     make([]evalLine, 0),
		lineIndex: make(map[string]int),
		total:     3,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
	}

	// Start 3 evals
	for i, id := range []string{"a/x", "b/y", "c/z"} {
		d.HandleEvent(ProgressEvent{
			EvalID:     id,
			PromptID:   string(rune('a' + i)),
			ConfigName: string(rune('x' + i)),
			Type:       EventStarting,
		})
	}

	d.HandleEvent(ProgressEvent{EvalID: "a/x", Type: EventPassed, FileCount: 2, ReviewScore: 8})
	d.HandleEvent(ProgressEvent{EvalID: "b/y", Type: EventFailed, Message: "verification failed"})
	d.HandleEvent(ProgressEvent{EvalID: "c/z", Type: EventError, Message: "timeout"})

	if d.passed != 1 {
		t.Errorf("expected 1 passed, got %d", d.passed)
	}
	if d.failed != 1 {
		t.Errorf("expected 1 failed, got %d", d.failed)
	}
	if d.errors != 1 {
		t.Errorf("expected 1 error, got %d", d.errors)
	}
	if d.completed != 3 {
		t.Errorf("expected 3 completed, got %d", d.completed)
	}
	// All completed evals tracked as lines
	completedCount := 0
	for _, l := range d.lines {
		if l.completed {
			completedCount++
		}
	}
	if completedCount != 3 {
		t.Errorf("expected 3 completed lines, got %d", completedCount)
	}
	// Verify review score is tracked
	if d.lines[0].reviewScore != 8 {
		t.Errorf("expected review score 8, got %d", d.lines[0].reviewScore)
	}
}

func TestDisplay_Finish(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines: []evalLine{
			{promptID: "crud", configName: "baseline", completed: true, passed: true, fileCount: 3, reviewScore: 8, duration: 34 * time.Second},
			{promptID: "crud", configName: "azure-mcp", completed: true, message: "verification failed", duration: 28 * time.Second},
		},
		lineIndex: make(map[string]int),
		total:     2,
		completed: 2,
		passed:    1,
		failed:    1,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
	}

	d.Finish()

	output := buf.String()
	// All completed evals show results
	if !strings.Contains(output, "PASSED") {
		t.Errorf("expected Finish output to contain 'PASSED', got %q", output)
	}
	if !strings.Contains(output, "FAILED") {
		t.Errorf("expected Finish output to contain 'FAILED', got %q", output)
	}
	if !strings.Contains(output, "3 files") {
		t.Errorf("expected Finish output to contain '3 files', got %q", output)
	}
	if !strings.Contains(output, "8/10") {
		t.Errorf("expected Finish output to contain '8/10' review score, got %q", output)
	}
	// Summary line
	if !strings.Contains(output, "Summary: 1/2 passed") {
		t.Errorf("expected Finish output to contain 'Summary: 1/2 passed', got %q", output)
	}
}

func TestDisplay_FinishWithReportDir(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines: []evalLine{
			{promptID: "p1", configName: "c1", completed: true, passed: true, fileCount: 2, duration: 10 * time.Second},
		},
		lineIndex: make(map[string]int),
		total:     1,
		completed: 1,
		passed:    1,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
		reportDir: "reports/20260321-171234/",
	}

	d.Finish()

	output := buf.String()
	if !strings.Contains(output, "Reports: reports/20260321-171234/") {
		t.Errorf("expected report dir in output, got %q", output)
	}
}

func TestFormatName_Short(t *testing.T) {
	d := &Display{width: 120}
	got := d.formatName("p1", "c1")
	if got != "p1/c1" {
		t.Errorf("expected 'p1/c1', got %q", got)
	}
}

func TestFormatName_Truncated(t *testing.T) {
	d := &Display{width: 120}
	got := d.formatName("very-long-prompt-id-that-exceeds-limit", "config-name")
	if len(got) > 38 {
		t.Errorf("expected name truncated to ≤38 chars, got %d: %q", len(got), got)
	}
	if !strings.HasSuffix(got, "..") {
		t.Errorf("expected truncated name to end with '..', got %q", got)
	}
}

func TestTermWidth_Default(t *testing.T) {
	w := TermWidth()
	if w <= 0 {
		t.Errorf("expected positive terminal width, got %d", w)
	}
}

func TestDisplay_ActivityTruncation(t *testing.T) {
	d := &Display{width: 80}
	maxW := d.activityWidth()
	longActivity := strings.Repeat("x", 200)
	truncated := d.truncateActivity(longActivity)
	if len(truncated) > maxW {
		t.Errorf("expected truncated activity ≤%d, got %d", maxW, len(truncated))
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Errorf("expected truncated activity to end with '...', got %q", truncated)
	}
}

func TestDisplay_PerEvalLines(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{
		lines:     make([]evalLine, 0),
		lineIndex: make(map[string]int),
		total:     3,
		w:         &buf,
		disabled:  false,
		width:     120,
		startTime: time.Now(),
	}

	// Start 3 evals — each should get its own line
	d.HandleEvent(ProgressEvent{EvalID: "a/x", PromptID: "a", ConfigName: "x", Type: EventStarting})
	d.HandleEvent(ProgressEvent{EvalID: "b/y", PromptID: "b", ConfigName: "y", Type: EventStarting})
	d.HandleEvent(ProgressEvent{EvalID: "c/z", PromptID: "c", ConfigName: "z", Type: EventStarting})

	if len(d.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(d.lines))
	}

	// Complete first, third still active — all 3 lines should persist
	d.HandleEvent(ProgressEvent{EvalID: "a/x", Type: EventPassed, FileCount: 2})
	if len(d.lines) != 3 {
		t.Errorf("expected 3 lines after completion, got %d", len(d.lines))
	}
	if !d.lines[0].completed {
		t.Error("expected line 0 completed")
	}
	if d.lines[1].completed || d.lines[2].completed {
		t.Error("expected lines 1,2 still active")
	}
}

func TestDisplay_CompletedEvalCount(t *testing.T) {
	d := &Display{
		lines: []evalLine{
			{completed: true},
			{completed: false},
			{completed: true},
		},
		lineIndex: make(map[string]int),
	}
	if got := d.CompletedEvalCount(); got != 2 {
		t.Errorf("expected CompletedEvalCount=2, got %d", got)
	}
}
