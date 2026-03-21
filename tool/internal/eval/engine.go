package eval

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/build"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/progress"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/prompt"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/review"
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
	Workers     int
	Timeout     time.Duration
	OutputDir   string
	SkipTests   bool
	SkipReview  bool
	VerifyBuild bool
	Debug       bool
	DryRun      bool
}

// Verifier evaluates generated code against prompt requirements.
type Verifier interface {
	Verify(ctx context.Context, originalPrompt string, workDir string, expectedCoverage string) (*report.VerifyResult, error)
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
	evaluator CopilotEvaluator
	reviewer  review.Reviewer
	verifier  Verifier
	opts      EngineOptions
}

// NewEngine creates a new Engine with the given evaluator and options.
func NewEngine(evaluator CopilotEvaluator, opts EngineOptions) *Engine {
	return NewEngineWithReviewer(evaluator, nil, nil, opts)
}

// NewEngineWithReviewer creates a new Engine with an evaluator, verifier, and reviewer.
func NewEngineWithReviewer(evaluator CopilotEvaluator, verifier Verifier, reviewer review.Reviewer, opts EngineOptions) *Engine {
	if opts.Workers <= 0 {
		opts.Workers = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Minute
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./reports"
	}
	return &Engine{
		evaluator: evaluator,
		reviewer:  reviewer,
		verifier:  verifier,
		opts:      opts,
	}
}

// EvalTask represents a single prompt+config evaluation to run.
type EvalTask struct {
	Prompt *prompt.Prompt
	Config config.ToolConfig
}

// Run executes evaluations for the given prompts crossed with configs.
func (e *Engine) Run(ctx context.Context, prompts []*prompt.Prompt, configs []config.ToolConfig) (*report.RunSummary, error) {
	// Build task list (cross product)
	var tasks []EvalTask
	for _, p := range prompts {
		for _, c := range configs {
			tasks = append(tasks, EvalTask{Prompt: p, Config: c})
		}
	}

	if e.opts.DryRun {
		return e.dryRun(tasks)
	}

	runID := time.Now().Format("20060102-150405")
	summary := &report.RunSummary{
		RunID:        runID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalPrompts: len(prompts),
		TotalConfigs: len(configs),
		TotalEvals:   len(tasks),
	}

	start := time.Now()

	// Progress display (disabled in debug mode)
	bar := progress.New(len(tasks), e.opts.Workers, e.opts.Debug)

	sem := make(chan struct{}, e.opts.Workers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t EvalTask) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			taskName := t.Prompt.ID + "/" + t.Config.Name
			bar.Start(taskName)
			taskStart := time.Now()

			evalReport := e.runSingleEval(ctx, t, runID)

			bar.Complete(taskName, evalReport.Success, time.Since(taskStart), evalReport.Error != "")

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
	bar.Done()

	summary.Duration = time.Since(start).Seconds()

	// Write JSON summary
	if _, err := report.WriteSummary(summary, e.opts.OutputDir); err != nil && e.opts.Debug {
		log.Printf("failed to write run summary: %v", err)
	}

	// Write HTML summary
	if _, err := report.WriteSummaryHTML(summary, e.opts.OutputDir); err != nil && e.opts.Debug {
		log.Printf("failed to write HTML summary: %v", err)
	}

	// Write Markdown summary
	if _, err := report.WriteSummaryMarkdown(summary, e.opts.OutputDir); err != nil && e.opts.Debug {
		log.Printf("failed to write Markdown summary: %v", err)
	}

	return summary, nil
}

func (e *Engine) runSingleEval(ctx context.Context, task EvalTask, runID string) *report.EvalReport {
	evalCtx, cancel := context.WithTimeout(ctx, e.opts.Timeout)
	defer cancel()

	debugPrefix := task.Prompt.ID + "/" + task.Config.Name
	start := time.Now()

	evalReport := &report.EvalReport{
		PromptID:   task.Prompt.ID,
		ConfigName: task.Config.Name,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		PromptMeta: map[string]any{
			"service":  task.Prompt.Service,
			"plane":    task.Prompt.Plane,
			"language": task.Prompt.Language,
			"category": task.Prompt.Category,
		},
		ConfigUsed: map[string]any{
			"name":  task.Config.Name,
			"model": task.Config.Model,
		},
	}

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Starting Copilot session...", debugPrefix)
	}

	// Setup workspace in OS temp dir (ephemeral)
	ws, err := NewWorkspace(task.Prompt.ID, task.Config.Name)
	if err != nil {
		evalReport.Error = fmt.Sprintf("workspace setup failed: %v", err)
		evalReport.ErrorDetails = err.Error()
		evalReport.Duration = time.Since(start).Seconds()
		return evalReport
	}
	defer ws.Cleanup()

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Workspace: %s", debugPrefix, ws.Dir)
	}

	// Run evaluation
	result, err := e.evaluator.Evaluate(evalCtx, task.Prompt, &task.Config, ws.Dir)
	if err != nil {
		evalReport.Error = fmt.Sprintf("evaluation failed: %v", err)
		evalReport.ErrorDetails = err.Error()
		evalReport.Duration = time.Since(start).Seconds()
		// Capture whatever session events were collected before failure
		if result != nil {
			evalReport.SessionEvents = result.SessionEvents
			evalReport.EventCount = result.EventCount
			evalReport.ToolCalls = result.ToolCalls
			evalReport.IsStub = result.IsStub
		}
		return evalReport
	}

	evalReport.EventCount = result.EventCount
	evalReport.ToolCalls = result.ToolCalls
	evalReport.SessionEvents = result.SessionEvents
	evalReport.IsStub = result.IsStub
	evalReport.Success = result.Success

	// Collect generated files from workspace
	generatedFiles, _ := ws.ListFiles()
	evalReport.GeneratedFiles = generatedFiles

	if e.opts.Debug {
		log.Printf("[DEBUG] %s: Session complete: %d tool calls, %d files generated, %s",
			debugPrefix, len(result.ToolCalls), len(generatedFiles), time.Since(start).Truncate(time.Millisecond))
	}

	// Copilot-based verification
	if e.verifier != nil {
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: Starting verification session...", debugPrefix)
		}
		verifyResult, err := e.verifier.Verify(evalCtx, task.Prompt.PromptText, ws.Dir, task.Prompt.ExpectedCoverage)
		if err != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: ERROR: verification failed: %v", debugPrefix, err)
			}
		} else {
			evalReport.Verification = verifyResult
			evalReport.Success = verifyResult.Pass
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
	if e.opts.VerifyBuild {
		buildResult, err := build.Verify(evalCtx, task.Prompt.Language, ws.Dir)
		if err != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: ERROR: build verification failed: %v", debugPrefix, err)
			}
		} else {
			evalReport.Build = buildResult
			if !buildResult.Success {
				evalReport.Success = false
			}
		}
	}

	// Code review (unless skipped)
	if !e.opts.SkipReview && e.reviewer != nil {
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: Starting review session...", debugPrefix)
		}
		referenceDir := ""
		if task.Prompt.ReferenceAnswer != "" {
			referenceDir = task.Prompt.ReferenceAnswer
		}
		reviewResult, err := e.reviewer.Review(evalCtx, task.Prompt.PromptText, ws.Dir, referenceDir)
		if err != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: ERROR: code review failed: %v", debugPrefix, err)
			}
		} else {
			evalReport.Review = reviewResult
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: Review score: %d/10", debugPrefix, reviewResult.OverallScore)
			}
		}

		// Capture reviewed (annotated) files — the review session may have added REVIEW: comments
		reviewedFiles, err := readReviewedFiles(ws.Dir)
		if err == nil && len(reviewedFiles) > 0 {
			evalReport.ReviewedFiles = reviewedFiles
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: Captured %d reviewed files with annotations", debugPrefix, len(reviewedFiles))
			}
		}
	}

	// Tool usage evaluation (compare expected vs actual tools)
	if len(task.Prompt.ExpectedTools) > 0 {
		evalReport.ToolUsage = evaluateToolUsage(task.Prompt.ExpectedTools, result.ToolCalls)
		if e.opts.Debug {
			log.Printf("[DEBUG] %s: Tool usage: match=%v, matched=%v, missing=%v",
				debugPrefix, evalReport.ToolUsage.Match, evalReport.ToolUsage.MatchedTools, evalReport.ToolUsage.MissingTools)
		}
	}

	evalReport.Duration = time.Since(start).Seconds()

	// Build the report directory path (includes config name)
	reportDir := filepath.Join(report.ReportDir(e.opts.OutputDir, runID, task.Prompt), task.Config.Name)

	// Copy generated files from temp workspace into report under generated-code/
	if len(generatedFiles) > 0 {
		codeDir := filepath.Join(reportDir, "generated-code")
		if _, err := ws.CopyFilesTo(codeDir); err != nil {
			if e.opts.Debug {
				log.Printf("[DEBUG] %s: ERROR: failed to copy generated files: %v", debugPrefix, err)
			}
		} else if e.opts.Debug {
			log.Printf("[DEBUG] %s: Copied %d generated files to %s", debugPrefix, len(generatedFiles), codeDir)
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
