package progress

import (
"fmt"
"io"
"os"
"sync"
"time"
)

type DisplayConfig struct {
Total     int
Workers   int
Writer    io.Writer
Disabled  bool
ReportDir string
}

type evalState struct {
promptID   string
configName string
startTime  time.Time
}

type Display struct {
total     int
completed int
passed    int
failed    int
errors    int
mu        sync.Mutex
w         io.Writer
disabled  bool
startTime time.Time
reportDir string
active    map[string]*evalState
}

func NewDisplay(cfg DisplayConfig) *Display {
w := cfg.Writer
if w == nil {
w = os.Stdout
}
disabled := cfg.Disabled
if !disabled && cfg.Writer == nil {
disabled = !IsTerminal(os.Stdout)
}
d := &Display{
total:     cfg.Total,
w:         w,
disabled:  disabled,
startTime: time.Now(),
reportDir: cfg.ReportDir,
active:    make(map[string]*evalState),
}
if !d.disabled {
fmt.Fprintf(d.w, "\nRunning %d evaluations (%d workers)\n\n", cfg.Total, cfg.Workers)
}
return d
}

func (d *Display) HandleEvent(evt ProgressEvent) {
if d.disabled {
return
}
d.mu.Lock()
defer d.mu.Unlock()

switch evt.Type {
case EventStarting:
d.active[evt.EvalID] = &evalState{
promptID:   evt.PromptID,
configName: evt.ConfigName,
startTime:  time.Now(),
}

case EventPassed:
d.completed++
d.passed++
if s, ok := d.active[evt.EvalID]; ok {
dur := time.Since(s.startTime)
score := ""
if evt.ReviewScore > 0 {
score = fmt.Sprintf("  %d/10", evt.ReviewScore)
}
fmt.Fprintf(d.w, "  ✅ %-40s %d files%s  %s\n",
s.promptID+"/"+s.configName, evt.FileCount, score, fmtDuration(dur))
delete(d.active, evt.EvalID)
}

case EventFailed:
d.completed++
d.failed++
if s, ok := d.active[evt.EvalID]; ok {
dur := time.Since(s.startTime)
msg := "verification failed"
if evt.Message != "" {
msg = evt.Message
}
fmt.Fprintf(d.w, "  ❌ %-40s %s  %s\n",
s.promptID+"/"+s.configName, msg, fmtDuration(dur))
delete(d.active, evt.EvalID)
}

case EventError:
d.completed++
d.errors++
if s, ok := d.active[evt.EvalID]; ok {
dur := time.Since(s.startTime)
msg := "ERROR"
if evt.Message != "" {
msg = evt.Message
}
fmt.Fprintf(d.w, "  ❌ %-40s %s  %s\n",
s.promptID+"/"+s.configName, msg, fmtDuration(dur))
delete(d.active, evt.EvalID)
}
}
}

func (d *Display) Finish() {
if d.disabled {
return
}
d.mu.Lock()
defer d.mu.Unlock()

elapsed := time.Since(d.startTime)
fmt.Fprintf(d.w, "\nSummary: %d/%d passed", d.passed, d.total)
fmt.Fprintf(d.w, "  ✅ %d", d.passed)
if d.failed > 0 {
fmt.Fprintf(d.w, "  ❌ %d", d.failed)
}
if d.errors > 0 {
fmt.Fprintf(d.w, "  ❌ %d errors", d.errors)
}
fmt.Fprintf(d.w, "  Duration: %s\n", fmtDuration(elapsed))
if d.reportDir != "" {
fmt.Fprintf(d.w, "Reports: %s\n", d.reportDir)
}
}

func (d *Display) Done() { d.Finish() }

func (d *Display) CompletedEvalCount() int {
d.mu.Lock()
defer d.mu.Unlock()
return d.completed
}

func IsTerminal(f *os.File) bool {
fi, err := f.Stat()
if err != nil {
return false
}
return fi.Mode()&os.ModeCharDevice != 0
}

func TermWidth() int { return 120 }
