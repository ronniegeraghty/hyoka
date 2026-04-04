package eval

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ronniegeraghty/hyoka/internal/config"
	"github.com/ronniegeraghty/hyoka/internal/criteria"
	"github.com/ronniegeraghty/hyoka/internal/logging"
	"github.com/ronniegeraghty/hyoka/internal/progress"
	"github.com/ronniegeraghty/hyoka/internal/prompt"
	"github.com/ronniegeraghty/hyoka/internal/report"
	"github.com/ronniegeraghty/hyoka/internal/review"
)

// EvalResult holds the raw output from a Copilot evaluation.
type EvalResult struct {
	GeneratedFiles []string
	EventCount     int
	ToolCalls      []string
	SessionEvents  []report.SessionEventRecord
	Success        bool
	Error          string
	ErrorDetails   string
	IsStub         bool
	StarterFiles   []string
}

// CopilotEvaluator defines the interface for running evaluations.
type CopilotEvaluator interface {
	Evaluate(ctx context.Context, prompt *prompt.Prompt, config *config.ToolConfig, workDir string) (*EvalResult, error)
}

// ReviewerFactory creates a reviewer for a specific config.
// Returns nil if no reviewer should be created (e.g., stub mode or review disabled).
type ReviewerFactory func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error)

// StubEvaluator returns placeholder results for testing.
type StubEvaluator struct{}

// Evaluate returns a stub result.
func (s *StubEvaluator) Evaluate(ctx context.Context, p *prompt.Prompt, cfg *config.ToolConfig, workDir string) (*EvalResult, error) {
	return &EvalResult{
		GeneratedFiles: []string{"stub_output.txt"},
		EventCount:     0,
		ToolCalls:      []string{},
		SessionEvents:  nil,
		Success:        true,
		Error:          "",
		IsStub:         true,
	}, nil
}

// EngineOptions configures the evaluation engine.
type EngineOptions struct {
	Workers      int
	MaxSessions  int // Maximum concurrent Copilot sessions (0 = workers × 2).
	OutputDir    string
	SkipTests    bool
	SkipReview   bool
	DryRun       bool
	ProgressMode string // "auto", "live", "log", "off"

	// Fan-out visibility (#34)
	ConfirmLargeRuns bool
	AutoConfirm      bool
	// Generator guardrails (#35)
	MaxSessionActions int
	MaxFiles          int
	MaxOutputSize     int64
	// Process lifecycle (#46)
	StrictCleanup bool // Fail run if orphaned processes detected after cleanup.
	// Session timeout — maximum duration for a single SendAndWait call
	// (generation or review). Defaults to 10 minutes. Per-prompt Timeout
	// frontmatter overrides this for the generation phase.
	SessionTimeout time.Duration
	// Resource monitoring (#45)
	MonitorResources bool
	// Tiered criteria (#30)
	CriteriaDir string // Directory containing attribute-matched criteria YAML files.
	// Directory exclusion (#63)
	ExcludeDirs []string // Directories to exclude from generated_files output.
	// Output writer for user-facing messages (defaults to os.Stdout).
	Stdout io.Writer
	// Tracker overrides the default process tracker (used in tests to avoid
	// killing real Copilot CLI processes during orphan scans).
	Tracker *ProcessTracker
}

// Engine orchestrates evaluation runs.
type Engine struct {
	evaluator       CopilotEvaluator
	reviewerFactory ReviewerFactory
	opts            EngineOptions
	tracker         *ProcessTracker
	criteriaSets    []criteria.CriteriaSet // Tier 2 attribute-matched criteria (#30)
}

// NewEngine creates a new Engine with the given evaluator and options.
func NewEngine(evaluator CopilotEvaluator, opts EngineOptions) *Engine {
	return NewEngineWithReviewerFactory(evaluator, nil, opts)
}

// NewEngineWithReviewer creates a new Engine with an evaluator and reviewer.
// Deprecated: Use NewEngineWithReviewerFactory instead.
func NewEngineWithReviewer(evaluator CopilotEvaluator, reviewer review.Reviewer, opts EngineOptions) *Engine {
	// Backward compatibility: wrap the single reviewer in a factory
	factory := func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error) {
		return reviewer, nil, nil
	}
	return NewEngineWithReviewerFactory(evaluator, factory, opts)
}

// NewEngineWithReviewerFactory creates a new Engine with an evaluator and reviewer factory.
func NewEngineWithReviewerFactory(evaluator CopilotEvaluator, factory ReviewerFactory, opts EngineOptions) *Engine {
	if opts.Workers <= 0 {
		w := runtime.NumCPU()
		if w > 8 {
			w = 8
		}
		opts.Workers = w
	}
	if opts.MaxSessions <= 0 {
		opts.MaxSessions = opts.Workers * 2
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./reports"
	}
	// Generator guardrail defaults (#35)
	if opts.MaxSessionActions <= 0 {
		opts.MaxSessionActions = 50
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 50
	}
	if opts.MaxOutputSize <= 0 {
		opts.MaxOutputSize = 1048576 // 1MB
	}
	if opts.SessionTimeout <= 0 {
		opts.SessionTimeout = 10 * time.Minute
	}
	// Resolve to absolute path so workspace directories passed to the Copilot CLI
	// are always absolute. Without this, the agent constructs wrong paths like
	// /home/user/reports/... instead of /home/user/projects/repo/reports/...
	if abs, err := filepath.Abs(opts.OutputDir); err == nil {
		opts.OutputDir = abs
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	tracker := opts.Tracker
	if tracker == nil {
		tracker = DefaultTracker
	}
	return &Engine{
		evaluator:       evaluator,
		reviewerFactory: factory,
		opts:            opts,
		tracker:         tracker,
	}
}

// printf writes user-facing output to the configured writer.
func (e *Engine) printf(format string, args ...any) {
	fmt.Fprintf(e.opts.Stdout, format, args...)
}

// loadCriteria loads Tier 2 criteria sets if CriteriaDir is configured.
func (e *Engine) loadCriteria() {
	if e.opts.CriteriaDir == "" {
		return
	}
	if _, err := os.Stat(e.opts.CriteriaDir); os.IsNotExist(err) {
		slog.Debug("Criteria directory does not exist, skipping", "dir", e.opts.CriteriaDir)
		return
	}
	sets, err := criteria.LoadDir(e.opts.CriteriaDir)
	if err != nil {
		slog.Warn("Failed to load criteria sets", "dir", e.opts.CriteriaDir, "error", err)
		return
	}
	e.criteriaSets = sets
	slog.Info("Loaded attribute-matched criteria", "sets", len(sets), "dir", e.opts.CriteriaDir)
}

// mergedCriteria returns the combined Tier 2 + Tier 3 evaluation criteria
// text for the given prompt.
func (e *Engine) mergedCriteria(p *prompt.Prompt) string {
	attrs := criteria.PromptAttrs{
		Language: p.Language,
		Service:  p.Service,
		Plane:    p.Plane,
		Category: p.Category,
		SDK:      p.SDKPackage,
	}
	matched := criteria.MatchingCriteria(e.criteriaSets, attrs)
	merged := criteria.MergeCriteria(matched, p.EvaluationCriteria)
	if merged == "" {
		return p.EvaluationCriteria
	}
	return merged
}

// SetPanelReviewer configures a multi-model review panel.
// Deprecated: Use NewEngineWithReviewerFactory instead.
func (e *Engine) SetPanelReviewer(pr *review.PanelReviewer) {
	// Backward compatibility: wrap the panel reviewer in a factory
	e.reviewerFactory = func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error) {
		pr.SetSessionTimeout(e.opts.SessionTimeout)
		return nil, pr, nil
	}
}

// EvalTask represents a single prompt+config evaluation to run.
type EvalTask struct {
	Prompt *prompt.Prompt
	Config config.ToolConfig
}

// Run executes evaluations for the given prompts crossed with configs.
func (e *Engine) Run(ctx context.Context, prompts []*prompt.Prompt, configs []config.ToolConfig) (*report.RunSummary, error) {
	// Load tiered criteria sets (#30) if configured.
	e.loadCriteria()

	// Build task list (cross product: prompts × configs)
	var tasks []EvalTask
	for _, p := range prompts {
		for _, c := range configs {
			tasks = append(tasks, EvalTask{Prompt: p, Config: c})
		}
	}

	// Pre-run summary (#34: fan-out visibility)
	evalCount := len(tasks)
	estimatedSessions := evalCount * 2 // generate + review per eval
	maxSessions := e.opts.Workers * 3
	slog.Info("Evaluation plan",
		"evaluations", evalCount,
		"prompts", len(prompts),
		"configs", len(configs),
		"estimated_sessions", estimatedSessions,
		"workers", e.opts.Workers,
		"max_sessions", maxSessions)
	e.printf("\n📊 Evaluation plan: %d evaluations (%d prompts × %d configs)\n", evalCount, len(prompts), len(configs))
	e.printf("   Estimated Copilot sessions: %d (%d × 2 for generate/review)\n", estimatedSessions, evalCount)
	e.printf("   Workers: %d | Max sessions: %d\n\n", e.opts.Workers, maxSessions)

	// Confirmation prompt for large runs (#34)
	if evalCount > 10 && e.opts.ConfirmLargeRuns && !e.opts.AutoConfirm {
		e.printf("⚠️  Large run detected (%d evaluations). Continue? [y/N] ", evalCount)
		var answer string
		_, _ = fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			return nil, fmt.Errorf("run aborted by user (use -y to skip confirmation)")
		}
	}

	if e.opts.DryRun {
		return e.dryRun(tasks)
	}

	// Resource monitor (#45) — opt-in via --monitor-resources.
	var resMonitor *ResourceMonitor
	if e.opts.MonitorResources {
		resMonitor = NewResourceMonitor(e.tracker, 5*time.Second)
		resMonitor.Start()
		defer resMonitor.Stop()
	}

	// Wrap context with cancel so signal handler can trigger graceful shutdown (#67).
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Ensure all tracked Copilot processes are cleaned up when Run exits.
	defer func() {
		if errs := e.tracker.TerminateAll(5 * time.Second); len(errs) > 0 {
			for _, err := range errs {
				slog.Warn("Process cleanup error", "error", err)
			}
		}
	}()

	// Set up signal handler so SIGINT/SIGTERM terminates spawned processes.
	sigCh := make(chan os.Signal, 1)
	notifyShutdownSignals(sigCh)
	// Unregister signal handler before closing the channel to prevent
	// a send-on-closed-channel panic (defers execute LIFO).
	defer close(sigCh)
	defer signal.Stop(sigCh)
	go func() {
		first := true
		for sig := range sigCh {
			if first {
				slog.Warn("Received signal — terminating tracked Copilot processes", "signal", sig.String())
				cancel() // Cancel context to unwind in-flight goroutines
				if errs := e.tracker.TerminateAll(5 * time.Second); len(errs) > 0 {
					for _, err := range errs {
						slog.Warn("Process cleanup error", "error", err)
					}
				}
				first = false
			} else {
				// Second signal: cancel context, allow brief grace period for
				// defers to run, then force exit (#67).
				slog.Warn("Received second signal — forcing exit", "signal", sig.String())
				cancel()
				time.Sleep(2 * time.Second)
				os.Exit(1)
			}
		}
	}()

	runID := time.Now().Format("20060102-150405")
	summary := &report.RunSummary{
		RunID:        runID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalPrompts: len(prompts),
		TotalConfigs: len(configs),
		TotalEvals:   len(tasks),
	}

	slog.Info("Starting run", "workers", e.opts.Workers, "max_sessions", e.opts.MaxSessions)

	start := time.Now()

	runDir := filepath.Join(e.opts.OutputDir, runID)

	// Progress display — mode is controlled by --progress flag.
	// When --log-level debug/info, main.go sets ProgressMode to "log" automatically.
	display := progress.NewDisplay(progress.DisplayConfig{
		Total:     len(tasks),
		Workers:   e.opts.Workers,
		ReportDir: runDir + "/",
		Mode:      progress.ProgressMode(e.opts.ProgressMode),
	})

	// Wire progress reporting if evaluator supports it
	if pr, ok := e.evaluator.(progress.Reporter); ok {
		pr.SetProgressFunc(display.HandleEvent)
	}

	sem := make(chan struct{}, e.opts.Workers)
	sessionSem := make(chan struct{}, e.opts.MaxSessions) // limits total concurrent Copilot sessions
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t EvalTask) {
			defer wg.Done()

			// Acquire session semaphore first to limit total Copilot sessions.
			sessionSem <- struct{}{}
			defer func() { <-sessionSem }()

			sem <- struct{}{}
			defer func() { <-sem }()

			taskName := t.Prompt.ID + "/" + t.Config.Name

			// Register eval with resource monitor if active (#45)
			if resMonitor != nil {
				resMonitor.RegisterEval(taskName)
			}

			display.HandleEvent(progress.ProgressEvent{
				EvalID:     taskName,
				PromptID:   t.Prompt.ID,
				ConfigName: t.Config.Name,
				Type:       progress.EventStarting,
				Message:    "Waiting for session...",
			})

			// Progress callbacks for runSingleEval
			sendPhase := func(phase progress.Phase) {
				display.HandleEvent(progress.ProgressEvent{
					EvalID: taskName, Type: progress.EventPhaseChange, Phase: phase,
				})
			}
			sendEvent := func(evtType progress.EventType, msg string) {
				display.HandleEvent(progress.ProgressEvent{
					EvalID: taskName, PromptID: t.Prompt.ID, ConfigName: t.Config.Name,
					Type: evtType, Message: msg,
				})
			}

			evalReport := e.runSingleEval(ctx, t, runID, sendPhase, sendEvent)

			// Attach per-eval resource stats (#45)
			if resMonitor != nil {
				if es := resMonitor.EvalStats(taskName); es != nil {
					evalReport.ResourceUsage = &report.ResourceStats{
						PeakCPUPercent: es.PeakCPUPercent,
						PeakMemoryMB:   es.PeakMemoryMB,
						SampleCount:    es.SampleCount,
					}
				}
			}

			evtType := progress.EventPassed
			msg := ""
			reviewScore := 0
			if evalReport.Review != nil {
				reviewScore = evalReport.Review.OverallScore
			}
			if evalReport.Error != "" {
				evtType = progress.EventError
				msg = "ERROR"
			} else if !evalReport.Success {
				evtType = progress.EventFailed
				if evalReport.Review != nil {
					msg = fmt.Sprintf("%d/%d criteria", evalReport.Review.OverallScore, evalReport.Review.MaxScore)
				}
			}
			display.HandleEvent(progress.ProgressEvent{
				EvalID:      taskName,
				PromptID:    t.Prompt.ID,
				ConfigName:  t.Config.Name,
				Type:        evtType,
				Message:     msg,
				FileCount:   len(evalReport.GeneratedFiles),
				ReviewScore: reviewScore,
			})

			mu.Lock()
			defer mu.Unlock()

			summary.Results = append(summary.Results, evalReport)

			if evalReport.Success {
				summary.Passed++
			} else if evalReport.Error != "" {
				summary.Errors++
			} else {
				summary.Failed++
			}
		}(task)
	}

	wg.Wait()
	display.Done()

	// Post-run orphan scan — terminate any leaked copilot processes (#46)
	// Only scan when using the DefaultTracker (production). Test-injected
	// trackers have no registrations, so every real copilot process looks
	// like an orphan — which would kill the user's Copilot CLI.
	if e.tracker == DefaultTracker {
		if orphans := e.tracker.TerminateOrphans(); orphans > 0 {
			slog.Warn("Terminated orphaned copilot processes", "count", orphans)
			if e.opts.StrictCleanup {
				return summary, fmt.Errorf("strict-cleanup: %d orphaned copilot processes detected and terminated", orphans)
			}
		}
	}

	summary.Duration = time.Since(start).Seconds()

	// Calculate per-phase average durations across all reports (#44)
	var genSum, reviewSum float64
	var genCount, reviewCount int
	for _, r := range summary.Results {
		if r.GenerationDuration > 0 {
			genSum += r.GenerationDuration
			genCount++
		}
		if r.ReviewDuration > 0 {
			reviewSum += r.ReviewDuration
			reviewCount++
		}
	}
	if genCount > 0 {
		summary.AvgGenerationDuration = genSum / float64(genCount)
	}
	if reviewCount > 0 {
		summary.AvgReviewDuration = reviewSum / float64(reviewCount)
	}

	// Attach aggregate resource stats and print summary (#45)
	if resMonitor != nil {
		rs := resMonitor.RunStats()
		summary.ResourceUsage = &report.RunResourceStats{
			PeakCPUPercent: rs.PeakCPUPercent,
			PeakMemoryMB:   rs.PeakMemoryMB,
			SessionCount:   rs.SessionCount,
		}
		e.printf("\n🔍 Resource usage: %s\n", resMonitor.SummaryLine())
	}

	// Write JSON summary
	if _, err := report.WriteSummary(summary, e.opts.OutputDir); err != nil {
		slog.Error("Failed to write run summary", "error", err)
	}

	// Write HTML summary
	if _, err := report.WriteSummaryHTML(summary, e.opts.OutputDir); err != nil {
		slog.Error("Failed to write HTML summary", "error", err)
	}

	// Write Markdown summary
	if _, err := report.WriteSummaryMarkdown(summary, e.opts.OutputDir); err != nil {
		slog.Error("Failed to write Markdown summary", "error", err)
	}

	return summary, nil
}

func (e *Engine) runSingleEval(ctx context.Context, task EvalTask, runID string, sendPhase func(progress.Phase), sendEvent func(progress.EventType, string)) *report.EvalReport {
	// Each phase gets its own independent timeout so a slow generation
	// doesn't starve build or review (fixes issue #3).
	genCtx, genCancel := context.WithCancel(ctx)
	defer genCancel()

	debugPrefix := task.Prompt.ID + "/" + task.Config.Name
	// Structured logger with eval context fields (#42)
	lg := logging.EvalLogger(task.Prompt.ID, task.Config.Name, "generation", 0)
	start := time.Now()

	evalReport := &report.EvalReport{
		PromptID:   task.Prompt.ID,
		ConfigName: task.Config.Name,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		PromptMeta: map[string]any{
			"service":     task.Prompt.Service,
			"plane":       task.Prompt.Plane,
			"language":    task.Prompt.Language,
			"category":    task.Prompt.Category,
			"description": task.Prompt.Description,
			"difficulty":  task.Prompt.Difficulty,
			"sdk_package": task.Prompt.SDKPackage,
		},
		ConfigUsed: map[string]any{
			"name":  task.Config.Name,
			"model": task.Config.Generator.Model,
		},
		// Guardrail limits recorded for report transparency (#35)
		GuardrailMaxTurns:      e.opts.MaxSessionActions,
		GuardrailMaxFiles:      e.opts.MaxFiles,
		GuardrailMaxOutputSize: e.opts.MaxOutputSize,
	}
	if len(task.Prompt.Tags) > 0 {
		evalReport.PromptMeta["tags"] = strings.Join(task.Prompt.Tags, ", ")
	}

	lg.Info("Starting Copilot session")

	// Build the report directory path early — workspace lives in the report tree (Issue 2)
	reportDir := filepath.Join(report.ReportDir(e.opts.OutputDir, runID, task.Prompt), task.Config.Name)
	codeDir := filepath.Join(reportDir, "generated-code")

	ws, err := NewWorkspaceAt(codeDir)
	if err != nil {
		evalReport.Error = fmt.Sprintf("workspace setup failed: %v", err)
		evalReport.ErrorDetails = err.Error()
		evalReport.ErrorCategory = "generation_failure"
		evalReport.FailureReason = fmt.Sprintf("Could not create workspace directory: %v", err)
		evalReport.Duration = time.Since(start).Seconds()
		return evalReport
	}

	// Create an isolated temporary workspace for the generator (#26).
	// The agent writes files here — not directly to the report tree.
	// After generation, files are copied to the persistent report directory.
	genDir, err := os.MkdirTemp("", "hyoka-gen-*")
	if err != nil {
		evalReport.Error = fmt.Sprintf("generator workspace setup failed: %v", err)
		evalReport.ErrorDetails = err.Error()
		evalReport.ErrorCategory = "generation_failure"
		evalReport.FailureReason = fmt.Sprintf("Could not create isolated generator workspace: %v", err)
		evalReport.Duration = time.Since(start).Seconds()
		return evalReport
	}
	defer os.RemoveAll(genDir)

	lg.Debug("Workspace created", "workspace", ws.Dir, "gen_dir", genDir)

	// Snapshot home directory and CWD before eval so we can recover misplaced files after
	homeDir, _ := os.UserHomeDir()
	var preEvalHomeFiles map[string]bool
	if homeDir != "" {
		preEvalHomeFiles = snapshotDir(homeDir)
	}
	// Also snapshot CWD — agents may write files relative to the process working directory
	cwdDir, _ := os.Getwd()
	var preEvalCwdFiles map[string]bool
	if cwdDir != "" && cwdDir != homeDir && cwdDir != ws.Dir {
		preEvalCwdFiles = snapshotDir(cwdDir)
	}

	// Run evaluation (generation phase — uses its own timeout)
	sendPhase(progress.PhaseGenerating)

	genStart := time.Now()
	result, err := e.evaluator.Evaluate(genCtx, task.Prompt, &task.Config, genDir)
	genCancel() // release generation context immediately
	evalReport.GenerationDuration = time.Since(genStart).Seconds()
	evalFailed := err != nil
	if evalFailed {
		if genCtx.Err() == context.Canceled {
			evalReport.Error = "generation cancelled (action limit reached)"
			evalReport.ErrorDetails = "context cancelled — the session exceeded the --max-session-actions limit"
			evalReport.ErrorCategory = "timeout"
			evalReport.FailureReason = "Generation cancelled due to action limit"
		} else {
			evalReport.Error = fmt.Sprintf("evaluation failed: %v", err)
			evalReport.ErrorDetails = err.Error()
			evalReport.ErrorCategory = "sdk_error"
			evalReport.FailureReason = fmt.Sprintf("SDK evaluation error: %v", err)
		}
		// Capture whatever session events were collected before failure
		if result != nil {
			evalReport.SessionEvents = result.SessionEvents
			evalReport.EventCount = result.EventCount
			evalReport.ToolCalls = result.ToolCalls
			evalReport.IsStub = result.IsStub
		}
		// Don't return early — continue to collect files and run review for diagnostics
	}

	if result != nil && !evalFailed {
		evalReport.EventCount = result.EventCount
		evalReport.ToolCalls = result.ToolCalls
		evalReport.SessionEvents = result.SessionEvents
		evalReport.IsStub = result.IsStub
		evalReport.Success = result.Success
		evalReport.StarterFiles = result.StarterFiles
	}

	// Collect generated files — workspace listing is the primary source since
	// ForceStop preserves files on disk.
	// First, recover any files the agent wrote to the home directory instead of the workspace.
	// The Copilot CLI sometimes creates files in ~ when the agent omits the path parameter.
	if homeDir != "" && preEvalHomeFiles != nil {
		recovered := recoverMisplacedFiles(homeDir, preEvalHomeFiles, genDir, debugPrefix)
		if recovered > 0 {
			lg.Info("Recovered misplaced files from home dir", "count", recovered)
		}
		// Post-recovery validation: flag anything recovery couldn't handle (#26)
		if remaining := ValidateWorkspaceContainment(homeDir, preEvalHomeFiles); len(remaining) > 0 {
			lg.Warn("Items still outside workspace after recovery (home)", "count", len(remaining), "items", remaining)
		}
	}
	// Also recover from CWD
	if cwdDir != "" && preEvalCwdFiles != nil {
		recovered := recoverMisplacedFiles(cwdDir, preEvalCwdFiles, genDir, debugPrefix)
		if recovered > 0 {
			lg.Info("Recovered misplaced files from CWD", "count", recovered)
		}
		if remaining := ValidateWorkspaceContainment(cwdDir, preEvalCwdFiles); len(remaining) > 0 {
			lg.Warn("Items still outside workspace after recovery (CWD)", "count", len(remaining), "items", remaining)
		}
	}

	// Copy generated files from isolated workspace to persistent report directory (#26)
	if err := copyDir(genDir, ws.Dir); err != nil {
		lg.Warn("Failed to copy generated files to report dir", "error", err)
	}

	generatedFiles, _ := ws.ListFiles()
	if len(generatedFiles) == 0 && result != nil && len(result.GeneratedFiles) > 0 {
		generatedFiles = result.GeneratedFiles
	}
	// Apply directory exclusion filter (#63)
	if len(e.opts.ExcludeDirs) > 0 {
		before := len(generatedFiles)
		generatedFiles = filterExcludedDirs(generatedFiles, e.opts.ExcludeDirs)
		if excluded := before - len(generatedFiles); excluded > 0 {
			lg.Debug("Excluded files by directory filter", "excluded", excluded, "remaining", len(generatedFiles))
		}
	}
	evalReport.GeneratedFiles = generatedFiles

	// Diagnostic: if 0 files generated, check if agent attempted file creation
	if len(generatedFiles) == 0 && !evalFailed {
		fileToolAttempts := 0
		for _, ev := range evalReport.SessionEvents {
			if ev.Type == "tool.execution_start" && isFileWriteTool(ev.ToolName) {
				fileToolAttempts++
			}
		}
		if fileToolAttempts > 0 {
			lg.Warn("0 files generated despite file-write tool attempts", "attempts", fileToolAttempts)
			if evalReport.Error == "" {
				evalReport.Error = fmt.Sprintf("0 files generated despite %d file-write tool attempts", fileToolAttempts)
				evalReport.ErrorCategory = "no_files"
				evalReport.FailureReason = fmt.Sprintf("Generator made %d file-write attempts but no files appeared in the workspace — files may have been written to the wrong location", fileToolAttempts)
				evalReport.Success = false
			}
		} else {
			lg.Warn("0 files generated — agent did not use any file-write tools")
			if evalReport.Error == "" {
				evalReport.Error = "0 files generated — agent did not create any files"
				evalReport.ErrorCategory = "no_files"
				evalReport.FailureReason = "Generator produced no files — the agent did not invoke any file-write tools"
				evalReport.Success = false
			}
		}
	}

	lg.Debug("Session complete",
		"tool_calls", len(evalReport.ToolCalls),
		"files_generated", len(generatedFiles),
		"elapsed", time.Since(start).Truncate(time.Millisecond).String())

	// Per-phase generation duration already captured above (evalReport.GenerationDuration).
	// Overall Duration is set at the end of the function after all phases complete.

	// Populate environment info from config and captured events
	var skillDirectories []string
	if task.Config.Generator != nil {
		for _, s := range task.Config.Generator.Skills {
			if s.Type == "local" && s.Path != "" {
				skillDirectories = append(skillDirectories, s.Path)
			}
		}
	}
	env := &report.EnvironmentInfo{
		Model:            task.Config.Generator.Model,
		SkillDirectories: skillDirectories,
		AvailableTools:   task.Config.Generator.AvailableTools,
		ExcludedTools:    task.Config.Generator.ExcludedTools,
		SafetyBoundaries: true,
		AllowCloud:       false,
		WorkingDirectory: ws.Dir,
	}
	// Extract MCP server names
	for name := range task.Config.Generator.MCPServers {
		env.MCPServers = append(env.MCPServers, name)
	}
	// Derive token usage, turn count, truncation, skills from events
	for _, ev := range evalReport.SessionEvents {
		switch ev.Type {
		case "assistant.usage":
			env.TotalInputTokens += ev.InputTokens
			env.TotalOutputTokens += ev.OutputTokens
		case "assistant.turn_start":
			env.TurnCount++
		case "session.truncation":
			env.ContextTruncated = true
		case "skill.invoked":
			if ev.SkillName != "" {
				env.SkillsInvoked = append(env.SkillsInvoked, ev.SkillName)
			}
		case "session.skills_loaded":
			if ev.Content != "" {
				env.SkillsLoaded = strings.Split(ev.Content, ", ")
			}
		}
	}
	evalReport.Environment = env

	// Generator guardrail checks (#35)
	if !evalFailed {
		// Check action count (count reasoning, message, and tool_execution_start events as actions)
		actionCount := 0
		for _, ev := range evalReport.SessionEvents {
			if ev.Type == "assistant.reasoning" || ev.Type == "assistant.message" || ev.Type == "tool.execution_start" {
				actionCount++
			}
		}
		if actionCount > e.opts.MaxSessionActions {
			reason := fmt.Sprintf("guardrail: action count %d exceeded limit of %d", actionCount, e.opts.MaxSessionActions)
			evalReport.GuardrailAbortReason = reason
			evalReport.Error = reason
			evalReport.Success = false
			lg.Warn("Guardrail triggered", "reason", reason, "actions", actionCount, "max_session_actions", e.opts.MaxSessionActions)
		}

		// Check file count
		if len(generatedFiles) > e.opts.MaxFiles {
			reason := fmt.Sprintf("guardrail: file count %d exceeded limit of %d", len(generatedFiles), e.opts.MaxFiles)
			evalReport.GuardrailAbortReason = reason
			evalReport.Error = reason
			evalReport.Success = false
			lg.Warn("Guardrail triggered", "reason", reason, "files", len(generatedFiles), "max_files", e.opts.MaxFiles)
		}

		// Check total output size
		var totalSize int64
		for _, f := range generatedFiles {
			absPath := f
			if !filepath.IsAbs(f) {
				absPath = filepath.Join(ws.Dir, f)
			}
			if info, err := os.Stat(absPath); err == nil {
				totalSize += info.Size()
			}
		}
		if totalSize > e.opts.MaxOutputSize {
			reason := fmt.Sprintf("guardrail: total output size %d bytes exceeded limit of %d bytes", totalSize, e.opts.MaxOutputSize)
			evalReport.GuardrailAbortReason = reason
			evalReport.Error = reason
			evalReport.Success = false
			lg.Warn("Guardrail triggered", "reason", reason, "total_size", totalSize, "max_size", e.opts.MaxOutputSize)
		}
	}

	// Code review — use panel reviewer if available, otherwise single reviewer
	// Uses its own independent timeout context (fixes issue #3).
	if !e.opts.SkipReview && len(generatedFiles) > 0 {
		reviewStart := time.Now()
		sendPhase(progress.PhaseReviewing)
		rlg := logging.WithPhase(lg, "review")

		// Create an isolated reviewer workspace with a copy of the generated
		// files. Reviewers operate on this copy and cannot modify the original
		// output in the report directory (#26).
		reviewWorkDir, err := NewReviewerWorkspace(ws.Dir)
		if err != nil {
			rlg.Warn("Reviewer workspace creation failed, using original", "error", err)
			reviewWorkDir = ws.Dir
		} else {
			defer os.RemoveAll(reviewWorkDir)
		}

		referenceDir := ""
		if task.Prompt.ReferenceAnswer != "" {
			referenceDir = task.Prompt.ReferenceAnswer
		}

		// Merge tiered evaluation criteria (#30)
		evalCriteria := e.mergedCriteria(task.Prompt)

		// Create reviewer for this specific config using the factory (#92)
		var reviewer review.Reviewer
		var panelReviewer *review.PanelReviewer
		if e.reviewerFactory != nil {
			reviewer, panelReviewer, err = e.reviewerFactory(&task.Config)
			if err != nil {
				rlg.Warn("Reviewer creation failed, skipping review", "error", err)
			}
		}

		if panelReviewer != nil {
			models := panelReviewer.Models()
			rlg.Debug("Starting review panel")
			sendEvent(progress.EventToolStart, fmt.Sprintf("Review panel: %v", models))
			panel, consolidated, err := panelReviewer.ReviewPanel(ctx, task.Prompt.PromptText, reviewWorkDir, referenceDir, evalCriteria)
			if err != nil {
				rlg.Error("Review panel failed", "error", err)
				sendEvent(progress.EventReasoning, fmt.Sprintf("Review panel failed: %v", err))
			} else {
				evalReport.ReviewPanel = panel
				evalReport.Review = consolidated
				// With criteria-based scoring, success = all criteria passed
				if !evalFailed {
					evalReport.Success = consolidated.Scores.AllPassed()
				}
				sendEvent(progress.EventToolComplete, fmt.Sprintf("Review complete: %d/%d criteria passed", consolidated.OverallScore, consolidated.MaxScore))
				rlg.Debug("Review panel complete",
					"reviewers", len(panel),
					"score", consolidated.OverallScore,
					"max_score", consolidated.MaxScore)
			}
		} else if reviewer != nil {
			rlg.Debug("Starting single review session")
			sendEvent(progress.EventToolStart, "Single model review")
			reviewResult, err := reviewer.Review(ctx, task.Prompt.PromptText, reviewWorkDir, referenceDir, evalCriteria)
			if err != nil {
				rlg.Error("Code review failed", "error", err)
				sendEvent(progress.EventReasoning, fmt.Sprintf("Review failed: %v", err))
			} else {
				evalReport.Review = reviewResult
				// With criteria-based scoring, success = all criteria passed
				if !evalFailed {
					evalReport.Success = reviewResult.Scores.AllPassed()
				}
				sendEvent(progress.EventToolComplete, fmt.Sprintf("Review complete: %d/%d criteria passed", reviewResult.OverallScore, reviewResult.MaxScore))
				rlg.Debug("Review complete",
					"score", reviewResult.OverallScore,
					"max_score", reviewResult.MaxScore)
			}
		}

		// Capture reviewed (annotated) files from the reviewer workspace
		reviewedFiles, err := readReviewedFiles(reviewWorkDir)
		if err == nil && len(reviewedFiles) > 0 {
			evalReport.ReviewedFiles = reviewedFiles
			rlg.Debug("Captured reviewed files", "count", len(reviewedFiles))
		}
		evalReport.ReviewDuration = time.Since(reviewStart).Seconds()
	}

	// Tool usage evaluation (compare expected vs actual tools)
	if len(task.Prompt.ExpectedTools) > 0 {
		evalReport.ToolUsage = evaluateToolUsage(task.Prompt.ExpectedTools, evalReport.ToolCalls)
		lg.Debug("Tool usage evaluated",
			"match", evalReport.ToolUsage.Match,
			"matched", evalReport.ToolUsage.MatchedTools,
			"missing", evalReport.ToolUsage.MissingTools)
	}

	// Copy reviewed (annotated) files into report under reviewed-code/
	if len(evalReport.ReviewedFiles) > 0 {
		reviewedDir := filepath.Join(reportDir, "reviewed-code")
		if err := writeReviewedFiles(reviewedDir, evalReport.ReviewedFiles); err != nil {
			lg.Error("Failed to write reviewed files", "error", err)
		} else {
			lg.Debug("Wrote reviewed files", "count", len(evalReport.ReviewedFiles), "dir", reviewedDir)
		}
	}

	// Build re-run command so users can reproduce this evaluation
	evalReport.RerunCommand = buildRerunCommand(task.Prompt.ID, task.Config.Name, e.opts)

	// Capture overall duration after all phases (generation, build, review) complete.
	evalReport.Duration = time.Since(start).Seconds()

	// Write JSON report
	reportPath, err := report.WriteReport(evalReport, e.opts.OutputDir, runID, task.Prompt)
	if err != nil {
		lg.Error("Failed to write report", "error", err)
	} else {
		lg.Debug("Report written", "path", reportPath)
	}

	// Write HTML report
	if _, err := report.WriteHTMLReport(evalReport, e.opts.OutputDir, runID,
		task.Prompt.Service, task.Prompt.Plane, task.Prompt.Language, task.Prompt.Category); err != nil {
		lg.Error("Failed to write HTML report", "error", err)
	}

	// Write Markdown report
	if _, err := report.WriteMarkdownReport(evalReport, e.opts.OutputDir, runID,
		task.Prompt.Service, task.Prompt.Plane, task.Prompt.Language, task.Prompt.Category); err != nil {
		lg.Error("Failed to write Markdown report", "error", err)
	}

	lg.Info("Evaluation complete",
		"success", evalReport.Success,
		"files_generated", len(evalReport.GeneratedFiles),
		"elapsed", fmt.Sprintf("%.2fs", evalReport.Duration))

	return evalReport
}

// buildRerunCommand constructs the CLI command to reproduce a single evaluation.
func buildRerunCommand(promptID, configName string, opts EngineOptions) string {
	parts := []string{"hyoka run"}
	parts = append(parts, "--prompt-id", promptID)
	parts = append(parts, "--config", configName)

	if opts.SkipTests {
		parts = append(parts, "--skip-tests")
	}
	if opts.SkipReview {
		parts = append(parts, "--skip-review")
	}
	if opts.MonitorResources {
		parts = append(parts, "--monitor-resources")
	}

	if opts.MaxSessionActions != 50 {
		parts = append(parts, fmt.Sprintf("--max-session-actions=%d", opts.MaxSessionActions))
	}

	return strings.Join(parts, " ")
}

func (e *Engine) dryRun(tasks []EvalTask) (*report.RunSummary, error) {
	summary := &report.RunSummary{
		RunID:        "dry-run",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalPrompts: 0,
		TotalConfigs: 0,
		TotalEvals:   len(tasks),
	}

	promptIDs := make(map[string]bool)
	configNames := make(map[string]bool)

	for _, t := range tasks {
		promptIDs[t.Prompt.ID] = true
		configNames[t.Config.Name] = true
	}

	summary.TotalPrompts = len(promptIDs)
	summary.TotalConfigs = len(configNames)

	return summary, nil
}

// evaluateToolUsage compares expected tools from prompt frontmatter with actual tool calls.
func evaluateToolUsage(expected, actual []string) *report.ToolUsageResult {
	actualSet := make(map[string]bool, len(actual))
	for _, t := range actual {
		actualSet[t] = true
	}

	var matched, missing []string
	expectedSet := make(map[string]bool, len(expected))
	for _, t := range expected {
		expectedSet[t] = true
		if actualSet[t] {
			matched = append(matched, t)
		} else {
			missing = append(missing, t)
		}
	}

	var extra []string
	for _, t := range actual {
		if !expectedSet[t] {
			extra = append(extra, t)
		}
	}

	return &report.ToolUsageResult{
		ExpectedTools: expected,
		ActualTools:   actual,
		MatchedTools:  matched,
		MissingTools:  missing,
		ExtraTools:    extra,
		Match:         len(missing) == 0,
	}
}

// readReviewedFiles reads all files from the workspace and returns them as ReviewedFile entries.
// Files that contain "REVIEW:" comments are considered annotated.
func readReviewedFiles(dir string) ([]report.ReviewedFile, error) {
	var reviewed []report.ReviewedFile
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), ".") || info.Size() > 1<<20 {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		if strings.Contains(content, "REVIEW:") {
			reviewed = append(reviewed, report.ReviewedFile{
				Path:    rel,
				Content: content,
			})
		}
		return nil
	})
	return reviewed, err
}

// writeReviewedFiles writes annotated files to the reviewed-code directory.
func writeReviewedFiles(dir string, files []report.ReviewedFile) error {
	for _, f := range files {
		dst := filepath.Join(dir, f.Path)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("creating dir for %s: %w", f.Path, err)
		}
		if err := os.WriteFile(dst, []byte(f.Content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", f.Path, err)
		}
	}
	return nil
}
