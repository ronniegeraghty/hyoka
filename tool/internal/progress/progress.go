package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
	ClearLine   = "\033[2K\r"
)

// Result represents a completed eval result for display.
type Result struct {
	Name     string
	Pass     bool
	Duration time.Duration
	Error    bool
}

// Bar displays a terminal progress bar for evaluation runs.
// Skipped in debug mode to avoid conflicting with verbose output.
type Bar struct {
	total     int
	workers   int
	completed int
	current   string
	results   []Result
	mu        sync.Mutex
	disabled  bool
	startTime time.Time
}

// New creates a new progress bar.
func New(total, workers int, debug bool) *Bar {
	b := &Bar{
		total:     total,
		workers:   workers,
		disabled:  debug,
		startTime: time.Now(),
	}
	if !b.disabled {
		fmt.Printf("%sRunning %d evaluations (%d workers)...%s\n", ColorBold, total, workers, ColorReset)
	}
	return b
}

// Start marks an evaluation as in-progress.
func (b *Bar) Start(name string) {
	if b.disabled {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = name
	b.render()
}

// Complete marks an evaluation as finished.
func (b *Bar) Complete(name string, pass bool, duration time.Duration, isError bool) {
	if b.disabled {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.completed++
	b.results = append(b.results, Result{
		Name:     name,
		Pass:     pass,
		Duration: duration,
		Error:    isError,
	})
	b.render()

	// Print the completed result on its own line
	icon := ColorGreen + "✅" + ColorReset
	if isError {
		icon = ColorRed + "⚠️" + ColorReset
	} else if !pass {
		icon = ColorRed + "❌" + ColorReset
	}
	fmt.Printf("%s  %s %s %s%s\n", ClearLine, icon, name, fmtDuration(duration), ColorReset)
}

// Done finalizes the progress display.
func (b *Bar) Done() {
	if b.disabled {
		return
	}
	elapsed := time.Since(b.startTime)
	passed := 0
	failed := 0
	errors := 0
	for _, r := range b.results {
		switch {
		case r.Error:
			errors++
		case r.Pass:
			passed++
		default:
			failed++
		}
	}

	fmt.Printf("\n%s%s━━━ Complete: %d/%d%s", ClearLine, ColorBold, b.completed, b.total, ColorReset)
	fmt.Printf("  %s%d passed%s", ColorGreen, passed, ColorReset)
	if failed > 0 {
		fmt.Printf("  %s%d failed%s", ColorRed, failed, ColorReset)
	}
	if errors > 0 {
		fmt.Printf("  %s%d errors%s", ColorRed, errors, ColorReset)
	}
	fmt.Printf("  %s\n", fmtDuration(elapsed))
}

func (b *Bar) render() {
	barWidth := 20
	filled := 0
	if b.total > 0 {
		filled = (b.completed * barWidth) / b.total
	}
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := ColorGreen + strings.Repeat("█", filled) + ColorReset + strings.Repeat("░", empty)

	name := b.current
	if len(name) > 40 {
		name = name[:37] + "..."
	}

	fmt.Printf("%s[%s] %d/%d  %s%s%s", ClearLine, bar, b.completed, b.total, ColorYellow, name, ColorReset)
}

func fmtDuration(d time.Duration) string {
	secs := d.Seconds()
	if secs < 60 {
		return fmt.Sprintf("%.0fs", secs)
	}
	return fmt.Sprintf("%.1fm", secs/60)
}
