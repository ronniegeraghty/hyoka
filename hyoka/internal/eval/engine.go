package eval

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ronniegeraghty/hyoka/internal/build"
	"github.com/ronniegeraghty/hyoka/internal/config"
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
	Workers         int
	MaxSessions     int           // Maximum concurrent Copilot sessions (0 = workers × 3).
	Timeout         time.Duration // Deprecated: use GenerateTimeout. Kept for backward compat.
	GenerateTimeout time.Duration // Independent timeout for code generation phase.
	VerifyTimeout   time.Duration // Independent timeout for verification phase.
	ReviewTimeout   time.Duration // Independent timeout for review phase.
	OutputDir       string
	SkipTests       bool
	SkipReview      bool
	VerifyBuild     bool
	Debug           bool
	DryRun          bool
	ProgressMode    string // "auto", "live", "log", "off"

	// Fan-out visibility (#34)
	ConfirmLargeRuns bool
	AutoConfirm      bool
	// Generator guardrails (#35)
	MaxTurns      int
	MaxFiles      int
	MaxOutputSize int64
}

// Verifier evaluates generated code against prompt requirements.
type Verifier interface {
	Verify(ctx context.Context, originalPrompt string, workDir string, evaluationCriteria string) (*report.VerifyResult, error)
}

// StubVerifier returns a placeholder pass result.
type StubVerifier struct{}

// Verify returns a stub verification pass.
func (s *StubVerifier) Verify(_ context.Context, _ string, _ string, _ string) (*report.VerifyResult, error) {
	return &report.VerifyResult{
		Pass:      true,
		Reasoning: "Verification skipped (stub mode)",
		Summary:   "Stub mode — no Copilot verification performed",
	}, nil
}

// Engine orchestrates evaluation runs.
type Engine struct {
	evaluator      CopilotEvaluator
	reviewer       review.Reviewer
	panelReviewer  *review.PanelReviewer
	verifier       Verifier
	opts           EngineOptions
}

// NewEngine creates a new Engine with the given evaluator and options.
func NewEngine(evaluator CopilotEvaluator, opts EngineOptions) *Engine {
	return NewEngineWithReviewer(evaluator, nil, nil, opts)
}

// NewEngineWithReviewer creates a new Engine with an evaluator, verifier, and reviewer.
func NewEngineWithReviewer(evaluator CopilotEvaluator, verifier Verifier, reviewer review.Reviewer, opts EngineOptions) *Engine {
	if opts.Workers <= 0 {
		w := runtime.NumCPU()
		if w > 8 {
			w = 8
		}
		opts.Workers = w
	}
	if opts.MaxSessions <= 0 {
		opts.MaxSessions = opts.Workers * 3
	}
	// Backward compat: if only the legacy Timeout is set, use it as GenerateTimeout.
	if opts.Timeout > 0 && opts.GenerateTimeout <= 0 {
		opts.GenerateTimeout = opts.Timeout
	}
	if opts.GenerateTimeout <= 0 {
		opts.GenerateTimeout = 10 * time.Minute
	}
	if opts.VerifyTimeout <= 0 {
		opts.VerifyTimeout = 5 * time.Minute
	}
	if opts.ReviewTimeout <= 0 {
		opts.ReviewTimeout = 5 * time.Minute
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./reports"
	}
	// Generator guardrail defaults (#35)
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = 25
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 50
	}
	if opts.MaxOutputSize <= 0 {
		opts.MaxOutputSize = 1048576 // 1MB
	}
	// Resolve to absolute path so workspace directories passed to the Copilot CLI
	// are always absolute. Without this, the agent constructs wrong paths like
	// /home/user/reports/... instead of /home/user/projects/repo/reports/...
	if abs, err := filepath.Abs(opts.OutputDir); err == nil {
		opts.OutputDir = abs
	}
	return &Engine{
		evaluator: evaluator,
		reviewer:  reviewer,
		verifier:  verifier,
		opts:      opts,
	}
}

// SetPanelReviewer configures a multi-model review panel.
// When set, the engine uses the panel instead of the single reviewer.
func (e *Engine) SetPanelReviewer(pr *review.PanelReviewer) {
	e.panelReviewer = pr
}

// EvalTask represents a single prompt+config evaluation to run.
type EvalTask struct {
	Prompt *prompt.Prompt
	Config config.ToolConfig
}

// Run executes evaluations for the given prompts crossed with configs.
func (e *Engine) Run(ctx context.Context, prompts []*prompt.Prompt, configs []config.ToolConfig) (*report.RunSummary, error) {
	// Build task list (cross product: prompts × configs)
	var tasks []EvalTask
	for _, p := range prompts {
		for _, c := range configs {
			tasks = append(tasks, EvalTask{Prompt: p, Config: c})
		}
	}

	if e.opts.DryRun {
		return e.dryRun(tasks)
	}

	// Ensure all tracked Copilot processes are cleaned up when Run exits.
	defer func() {
		if errs := DefaultTracker.TerminateAll(5 * time.Second); len(errs) > 0 {
			for _, err := range errs {
				log.Printf("[WARN] process cleanup error: %v", err)
			}
		}
	}()

	// Set up signal handler so SIGINT/SIGTERM terminates spawned processes.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}
		log.Printf("[WARN] Received %v — terminating tracked Copilot processes...", sig)
		if errs := DefaultTracker.TerminateAll(5 * time.Second); len(errs) > 0 {
			for _, err := range errs {
				log.Printf("[WARN] process cleanup error: %v", err)
			}
		}
	}()
	defer signal.Stop(sigCh)
	defer close(sigCh)

	runID := time.Now().Format("20060102-150405")
	summary := &report.RunSummary{
		RunID:        runID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalPrompts: len(prompts),
		TotalConfigs: len(configs),
		TotalEvals:   len(tasks),
	}

	log.Printf("Starting run: %d workers, %d max sessions", e.opts.Workers, e.opts.MaxSessions)

	start := time.Now()

	runDir := filepath.Join(e.opts.OutputDir, runID)

	// Progress display (disabled in debug mode or when stdout is not a terminal)
	display := progress.NewDisplay(progress.DisplayConfig{
		Total:     len(tasks),
		Workers:   e.opts.Workers,
		Disabled:  e.opts.Debug,
		ReportDir: runDir + "/",
		Mode:      progress.ProgressMode(e.opts.ProgressMode),
	})

	// Wire progress reporting if evaluator supports it
	if pr, ok := e.evaluator.(progress.Reporter); ok && !e.opts.Debug {
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
			display.HandleEvent(progress.ProgressEvent{
				EvalID:     taskName,
				PromptID:   t.Prompt.ID,
				ConfigName: t.Config.Name,
				Type:       progress.EventStarting,
				Message:    "Waiting for session...",
			})

			// Progress callback for phase transitions within runSingleEval
			sendPhase := func(phase progress.Phase) {
				display.HandleEvent(progress.ProgressEvent{
					EvalID: taskName, Type: progress.EventPhaseChange, Phase: phase,
				})
			}

			evalReport := e.runSingleEval(ctx, t, runID, sendPhase)

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
				if evalReport.Verification != nil && !evalReport.Verification.Pass {
					msg = "verification failed"
				}
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

	summary.Duration = time.Since(start).Seconds()

	// Write JSON summary
	if _, err := report.WriteSummary(summary, e.opts.OutputDir); err != nil {
		log.Printf("failed to write run summary: %v", err)
	}

	// Write HTML summary
	if _, err := report.WriteSummaryHTML(summary, e.opts.OutputDir); err != nil {
		log.Printf("failed to write HTML summary: %v", err)
	}

	// Write Markdown summary
	if _, err := report.WriteSummaryMarkdown(summary, e.opts.OutputDir); err != nil {
		log.Printf("failed to write Markdown summary: %v", err)
	}

	return summary, nil
}

func (e *Engine) runSingleEval(ctx context.Context, task EvalTask, runID string, sendPhase func(progress.Phase)) *report.EvalReport {
	// Each phase gets its own independent timeout so a slow generation
	// doesn't starve verification or review (fixes issue #3).
	genCtx, genCancel := context.WithTimeout(ctx, e.opts.GenerateTimeout)
	defer genCancel()

	debugPrefix := task.Prompt.ID + "/" + task.Config.Name
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
			"model": task.Config.Model,
		},
	}
	if len(task.Prompt.Tags) > 0 {
		evalReport.PromptMeta["tags"] = strings.Join(task.Prompt.Tags, ", ")
	}

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Starting Copilot session...", debugPrefix)
	}

	// Build the report directory path early — workspace lives in the report tree (Issue 2)
	reportDir := filepath.Join(report.ReportDir(e.opts.OutputDir, runID, task.Prompt), task.Config.Name)
	codeDir := filepath.Join(reportDir, "generated-code")

	ws, err := NewWorkspaceAt(codeDir)
	if err != nil {
		evalReport.Error = fmt.Sprintf("workspace setup failed: %v", err)
		evalReport.ErrorDetails = err.Error()
		evalReport.Duration = time.Since(start).Seconds()
		return evalReport
	}

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Workspace: %s", debugPrefix, ws.Dir)
	}

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
	result, err := e.evaluator.Evaluate(genCtx, task.Prompt, &task.Config, ws.Dir)
	genCancel() // release generation timeout immediately
	evalFailed := err != nil
	if evalFailed {
		if genCtx.Err() == context.DeadlineExceeded {
			evalReport.Error = fmt.Sprintf("generation timed out after %s", e.opts.GenerateTimeout)
			evalReport.ErrorDetails = fmt.Sprintf("context deadline exceeded — consider increasing --generate-timeout (currently %s)", e.opts.GenerateTimeout)
		} else {
			evalReport.Error = fmt.Sprintf("evaluation failed: %v", err)
			evalReport.ErrorDetails = err.Error()
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
		recovered := recoverMisplacedFiles(homeDir, preEvalHomeFiles, ws.Dir, debugPrefix, e.opts.Debug)
		if recovered > 0 {
			log.Printf("%s: Recovered %d misplaced files from home dir to workspace", debugPrefix, recovered)
		}
	}
	// Also recover from CWD
	if cwdDir != "" && preEvalCwdFiles != nil {
		recovered := recoverMisplacedFiles(cwdDir, preEvalCwdFiles, ws.Dir, debugPrefix, e.opts.Debug)
		if recovered > 0 {
			log.Printf("%s: Recovered %d misplaced files from CWD to workspace", debugPrefix, recovered)
		}
	}

	generatedFiles, _ := ws.ListFiles()
	if len(generatedFiles) == 0 && result != nil && len(result.GeneratedFiles) > 0 {
		generatedFiles = result.GeneratedFiles
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
			log.Printf("WARNING %s: 0 files generated despite %d file-write tool attempts — files may have been written to wrong location", debugPrefix, fileToolAttempts)
			if evalReport.Error == "" {
				evalReport.Error = fmt.Sprintf("0 files generated despite %d file-write tool attempts", fileToolAttempts)
				evalReport.Success = false
			}
		} else {
			log.Printf("WARNING %s: 0 files generated — agent did not use any file-write tools", debugPrefix)
			if evalReport.Error == "" {
				evalReport.Error = "0 files generated — agent did not create any files"
				evalReport.Success = false
			}
		}
	}

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Session complete: %d tool calls, %d files generated, %s",
			debugPrefix, len(evalReport.ToolCalls), len(generatedFiles), time.Since(start).Truncate(time.Millisecond))
	}

	// Capture generation duration BEFORE review/verification so it only reflects
	// the time the generator agent took, not the additional review time.
	evalReport.Duration = time.Since(start).Seconds()

	// Copilot-based verification (skip if eval hard-failed with no files)
	// Uses its own independent timeout context (fixes issue #3).
	if e.verifier != nil && len(generatedFiles) > 0 {
		sendPhase(progress.PhaseVerifying)
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: Starting verification session...", debugPrefix)
		}
		verifyCtx, verifyCancel := context.WithTimeout(ctx, e.opts.VerifyTimeout)
		verifyResult, err := e.verifier.Verify(verifyCtx, task.Prompt.PromptText, ws.Dir, task.Prompt.EvaluationCriteria)
		verifyCancel()
		if err != nil {
			log.Printf("%s: verification error: %v", debugPrefix, err)
			if evalReport.Error == "" {
				evalReport.Error = fmt.Sprintf("verification error: %v", err)
				evalReport.ErrorDetails = err.Error()
			}
			evalReport.Success = false
		} else {
			evalReport.Verification = verifyResult
			if !evalFailed {
				evalReport.Success = verifyResult.Pass
			}
			if e.opts.Debug {
				passStr := "FAIL"
				if verifyResult.Pass {
					passStr = "PASS"
				}
				log.Printf("[DEBUG] %s: Verification: %s — %s", debugPrefix, passStr, verifyResult.Summary)
			}
		}
	}

	// Optional build verification (--verify-build flag)
	if e.opts.VerifyBuild && len(generatedFiles) > 0 {
		buildCtx, buildCancel := context.WithTimeout(ctx, e.opts.VerifyTimeout)
		buildResult, err := build.Verify(buildCtx, task.Prompt.Language, ws.Dir)
		buildCancel()
		if err != nil {
			log.Printf("%s: build verification error: %v", debugPrefix, err)
			if evalReport.Error == "" {
				evalReport.Error = fmt.Sprintf("build verification error: %v", err)
				evalReport.ErrorDetails = err.Error()
			}
			evalReport.Success = false
		} else {
			evalReport.Build = buildResult
			if !buildResult.Success {
				evalReport.Success = false
			}
		}
	}

	// Code review — use panel reviewer if available, otherwise single reviewer
	// Uses its own independent timeout context (fixes issue #3).
	if !e.opts.SkipReview && len(generatedFiles) > 0 {
		sendPhase(progress.PhaseReviewing)
		reviewCtx, reviewCancel := context.WithTimeout(ctx, e.opts.ReviewTimeout)
		referenceDir := ""
		if task.Prompt.ReferenceAnswer != "" {
			referenceDir = task.Prompt.ReferenceAnswer
		}

		if e.panelReviewer != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: Starting review panel...", debugPrefix)
			}
			panel, consolidated, err := e.panelReviewer.ReviewPanel(reviewCtx, task.Prompt.PromptText, ws.Dir, referenceDir, task.Prompt.EvaluationCriteria)
			if err != nil {
				if e.opts.Debug {
					log.Printf("[DEBUG] %s: ERROR: review panel failed: %v", debugPrefix, err)
				}
			} else {
				evalReport.ReviewPanel = panel
				evalReport.Review = consolidated
				// With criteria-based scoring, success = all criteria passed
				if !evalFailed {
					evalReport.Success = consolidated.Scores.AllPassed()
				}
				if e.opts.Debug {
					log.Printf("[DEBUG] %s: Review panel: %d reviewers, consensus score: %d/%d criteria",
						debugPrefix, len(panel), consolidated.OverallScore, consolidated.MaxScore)
				}
			}
		} else if e.reviewer != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: Starting single review session...", debugPrefix)
			}
			reviewResult, err := e.reviewer.Review(reviewCtx, task.Prompt.PromptText, ws.Dir, referenceDir, task.Prompt.EvaluationCriteria)
			if err != nil {
				if e.opts.Debug {
					log.Printf("[DEBUG] %s: ERROR: code review failed: %v", debugPrefix, err)
				}
			} else {
				evalReport.Review = reviewResult
				// With criteria-based scoring, success = all criteria passed
				if !evalFailed {
					evalReport.Success = reviewResult.Scores.AllPassed()
				}
				if e.opts.Debug {
					log.Printf("[DEBUG] %s: Review score: %d/%d criteria", debugPrefix, reviewResult.OverallScore, reviewResult.MaxScore)
				}
			}
		}

		// Capture reviewed (annotated) files
		reviewedFiles, err := readReviewedFiles(ws.Dir)
		if err == nil && len(reviewedFiles) > 0 {
			evalReport.ReviewedFiles = reviewedFiles
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: Captured %d reviewed files with annotations", debugPrefix, len(reviewedFiles))
			}
		}
		reviewCancel()
	}

	// Tool usage evaluation (compare expected vs actual tools)
	if len(task.Prompt.ExpectedTools) > 0 {
		evalReport.ToolUsage = evaluateToolUsage(task.Prompt.ExpectedTools, evalReport.ToolCalls)
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: Tool usage: match=%v, matched=%v, missing=%v",
				debugPrefix, evalReport.ToolUsage.Match, evalReport.ToolUsage.MatchedTools, evalReport.ToolUsage.MissingTools)
		}
	}

	// Copy reviewed (annotated) files into report under reviewed-code/
	if len(evalReport.ReviewedFiles) > 0 {
		reviewedDir := filepath.Join(reportDir, "reviewed-code")
		if err := writeReviewedFiles(reviewedDir, evalReport.ReviewedFiles); err != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: ERROR: failed to write reviewed files: %v", debugPrefix, err)
			}
		} else if e.opts.Debug {
			log.Printf("[DEBUG] %s: Wrote %d reviewed files to %s", debugPrefix, len(evalReport.ReviewedFiles), reviewedDir)
		}
	}

	// Build re-run command so users can reproduce this evaluation
	evalReport.RerunCommand = buildRerunCommand(task.Prompt.ID, task.Config.Name, e.opts)

	// Write JSON report
	reportPath, err := report.WriteReport(evalReport, e.opts.OutputDir, runID, task.Prompt)
	if err != nil {
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: ERROR: failed to write report: %v", debugPrefix, err)
		}
	} else if e.opts.Debug {
		log.Printf("[DEBUG] %s: report written to %s", debugPrefix, reportPath)
	}

	// Write HTML report
	if _, err := report.WriteHTMLReport(evalReport, e.opts.OutputDir, runID,
		task.Prompt.Service, task.Prompt.Plane, task.Prompt.Language, task.Prompt.Category); err != nil {
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: ERROR: failed to write HTML report: %v", debugPrefix, err)
		}
	}

	// Write Markdown report
	if _, err := report.WriteMarkdownReport(evalReport, e.opts.OutputDir, runID,
		task.Prompt.Service, task.Prompt.Plane, task.Prompt.Language, task.Prompt.Category); err != nil {
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: ERROR: failed to write Markdown report: %v", debugPrefix, err)
		}
	}

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
	if opts.VerifyBuild {
		parts = append(parts, "--verify-build")
	}

	// Include non-default timeouts.
	// Default generate timeout is 10m (600s), verify and review are 5m (300s).
	if opts.GenerateTimeout != 10*time.Minute {
		parts = append(parts, fmt.Sprintf("--generate-timeout=%d", int(opts.GenerateTimeout.Seconds())))
	}
	if opts.VerifyTimeout != 5*time.Minute {
		parts = append(parts, fmt.Sprintf("--verify-timeout=%d", int(opts.VerifyTimeout.Seconds())))
	}
	if opts.ReviewTimeout != 5*time.Minute {
		parts = append(parts, fmt.Sprintf("--review-timeout=%d", int(opts.ReviewTimeout.Seconds())))
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
