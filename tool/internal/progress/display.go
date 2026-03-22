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
	phase       Phase  // Current phase (generating/verifying/reviewing)
	icon        string // Activity icon within the phase
	activity    string // Activity description within the phase
	startTime   time.Time
	completed   bool
	passed      bool
	errored     bool
	fileCount   int
	reviewScore int
	message     string        // failure/error message
	duration    time.Duration // set on completion
}

// Display renders multi-line per-eval progress with live status updates.
// Each eval gets a dedicated line that persists through completion.
// Lines are appended as evals start and grow the display downward.
type Display struct {
	lines         []evalLine     // ordered list of all evals
	lineIndex     map[string]int // evalID → index in lines
	total         int
	completed     int
	passed        int
	failed        int
	errors        int
	mu            sync.Mutex
	w             io.Writer
	rendered      bool
	disabled      bool
	width         int
	startTime     time.Time
	reportDir     string
	lastLineCount int // lines rendered on previous redraw (for ANSI cursor)
}

// NewDisplay creates a multi-line progress display.
// Automatically disabled when stdout is not a terminal.
func NewDisplay(cfg DisplayConfig) *Display {
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	disabled := cfg.Disabled
	if !disabled {
		disabled = !IsTerminal(os.Stdout)
	}

	d := &Display{
		lines:     make([]evalLine, 0, cfg.Total),
		lineIndex: make(map[string]int),
		total:     cfg.Total,
		w:         w,
		disabled:  disabled,
		width:     TermWidth(),
		startTime: time.Now(),
		reportDir: cfg.ReportDir,
	}

	if !d.disabled {
		fmt.Fprintf(d.w, "\n%sRunning %d evaluations (%d workers)%s\n\n",
			ColorBold, cfg.Total, cfg.Workers, ColorReset)
		d.redraw()
	}

	return d
}

// phaseIcon returns the primary status icon for the current eval phase.
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

// phaseLabel returns the short label for the current eval phase.
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
		idx := len(d.lines)
		d.lines = append(d.lines, evalLine{
			evalID:     evt.EvalID,
			promptID:   evt.PromptID,
			configName: evt.ConfigName,
			phase:      PhaseGenerating,
			icon:       "⏳",
			activity:   "Waiting for session...",
			startTime:  time.Now(),
		})
		d.lineIndex[evt.EvalID] = idx

	case EventPhaseChange:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].phase = evt.Phase
			d.lines[idx].icon = phaseIcon(evt.Phase)
			if evt.Message != "" {
				d.lines[idx].activity = evt.Message
			} else {
				d.lines[idx].activity = phaseLabel(evt.Phase) + "..."
			}
		}

	case EventSendingPrompt:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].icon = "→"
			d.lines[idx].activity = evt.Message
		}

	case EventReasoning:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].icon = "💭"
			d.lines[idx].activity = evt.Message
		}

	case EventToolStart:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].icon = "⚙"
			d.lines[idx].activity = evt.Message
		}

	case EventToolComplete:
		// Don't change icon — keep the current phase icon (⚙/🔄/🔍/📝).
		// Just update activity text so it shows what completed.
		// Changing to ✓ makes it look like the eval is done.
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].activity = evt.Message
		}

	case EventWritingFile:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].icon = "📝"
			d.lines[idx].activity = evt.Message
		}

	case EventWaiting:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].icon = "⏳"
			d.lines[idx].activity = evt.Message
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
			delete(d.lineIndex, evt.EvalID)
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
			delete(d.lineIndex, evt.EvalID)
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
			delete(d.lineIndex, evt.EvalID)
		}
	}

	d.redraw()
}

// Finish stops the ANSI refresh loop and prints all final results as static output.
func (d *Display) Finish() {
	if d.disabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear the live display region
	if d.rendered && d.lastLineCount > 0 {
		fmt.Fprintf(d.w, "\033[%dA", d.lastLineCount)
		for i := 0; i < d.lastLineCount; i++ {
			fmt.Fprintf(d.w, "\033[2K\n")
		}
		fmt.Fprintf(d.w, "\033[%dA", d.lastLineCount)
	}

	// Print all eval lines as static output (no ANSI cursor movement)
	for i := range d.lines {
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

	// Summary line
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

// Done finalizes the display with a summary line (backward compat — prefer Finish).
func (d *Display) Done() {
	d.Finish()
}

// CompletedEvalCount returns the number of completed evals (for testing).
func (d *Display) CompletedEvalCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	count := 0
	for _, l := range d.lines {
		if l.completed {
			count++
		}
	}
	return count
}

func (d *Display) redraw() {
	// Move cursor to start of display area
	if d.rendered && d.lastLineCount > 0 {
		fmt.Fprintf(d.w, "\033[%dA", d.lastLineCount)
	}

	actW := d.activityWidth()

	// Render all eval lines (completed show results, active show current activity)
	for i := range d.lines {
		l := &d.lines[i]
		fmt.Fprintf(d.w, "\033[K") // clear to end of line
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
			pLabel := phaseLabel(l.phase)
			pIcon := phaseIcon(l.phase)
			activity := d.truncateActivity(l.activity)
			elapsed := fmtDuration(time.Since(l.startTime))
			fmt.Fprintf(d.w, "  %-40s %s %-12s %s %-*s %6s",
				name, pIcon, pLabel, l.icon, actW-16, activity, elapsed)
		}
		fmt.Fprint(d.w, "\n")
	}

	// Blank line + summary
	fmt.Fprintf(d.w, "\033[K\n")
	fmt.Fprintf(d.w, "\033[K  Completed: %d/%d", d.completed, d.total)
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

	d.lastLineCount = len(d.lines) + 2 // eval lines + blank + summary
	d.rendered = true
}

func (d *Display) formatName(promptID, configName string) string {
	name := promptID + "/" + configName
	const maxLen = 38
	if len(name) > maxLen {
		name = name[:maxLen-2] + ".."
	}
	return name
}

func (d *Display) activityWidth() int {
	// Layout: 2 indent + 40 name + 1 space + ~3 icon + 1 space + activity + 2 spaces + 6 elapsed
	w := d.width - 55
	if w < 20 {
		w = 20
	}
	return w
}

func (d *Display) truncateActivity(s string) string {
	maxLen := d.activityWidth()
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// TermWidth returns the terminal width from COLUMNS env var, defaulting to 120.
func TermWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}
	return 120
}

// IsTerminal reports whether f is connected to a terminal device.
func IsTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
