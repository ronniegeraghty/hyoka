// Package progress provides progress reporting and display for the evaluation tool.
package progress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ProgressMode controls the rendering strategy.
type ProgressMode string

const (
	ModeAuto ProgressMode = "auto" // ANSI if TTY, log otherwise
	ModeLive ProgressMode = "live" // Force ANSI (cursor save/restore)
	ModeLog  ProgressMode = "log"  // Append-only phase lines (no cursor movement)
	ModeOff  ProgressMode = "off"  // No progress output
)

// DisplayConfig controls the progress display.
type DisplayConfig struct {
	Total     int
	Workers   int
	Writer    io.Writer
	Disabled  bool
	ReportDir string
	Mode      ProgressMode // "" or "auto" uses auto-detection
}

type evalStatus int

const (
	evalActive evalStatus = iota
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
// position with DECSC (\0337), and redraws the eval region on a 500ms
// timer using DECRC (\0338) + clear-to-end (\033[J). Lines are appended
// dynamically as evals start — there are no pre-allocated "waiting" lines.
// The ANSI region grows downward as new evals begin.
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

	// Dynamic eval lines — grows as evals start (not pre-allocated).
	lines     []*evalLine
	lineIndex map[string]int

	// ANSI redraw timer
	ticker *time.Ticker
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewDisplay creates a progress display. When Writer is nil, it writes to
// os.Stdout and enables ANSI rendering if stdout is a terminal.
// Mode overrides auto-detection: "live" forces ANSI, "log" forces append-only,
// "off" disables output entirely.
func NewDisplay(cfg DisplayConfig) *Display {
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	disabled := cfg.Disabled
	ansi := false

	switch cfg.Mode {
	case ModeOff:
		disabled = true
	case ModeLive:
		if !disabled {
			ansi = true
		}
	case ModeLog:
		// Non-ANSI mode: append-only lines, no cursor movement
		ansi = false
	default: // ModeAuto or ""
		if !disabled && cfg.Writer == nil {
			if IsTerminal(os.Stdout) {
				ansi = true
			} else {
				disabled = true
			}
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
		lines:     []*evalLine{},
		lineIndex: make(map[string]int),
	}

	if d.ansi && cfg.Total > 0 {
		fmt.Fprintf(d.w, "\nRunning %d evaluations (%d workers)\n", cfg.Total, cfg.Workers)
		fmt.Fprint(d.w, "\0337") // DECSC: save cursor position
		d.drawRegion()
		d.stopCh = make(chan struct{})
		d.ticker = time.NewTicker(500 * time.Millisecond)
		d.wg.Add(1)
		go d.redrawLoop()
	} else if !d.disabled {
		fmt.Fprintf(d.w, "\nRunning %d evaluations (%d workers)\n\n", cfg.Total, cfg.Workers)
	}

	return d
}

// --- ANSI fixed-region rendering (terminal only) ---

// buildRegion renders the eval region (started lines + summary) into a buffer.
// Only lines for evals that have started are included — no waiting placeholders.
func (d *Display) buildRegion() []byte {
	var buf bytes.Buffer
	for _, l := range d.lines {
		d.writeEvalLine(&buf, l)
	}
	buf.WriteByte('\n')
	d.writeSummaryLine(&buf)
	return buf.Bytes()
}

// drawRegion writes the eval region to the output writer atomically.
func (d *Display) drawRegion() {
	d.w.Write(d.buildRegion())
}

// redrawRegion restores cursor to the saved position, clears everything
// below, and redraws the region — all as a single atomic write.
// Uses DECRC (\0338) + ED (\033[J) which is more widely supported than
// the SCO \033[u sequence.
func (d *Display) redrawRegion() {
	region := d.buildRegion()
	// Prepend restore-cursor + clear-to-end-of-screen, write everything at once
	var buf bytes.Buffer
	buf.WriteString("\0338\033[J")
	buf.Write(region)
	d.w.Write(buf.Bytes())
}

func (d *Display) writeEvalLine(buf *bytes.Buffer, l *evalLine) {
	switch l.status {
	case evalActive:
		activity := l.activity
		if activity == "" && l.phase != "" {
			activity = string(l.phase)
		}
		if activity != "" {
			fmt.Fprintf(buf, "  🔄 %-40s  %s  %s\n", l.name, activity, fmtDuration(time.Since(l.startTime)))
		} else {
			fmt.Fprintf(buf, "  🔄 %-40s  %s\n", l.name, fmtDuration(time.Since(l.startTime)))
		}
	case evalPassed:
		score := ""
		if l.reviewScore > 0 {
			score = fmt.Sprintf("  %d/10", l.reviewScore)
		}
		fmt.Fprintf(buf, "  ✅ %-40s %d files%s  %s\n", l.name, l.fileCount, score, fmtDuration(l.duration))
	case evalFailed, evalError:
		msg := l.message
		if msg == "" {
			msg = "failed"
		}
		fmt.Fprintf(buf, "  ❌ %-40s %s  %s\n", l.name, msg, fmtDuration(l.duration))
	}
}

func (d *Display) writeSummaryLine(buf *bytes.Buffer) {
	if d.completed == d.total && d.total > 0 {
		fmt.Fprintf(buf, "  Summary: %d/%d passed", d.passed, d.total)
	} else {
		fmt.Fprintf(buf, "  %d/%d completed", d.completed, d.total)
	}
	if d.passed > 0 {
		fmt.Fprintf(buf, "  ✅ %d", d.passed)
	}
	if d.failed > 0 {
		fmt.Fprintf(buf, "  ❌ %d", d.failed)
	}
	if d.errors > 0 {
		fmt.Fprintf(buf, "  ❌ %d errors", d.errors)
	}
	fmt.Fprintf(buf, "  %s\n", fmtDuration(time.Since(d.startTime)))
}

func (d *Display) redrawLoop() {
	defer d.wg.Done()
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
	idx := len(d.lines)
	d.lines = append(d.lines, &evalLine{
		name: promptID + "/" + configName,
	})
	d.lineIndex[evalID] = idx
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
		l := d.lines[idx]
		l.status = evalActive
		l.startTime = time.Now()
		l.activity = evt.Message
		if !d.ansi {
			fmt.Fprintf(d.w, "  ▶ %-40s  starting...\n", l.name)
		}

	case EventSendingPrompt, EventReasoning, EventToolStart, EventToolComplete,
		EventWritingFile, EventWaiting:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].activity = evt.Message
			if !d.ansi && evt.Message != "" {
				prefix := "  "
				switch evt.Type {
				case EventToolStart:
					prefix = "    🔧"
				case EventToolComplete:
					prefix = "    ✓ "
				case EventWritingFile:
					prefix = "    📄"
				case EventSendingPrompt:
					prefix = "    📨"
				case EventReasoning:
					prefix = "    💭"
				default:
					prefix = "    ⏳"
				}
				fmt.Fprintf(d.w, "%s [%s] %s\n", prefix, d.lines[idx].name, evt.Message)
			}
		}

	case EventPhaseChange:
		if idx, ok := d.lineIndex[evt.EvalID]; ok {
			d.lines[idx].phase = evt.Phase
			d.lines[idx].activity = string(evt.Phase)
			if !d.ansi {
				fmt.Fprintf(d.w, "  ▶ %-40s  %s...\n", d.lines[idx].name, evt.Phase)
			}
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

	// ANSI mode: stop timer, wait for redrawLoop to exit, then final redraw
	if d.ansi && d.ticker != nil {
		d.ticker.Stop()
		close(d.stopCh)
		d.wg.Wait()
		d.mu.Lock()
		if d.total > 0 {
			d.redrawRegion()
		}
		// Print reports path below the region (no cursor restore — this is final)
		if d.reportDir != "" {
			fmt.Fprintf(d.w, "\nReports: %s\n", d.reportDir)
		} else {
			fmt.Fprintln(d.w)
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

func fmtDuration(d time.Duration) string {
	secs := d.Seconds()
	if secs < 60 {
		return fmt.Sprintf("%.0fs", secs)
	}
	return fmt.Sprintf("%.1fm", secs/60)
}
