package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// DisplayConfig controls the progress display.
type DisplayConfig struct {
	Total     int
	Workers   int
	Writer    io.Writer
	Disabled  bool
	ReportDir string
}

type evalStatus int

const (
	evalPending evalStatus = iota
	evalActive
	evalPassed
	evalFailed
	evalError
)

// evalLine holds the state for one eval slot in the fixed-region display.
type evalLine struct {
	name        string
	status      evalStatus
	startTime   time.Time
	activity    string
	phase       Phase
	fileCount   int
	reviewScore int
	message     string
	duration    time.Duration
}

// Display renders live progress for evaluation runs.
//
// In ANSI mode (real terminal), it prints a header, saves the cursor
// position, and redraws the eval region on a 500ms timer using
// save/restore cursor (\033[s / \033[u) + clear-to-end (\033[J).
// This avoids cursor-up arithmetic that breaks when emoji or wide
// characters cause line wrapping.
//
// In non-ANSI mode (piped output, test buffers), it prints result lines
// inline as evals complete, then a summary on Finish.
type Display struct {
	total     int
	workers   int
	completed int
	passed    int
	failed    int
	errors    int
	mu        sync.Mutex
	w         io.Writer
	disabled  bool
	ansi      bool
	startTime time.Time
	reportDir string

	// Fixed-region state: ordered eval slots, assigned on EventStarting.
	lines     []*evalLine
	lineIndex map[string]int
	nextSlot  int

	// ANSI redraw timer
	ticker *time.Ticker
	stopCh chan struct{}
}

// NewDisplay creates a progress display. When Writer is nil, it writes to
// os.Stdout and enables ANSI rendering if stdout is a terminal.
func NewDisplay(cfg DisplayConfig) *Display {
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	disabled := cfg.Disabled
	ansi := false
	if !disabled && cfg.Writer == nil {
		if IsTerminal(os.Stdout) {
			ansi = true
		} else {
			disabled = true
		}
	}

	d := &Display{
		total:     cfg.Total,
		workers:   cfg.Workers,
		w:         w,
		disabled:  disabled,
		ansi:      ansi,
		startTime: time.Now(),
		reportDir: cfg.ReportDir,
		lines:     make([]*evalLine, cfg.Total),
		lineIndex: make(map[string]int),
	}

	for i := range d.lines {
		d.lines[i] = &evalLine{status: evalPending}
	}

	if d.ansi && cfg.Total > 0 {
		fmt.Fprintf(d.w, "\nRunning %d evaluations (%d workers)\n", cfg.Total, cfg.Workers)
		fmt.Fprint(d.w, "\033[s") // save cursor position
		d.drawRegion()
		d.stopCh = make(chan struct{})
		d.ticker = time.NewTicker(500 * time.Millisecond)
		go d.redrawLoop()
	} else if !d.disabled {
		fmt.Fprintf(d.w, "\nRunning %d evaluations (%d workers)\n\n", cfg.Total, cfg.Workers)
	}

	return d
}

// --- ANSI fixed-region rendering (terminal only) ---

func (d *Display) drawRegion() {
	for i := 0; i < d.total; i++ {
		d.drawEvalLine(d.lines[i])
	}
	fmt.Fprintln(d.w)
	d.drawSummaryLine()
}

func (d *Display) redrawRegion() {
	// Restore saved cursor position, then clear everything below it.
	// This is more reliable than cursor-up because it doesn't depend
	// on counting wrapped lines caused by emoji / wide characters.
	fmt.Fprint(d.w, "\033[u\033[J")
	d.drawRegion()
}

func (d *Display) drawEvalLine(l *evalLine) {
	switch l.status {
	case evalPending:
		fmt.Fprintf(d.w, "  \033[2m⏳ (waiting)\033[0m\n")
	case evalActive:
		activity := l.activity
		if activity == "" && l.phase != "" {
			activity = string(l.phase)
		}
		if activity != "" {
			fmt.Fprintf(d.w, "  🔄 %-40s  %s  %s\n", l.name, activity, fmtDuration(time.Since(l.startTime)))
		} else {
			fmt.Fprintf(d.w, "  🔄 %-40s  %s\n", l.name, fmtDuration(time.Since(l.startTime)))
		}
	case evalPassed:
		score := ""
		if l.reviewScore > 0 {
			score = fmt.Sprintf("  %d/10", l.reviewScore)
		}
		fmt.Fprintf(d.w, "  ✅ %-40s %d files%s  %s\n", l.name, l.fileCount, score, fmtDuration(l.duration))
	case evalFailed, evalError:
		msg := l.message
		if msg == "" {
			msg = "failed"
		}
		fmt.Fprintf(d.w, "  ❌ %-40s %s  %s\n", l.name, msg, fmtDuration(l.duration))
	}
}

func (d *Display) drawSummaryLine() {
	if d.completed == d.total && d.total > 0 {
		fmt.Fprintf(d.w, "  Summary: %d/%d passed", d.passed, d.total)
	} else {
		fmt.Fprintf(d.w, "  %d/%d completed", d.completed, d.total)
	}
	if d.passed > 0 {
		fmt.Fprintf(d.w, "  ✅ %d", d.passed)
	}
	if d.failed > 0 {
		fmt.Fprintf(d.w, "  ❌ %d", d.failed)
	}
	if d.errors > 0 {
		fmt.Fprintf(d.w, "  ❌ %d errors", d.errors)
	}
	fmt.Fprintf(d.w, "  %s\n", fmtDuration(time.Since(d.startTime)))
}

func (d *Display) redrawLoop() {
	for {
		select {
		case <-d.stopCh:
			return
		case <-d.ticker.C:
			d.mu.Lock()
			d.redrawRegion()
			d.mu.Unlock()
		}
	}
}

// --- Slot assignment ---

func (d *Display) getOrAssignSlot(evalID, promptID, configName string) int {
	if idx, ok := d.lineIndex[evalID]; ok {
		return idx
	}
	if d.nextSlot >= len(d.lines) {
		return -1
	}
	idx := d.nextSlot
	d.nextSlot++
	d.lineIndex[evalID] = idx
	d.lines[idx].name = promptID + "/" + configName
	return idx
}

// --- Event handling ---

// HandleEvent updates internal state from engine/evaluator events.
// In ANSI mode, rendering happens on the timer — not here.
// In non-ANSI mode, completion events print inline.
func (d *Display) HandleEvent(evt ProgressEvent) {
	if d.disabled {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	switch evt.Type {
	case EventStarting:
		idx := d.getOrAssignSlot(evt.EvalID, evt.PromptID, evt.ConfigName)
		if idx < 0 {
			return
		}
		l := d.lines[idx]
		l.status = evalActive
		l.startTime = time.Now()
		l.activity = evt.Message

	case EventSendingPrompt, EventReasoning, EventToolStart, EventToolComplete,
		EventWritingFile, EventWaiting:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].activity = evt.Message
		}

	case EventPhaseChange:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].phase = evt.Phase
			d.lines[idx].activity = string(evt.Phase)
		}

	case EventPassed:
		d.completed++
		d.passed++
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			l := d.lines[idx]
			l.status = evalPassed
			l.duration = time.Since(l.startTime)
			l.fileCount = evt.FileCount
			l.reviewScore = evt.ReviewScore
			if !d.ansi {
				score := ""
				if evt.ReviewScore > 0 {
					score = fmt.Sprintf("  %d/10", evt.ReviewScore)
				}
				fmt.Fprintf(d.w, "  ✅ %-40s %d files%s  %s\n",
					l.name, evt.FileCount, score, fmtDuration(l.duration))
			}
		}

	case EventFailed:
		d.completed++
		d.failed++
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			l := d.lines[idx]
			l.status = evalFailed
			l.duration = time.Since(l.startTime)
			l.message = evt.Message
			if l.message == "" {
				l.message = "verification failed"
			}
			if !d.ansi {
				fmt.Fprintf(d.w, "  ❌ %-40s %s  %s\n",
					l.name, l.message, fmtDuration(l.duration))
			}
		}

	case EventError:
		d.completed++
		d.errors++
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			l := d.lines[idx]
			l.status = evalError
			l.duration = time.Since(l.startTime)
			l.message = evt.Message
			if l.message == "" {
				l.message = "ERROR"
			}
			if !d.ansi {
				fmt.Fprintf(d.w, "  ❌ %-40s %s  %s\n",
					l.name, l.message, fmtDuration(l.duration))
			}
		}
	}
}

// Finish stops the redraw timer, renders final state, and prints the summary.
func (d *Display) Finish() {
	if d.disabled {
		return
	}

	// ANSI mode: stop timer, final redraw, print reports below region
	if d.ansi && d.ticker != nil {
		d.ticker.Stop()
		close(d.stopCh)
		time.Sleep(10 * time.Millisecond)
		d.mu.Lock()
		if d.total > 0 {
			d.redrawRegion()
		}
		fmt.Fprintln(d.w)
		if d.reportDir != "" {
			fmt.Fprintf(d.w, "Reports: %s\n", d.reportDir)
		}
		d.mu.Unlock()
		return
	}

	// Non-ANSI mode: print summary below inline results
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Fprint(d.w, "\n\n")
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

// Done is an alias for Finish.
func (d *Display) Done() { d.Finish() }

// CompletedEvalCount returns the number of evals that have completed.
func (d *Display) CompletedEvalCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.completed
}

// IsTerminal reports whether f is connected to a terminal.
func IsTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// TermWidth returns the assumed terminal width.
func TermWidth() int { return 120 }
