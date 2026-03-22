package progress

import (
"bytes"
"strings"
"testing"
)

func TestDisplay_BasicFlow(t *testing.T) {
var buf bytes.Buffer
d := NewDisplay(DisplayConfig{Total: 2, Workers: 2, Writer: &buf, Disabled: false})

d.HandleEvent(ProgressEvent{EvalID: "a", PromptID: "p1", ConfigName: "c1", Type: EventStarting})
d.HandleEvent(ProgressEvent{EvalID: "b", PromptID: "p2", ConfigName: "c2", Type: EventStarting})
d.HandleEvent(ProgressEvent{EvalID: "a", Type: EventPassed, FileCount: 3, ReviewScore: 8})
d.HandleEvent(ProgressEvent{EvalID: "b", Type: EventFailed, Message: "bad code"})
d.Finish()

out := buf.String()
if !strings.Contains(out, "p1/c1") {
t.Errorf("expected p1/c1 in output, got %q", out)
}
if !strings.Contains(out, "p2/c2") {
t.Errorf("expected p2/c2 in output, got %q", out)
}
if !strings.Contains(out, "✅") {
t.Errorf("expected ✅ in output")
}
if !strings.Contains(out, "❌") {
t.Errorf("expected ❌ in output")
}
if !strings.Contains(out, "3 files") {
t.Errorf("expected '3 files' in output")
}
if !strings.Contains(out, "8/10") {
t.Errorf("expected '8/10' in output")
}
if !strings.Contains(out, "Summary: 1/2") {
t.Errorf("expected summary in output")
}
}

func TestDisplay_CompletedCount(t *testing.T) {
var buf bytes.Buffer
d := NewDisplay(DisplayConfig{Total: 3, Workers: 2, Writer: &buf, Disabled: false})
d.HandleEvent(ProgressEvent{EvalID: "a", PromptID: "p", ConfigName: "c", Type: EventStarting})
d.HandleEvent(ProgressEvent{EvalID: "a", Type: EventPassed, FileCount: 1})
if d.CompletedEvalCount() != 1 {
t.Errorf("expected 1 completed, got %d", d.CompletedEvalCount())
}
}

func TestDisplay_ReportDir(t *testing.T) {
var buf bytes.Buffer
d := NewDisplay(DisplayConfig{Total: 1, Workers: 1, Writer: &buf, ReportDir: "reports/123/"})
d.HandleEvent(ProgressEvent{EvalID: "a", PromptID: "p", ConfigName: "c", Type: EventStarting})
d.HandleEvent(ProgressEvent{EvalID: "a", Type: EventPassed, FileCount: 2})
d.Finish()
if !strings.Contains(buf.String(), "reports/123/") {
t.Errorf("expected report dir in output")
}
}

func TestDisplay_Disabled(t *testing.T) {
var buf bytes.Buffer
d := NewDisplay(DisplayConfig{Total: 1, Workers: 1, Writer: &buf, Disabled: true})
d.HandleEvent(ProgressEvent{EvalID: "a", PromptID: "p", ConfigName: "c", Type: EventStarting})
d.HandleEvent(ProgressEvent{EvalID: "a", Type: EventPassed, FileCount: 1})
d.Finish()
if buf.Len() != 0 {
t.Errorf("expected no output when disabled, got %q", buf.String())
}
}
