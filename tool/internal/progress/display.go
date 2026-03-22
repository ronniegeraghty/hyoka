package progress

import (
"fmt"
"io"
"os"
"strconv"
"sync"
"time"
)

// DisplayConfig configures the multi-line progress display.
type DisplayConfig struct {
Total     int       // Total evaluations
Workers   int       // Parallel workers
Writer    io.Writer // Output writer (default: os.Stdout)
Disabled  bool      // Force disabled (debug, dry-run, piped output)
ReportDir string    // Report directory path (shown in final output)
}

// evalLine tracks one eval's display line (active or completed).
type evalLine struct {
evalID      string
promptID    string
configName  string
phase       Phase
activity    string
startTime   time.Time
completed   bool
passed      bool
errored     bool
fileCount   int
reviewScore int
message     string
duration    time.Duration
}

// Display renders multi-line per-eval progress.
// Pre-allocates N lines (one per eval) and redraws them in-place.
type Display struct {
lines     []evalLine
lineIndex map[string]int
order     []string // evalIDs in display order
total     int
completed int
passed    int
failed    int
errors    int
mu        sync.Mutex
w         io.Writer
disabled  bool
width     int
startTime time.Time
reportDir string
// ANSI: we always render exactly (total + 2) lines: N eval lines + blank + summary
// On first render we print them. On subsequent renders we cursor-up by that fixed amount.
fixedLines int
rendered   bool
}

// NewDisplay creates a multi-line progress display.
func NewDisplay(cfg DisplayConfig) *Display {
w := cfg.Writer
if w == nil {
w = os.Stdout
}

disabled := cfg.Disabled
if !disabled {
disabled = !IsTerminal(os.Stdout)
}

fixedLines := cfg.Total + 2 // eval lines + blank + summary

d := &Display{
lines:      make([]evalLine, cfg.Total),
lineIndex:  make(map[string]int),
order:      make([]string, 0, cfg.Total),
total:      cfg.Total,
w:          w,
disabled:   disabled,
width:      TermWidth(),
startTime:  time.Now(),
reportDir:  cfg.ReportDir,
fixedLines: fixedLines,
}

// Initialize all lines as empty placeholders
for i := range d.lines {
d.lines[i].activity = ""
}

if !d.disabled {
fmt.Fprintf(d.w, "\n%sRunning %d evaluations (%d workers)%s\n\n",
ColorBold, cfg.Total, cfg.Workers, ColorReset)
// Print initial blank lines to reserve space
for i := 0; i < fixedLines; i++ {
fmt.Fprint(d.w, "\n")
}
d.rendered = true
d.redraw()
}

return d
}

func phaseIcon(p Phase) string {
switch p {
case PhaseGenerating:
return "🔄"
case PhaseVerifying:
return "🔍"
case PhaseReviewing:
return "📝"
default:
return "⏳"
}
}

func phaseLabel(p Phase) string {
switch p {
case PhaseGenerating:
return "Generating"
case PhaseVerifying:
return "Verifying"
case PhaseReviewing:
return "Reviewing"
default:
return "Starting"
}
}

// HandleEvent processes a progress event and redraws the display.
func (d *Display) HandleEvent(evt ProgressEvent) {
if d.disabled {
return
}

d.mu.Lock()
defer d.mu.Unlock()

switch evt.Type {
case EventStarting:
slot := len(d.order)
if slot >= d.total {
return // shouldn't happen
}
d.order = append(d.order, evt.EvalID)
d.lineIndex[evt.EvalID] = slot
d.lines[slot] = evalLine{
evalID:     evt.EvalID,
promptID:   evt.PromptID,
configName: evt.ConfigName,
phase:      PhaseGenerating,
activity:   "⏳ Waiting for session...",
startTime:  time.Now(),
}

case EventPhaseChange:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].phase = evt.Phase
if evt.Message != "" {
d.lines[idx].activity = evt.Message
} else {
d.lines[idx].activity = phaseLabel(evt.Phase) + "..."
}
}

case EventSendingPrompt:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].activity = "→ " + evt.Message
}

case EventReasoning:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].activity = "💭 " + evt.Message
}

case EventToolStart:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].activity = "⚙ " + evt.Message
}

case EventToolComplete:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
if evt.Message != "" {
d.lines[idx].activity = "⚙ " + evt.Message + " done"
}
}

case EventWritingFile:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].activity = "📝 " + evt.Message
}

case EventWaiting:
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].activity = "⏳ " + evt.Message
}

case EventPassed:
d.completed++
d.passed++
if idx, ok := d.lineIndex[evt.EvalID]; ok {
d.lines[idx].completed = true
d.lines[idx].passed = true
d.lines[idx].fileCount = evt.FileCount
d.lines[idx].reviewScore = evt.ReviewScore
d.lines[idx].duration = time.Since(d.lines[idx].startTime)
}

case EventFailed:
d.completed++
d.failed++
if idx, ok := d.lineIndex[evt.EvalID]; ok {
msg := "verification failed"
if evt.Message != "" {
msg = evt.Message
}
d.lines[idx].completed = true
d.lines[idx].message = msg
d.lines[idx].duration = time.Since(d.lines[idx].startTime)
}

case EventError:
d.completed++
d.errors++
if idx, ok := d.lineIndex[evt.EvalID]; ok {
msg := "ERROR"
if evt.Message != "" {
msg = evt.Message
}
d.lines[idx].completed = true
d.lines[idx].errored = true
d.lines[idx].message = msg
d.lines[idx].duration = time.Since(d.lines[idx].startTime)
}
}

d.redraw()
}

func (d *Display) redraw() {
// Always move up by the fixed line count
fmt.Fprintf(d.w, "\033[%dA", d.fixedLines)

actW := d.activityWidth()
started := len(d.order)

for i := 0; i < d.total; i++ {
fmt.Fprintf(d.w, "\033[2K") // clear line
if i < started {
l := &d.lines[i]
name := d.formatName(l.promptID, l.configName)
if l.completed {
if l.errored {
fmt.Fprintf(d.w, "  %-40s ❌ %s  %s", name, l.message, fmtDuration(l.duration))
} else if !l.passed {
fmt.Fprintf(d.w, "  %-40s ❌ FAILED  %s  %s", name, l.message, fmtDuration(l.duration))
} else {
score := ""
if l.reviewScore > 0 {
score = fmt.Sprintf("  %d/10", l.reviewScore)
}
fmt.Fprintf(d.w, "  %-40s ✅ PASSED  %d files%s  %s", name, l.fileCount, score, fmtDuration(l.duration))
}
} else {
pIcon := phaseIcon(l.phase)
pLabel := phaseLabel(l.phase)
activity := d.truncate(l.activity, actW-12)
elapsed := fmtDuration(time.Since(l.startTime))
fmt.Fprintf(d.w, "  %-40s %s %-12s %-*s %6s",
name, pIcon, pLabel, actW-12, activity, elapsed)
}
}
fmt.Fprint(d.w, "\n")
}

// Blank + summary
fmt.Fprintf(d.w, "\033[2K\n")
fmt.Fprintf(d.w, "\033[2K  Completed: %d/%d", d.completed, d.total)
if d.passed > 0 {
fmt.Fprintf(d.w, "  %s✅ %d%s", ColorGreen, d.passed, ColorReset)
}
if d.failed > 0 {
fmt.Fprintf(d.w, "  %s❌ %d%s", ColorRed, d.failed, ColorReset)
}
if d.errors > 0 {
fmt.Fprintf(d.w, "  %s❌ %d errors%s", ColorRed, d.errors, ColorReset)
}
fmt.Fprint(d.w, "\n")
}

// Finish prints final static results.
func (d *Display) Finish() {
if d.disabled {
return
}

d.mu.Lock()
defer d.mu.Unlock()

// Clear the live display if it was rendered
if d.rendered && d.fixedLines > 0 {
fmt.Fprintf(d.w, "\033[%dA", d.fixedLines)
for i := 0; i < d.fixedLines; i++ {
fmt.Fprintf(d.w, "\033[2K\n")
}
fmt.Fprintf(d.w, "\033[%dA", d.fixedLines)
}

// Print final static lines
for i := 0; i < len(d.order); i++ {
l := &d.lines[i]
name := d.formatName(l.promptID, l.configName)
if l.errored {
fmt.Fprintf(d.w, "  %-40s ❌ %s  %s\n", name, l.message, fmtDuration(l.duration))
} else if l.completed && !l.passed {
fmt.Fprintf(d.w, "  %-40s ❌ FAILED  %s  %s\n", name, l.message, fmtDuration(l.duration))
} else if l.completed {
score := ""
if l.reviewScore > 0 {
score = fmt.Sprintf("  %d/10", l.reviewScore)
}
fmt.Fprintf(d.w, "  %-40s ✅ PASSED  %d files%s  %s\n", name, l.fileCount, score, fmtDuration(l.duration))
}
}

elapsed := time.Since(d.startTime)
fmt.Fprintf(d.w, "\n%sSummary: %d/%d passed%s", ColorBold, d.passed, d.total, ColorReset)
fmt.Fprintf(d.w, "  %s✅ %d%s", ColorGreen, d.passed, ColorReset)
if d.failed > 0 {
fmt.Fprintf(d.w, "  %s❌ %d%s", ColorRed, d.failed, ColorReset)
}
if d.errors > 0 {
fmt.Fprintf(d.w, "  %s❌ %d errors%s", ColorRed, d.errors, ColorReset)
}
fmt.Fprintf(d.w, "  Duration: %s\n", fmtDuration(elapsed))
if d.reportDir != "" {
fmt.Fprintf(d.w, "Reports: %s\n", d.reportDir)
}
}

// Done is an alias for Finish.
func (d *Display) Done() { d.Finish() }

// CompletedEvalCount returns the number of completed evals.
func (d *Display) CompletedEvalCount() int {
d.mu.Lock()
defer d.mu.Unlock()
c := 0
for _, l := range d.lines {
if l.completed {
c++
}
}
return c
}

func (d *Display) formatName(promptID, configName string) string {
name := promptID + "/" + configName
if len(name) > 38 {
name = name[:36] + ".."
}
return name
}

func (d *Display) activityWidth() int {
w := d.width - 55
if w < 20 {
w = 20
}
return w
}

func (d *Display) truncate(s string, maxLen int) string {
if maxLen <= 0 {
return ""
}
if len(s) > maxLen {
return s[:maxLen-3] + "..."
}
return s
}

// TermWidth returns terminal width.
func TermWidth() int {
if cols := os.Getenv("COLUMNS"); cols != "" {
if n, err := strconv.Atoi(cols); err == nil && n > 0 {
return n
}
}
return 120
}

// IsTerminal reports whether f is a terminal.
func IsTerminal(f *os.File) bool {
fi, err := f.Stat()
if err != nil {
return false
}
return fi.Mode()&os.ModeCharDevice != 0
}
