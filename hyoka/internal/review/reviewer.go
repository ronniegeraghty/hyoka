// Package review provides code review functionality using Copilot.
package review

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/hyoka/internal/logging"
	"github.com/ronniegeraghty/hyoka/internal/utils"
)

// Reviewer runs LLM-as-judge code reviews via a separate Copilot session.
type Reviewer interface {
	Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error)
}

// CopilotReviewer uses a Copilot session to perform code reviews.
type CopilotReviewer struct {
	client            *copilot.Client
	model             string
	maxSessionActions int
	skillDirectories  []string
	ignoreDirs        map[string]bool
	promptID          string
	configName        string
}

// NewCopilotReviewer creates a reviewer backed by the given Copilot client.
func NewCopilotReviewer(client *copilot.Client, model string, maxSessionActions int, promptID, configName string) *CopilotReviewer {
	if model == "" {
		model = "claude-sonnet-4.5"
	}
	return &CopilotReviewer{client: client, model: model, maxSessionActions: maxSessionActions, promptID: promptID, configName: configName}
}

// SetSkillDirectories configures skill directories for the review session.
func (r *CopilotReviewer) SetSkillDirectories(dirs []string) {
	r.skillDirectories = dirs
}

// SetPromptContext sets the prompt ID and config name for structured logging.
func (r *CopilotReviewer) SetPromptContext(promptID, configName string) {
	r.promptID = promptID
	r.configName = configName
}

// SetIgnoreDirs configures directories to exclude when reading files for the
// review prompt. Files remain on disk in the workspace for build skills to use.
func (r *CopilotReviewer) SetIgnoreDirs(dirs map[string]bool) {
	r.ignoreDirs = dirs
}

// Review creates a separate Copilot session, sends the review prompt, and parses results.
func (r *CopilotReviewer) Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error) {
	lg := logging.ReviewLogger(r.promptID, r.configName, r.model)
	lg.Debug("Reading generated files", "workDir", workDir)
	generatedFiles, err := utils.ReadDirFilesFiltered(workDir, r.ignoreDirs)
	if err != nil {
		return nil, fmt.Errorf("reading generated files: %w", err)
	}
	if len(generatedFiles) == 0 {
		return nil, fmt.Errorf("no generated files found in %s", workDir)
	}
	lg.Debug("Generated files loaded", "file_count", len(generatedFiles))

	var referenceFiles map[string]string
	if referenceDir != "" {
		referenceFiles, err = utils.ReadDirFilesFiltered(referenceDir, r.ignoreDirs)
		if err != nil {
			// Non-fatal: proceed without reference
			lg.Warn("Could not read reference files, proceeding without", "referenceDir", referenceDir, "error", err)
			referenceFiles = nil
		}
	}

	reviewPrompt := BuildReviewPrompt(originalPrompt, generatedFiles, referenceFiles, evaluationCriteria)

	// Create isolated config directory to prevent user-level skills from
	// leaking into the review session (#21).
	configDir, err := os.MkdirTemp("", "hyoka-config-*")
	if err != nil {
		return nil, fmt.Errorf("creating isolated config dir: %w", err)
	}
	defer os.RemoveAll(configDir)

	// Capture the assistant's response and all session events
	var assistantContent strings.Builder
	var reviewEvents []ReviewEvent
	var mu sync.Mutex

	var actionCounter int
	var actionLimitHit bool
	reviewCtx, reviewCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer reviewCancel()

	eventHandler := func(event copilot.SessionEvent) {
		mu.Lock()
		defer mu.Unlock()

		// Count actions and enforce limit
		switch event.Type {
		case copilot.SessionEventTypeAssistantReasoning,
			copilot.SessionEventTypeAssistantMessage,
			copilot.SessionEventTypeToolExecutionStart:
			actionCounter++
			if r.maxSessionActions > 0 && actionCounter > r.maxSessionActions && !actionLimitHit {
				actionLimitHit = true
				lg.Warn("Action limit reached, cancelling session",
					"actions", actionCounter, "max_session_actions", r.maxSessionActions)
				reviewCancel()
			}
		}

		// Log review events at debug level for visibility during runs.
		switch event.Type {
		case copilot.SessionEventTypeAssistantTurnStart:
			lg.Debug("Turn started")
		case copilot.SessionEventTypeAssistantTurnEnd:
			lg.Debug("Turn ended")
		case copilot.SessionEventTypeAssistantMessage:
			if event.Data.Content != nil {
				lg.Debug("Assistant message",
					"content_len", len(*event.Data.Content))
			}
		case copilot.SessionEventTypeToolExecutionStart:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			lg.Debug("Tool start", "tool", toolName)
		case copilot.SessionEventTypeToolExecutionComplete:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			lg.Debug("Tool done", "tool", toolName)
		case copilot.SessionEventTypeAssistantUsage:
			lg.Debug("Token usage")
		}

		if event.Type == copilot.SessionEventTypeAssistantMessage && event.Data.Content != nil {
			assistantContent.WriteString(*event.Data.Content)
		}

		// Capture all events for the report timeline
		evt := ReviewEvent{Type: string(event.Type)}
		if event.Data.ToolName != nil {
			evt.ToolName = *event.Data.ToolName
		}
		if event.Data.Content != nil {
			evt.Content = *event.Data.Content
		}
		if event.Data.Arguments != nil {
			if argsBytes, err := json.Marshal(event.Data.Arguments); err == nil {
				evt.ToolArgs = string(argsBytes)
			}
		}
		if event.Data.Result != nil {
			if event.Data.Result.Content != nil {
				evt.Result = *event.Data.Result.Content
			}
		}
		if event.Data.Error != nil {
			if event.Data.Error.ErrorClass != nil {
				evt.Error = event.Data.Error.ErrorClass.Message
			} else if event.Data.Error.String != nil {
				evt.Error = *event.Data.Error.String
			}
		}
		if event.Data.Duration != nil {
			evt.Duration = *event.Data.Duration
		}
		reviewEvents = append(reviewEvents, evt)
	}

	lg.Info("Starting review session", "skills", len(r.skillDirectories), "work_dir", workDir)
	lg.Debug("Creating review session")
	session, err := r.client.CreateSession(reviewCtx, &copilot.SessionConfig{
		Model: r.model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code review judge evaluating another AI agent's work. Actively verify the code: attempt to build it, check if SDK packages are the latest versions, and test any claims. Score each criterion as pass/fail per the rubric. Respond with ONLY valid JSON. No markdown, no explanation.",
		},
		ConfigDir:           configDir,
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    r.skillDirectories,
		OnEvent:             eventHandler,
	})
	if err != nil {
		lg.Error("Failed to create review session", "error", err)
		return nil, fmt.Errorf("creating review session: %w", err)
	}
	// Clean up session state (#62). DeleteSession removes session-state dir
	// and SQLite entry while client is still connected. Then Disconnect
	// releases in-memory resources.
	defer func() {
		deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer deleteCancel()
		if err := r.client.DeleteSession(deleteCtx, session.SessionID); err != nil {
			lg.Debug("Session delete failed", "sessionID", session.SessionID, "error", err)
		}
		done := make(chan struct{})
		go func() { session.Disconnect(); close(done) }()
		select {
		case <-done:
		case <-time.After(15 * time.Second):
		}
	}()

	lg.Debug("Sending review prompt")
	_, err = session.SendAndWait(reviewCtx, copilot.MessageOptions{
		Prompt: reviewPrompt,
	})
	if err != nil {
		lg.Error("Review session send failed", "error", err)
		return nil, fmt.Errorf("review session send: %w", err)
	}

	mu.Lock()
	responseText := assistantContent.String()
	capturedEvents := make([]ReviewEvent, len(reviewEvents))
	copy(capturedEvents, reviewEvents)
	mu.Unlock()

	result, err := parseReviewResponse(responseText)
	if err != nil {
		lg.Error("Failed to parse review response", "error", err)
		return nil, err
	}
	result.Events = capturedEvents
	lg.Info("Review complete", "overall_score", result.OverallScore, "max_score", result.MaxScore)
	return result, nil
}

// StubReviewer returns placeholder review results for testing.
type StubReviewer struct{}

// Review returns a stub review result.
func (s *StubReviewer) Review(_ context.Context, _ string, _ string, _ string, _ string) (*ReviewResult, error) {
	return &ReviewResult{
		Scores: ReviewScores{
			Criteria: []CriterionResult{
				{Name: "stub_criterion", Passed: true, Reason: "stub mode"},
			},
		},
		OverallScore: 1,
		MaxScore:     1,
		Summary:      "Review skipped (stub evaluator)",
		Issues:       []string{},
		Strengths:    []string{},
	}, nil
}

// parseReviewResponse extracts the JSON ReviewResult from the LLM response.
func parseReviewResponse(text string) (*ReviewResult, error) {
	// Try to find JSON in the response (LLM may wrap it in markdown fences)
	jsonStr := utils.ExtractJSON(text)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in review response: %.200s", text)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parsing review JSON: %w (response: %.200s)", err, jsonStr)
	}
	// Ensure MaxScore and OverallScore are consistent with criteria
	if result.MaxScore == 0 && len(result.Scores.Criteria) > 0 {
		result.MaxScore = result.Scores.TotalCount()
	}
	if result.OverallScore == 0 && len(result.Scores.Criteria) > 0 {
		result.OverallScore = result.Scores.PassedCount()
	}
	return &result, nil
}

// PanelReviewer runs multiple reviewers in parallel and consolidates results.
type PanelReviewer struct {
	clientOpts        *copilot.ClientOptions
	models            []string // first model is the consolidator
	maxSessionActions int
	skillDirectories  []string
	ignoreDirs        map[string]bool
	promptID          string
	configName        string
}

// NewPanelReviewer creates a panel reviewer that runs multiple models concurrently.
// The first model in the list is used as the consolidator.
func NewPanelReviewer(clientOpts *copilot.ClientOptions, models []string, maxSessionActions int, promptID, configName string) *PanelReviewer {
	return &PanelReviewer{
		clientOpts:        clientOpts,
		models:            models,
		maxSessionActions: maxSessionActions,
		promptID:          promptID,
		configName:        configName,
	}
}

// SetSkillDirectories configures skill directories for all review sessions.
func (p *PanelReviewer) SetSkillDirectories(dirs []string) {
	p.skillDirectories = dirs
}

// SetPromptContext sets the prompt ID and config name for structured logging.
func (p *PanelReviewer) SetPromptContext(promptID, configName string) {
	p.promptID = promptID
	p.configName = configName
}

// SetIgnoreDirs configures directories to exclude when reading files for the
// review prompt. Files remain on disk in the workspace for build skills to use.
func (p *PanelReviewer) SetIgnoreDirs(dirs map[string]bool) {
	p.ignoreDirs = dirs
}

// Models returns the list of reviewer models.
func (p *PanelReviewer) Models() []string {
	return p.models
}

// ReviewPanel runs all reviewer models sequentially and returns individual results
// plus a consolidated result. The consolidated result is produced by the first model
// in the list, which receives all other reviewers' outputs.
// Reviews run one at a time so each Copilot session starts, completes, and stops
// before the next begins, reducing peak memory usage.
func (p *PanelReviewer) ReviewPanel(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (panel []ReviewResult, consolidated *ReviewResult, err error) {
	plg := slog.With("prompt", p.promptID, "config", p.configName, "role", "review-panel")
	plg.Info("Starting sequential panel review", "model_count", len(p.models), "models", p.models)
	if len(p.models) == 0 {
		return nil, nil, fmt.Errorf("no reviewer models configured")
	}

	generatedFiles, err := utils.ReadDirFilesFiltered(workDir, p.ignoreDirs)
	if err != nil || len(generatedFiles) == 0 {
		return nil, nil, fmt.Errorf("no generated files to review in %s", workDir)
	}
	plg.Debug("Generated files loaded", "file_count", len(generatedFiles))

	var referenceFiles map[string]string
	if referenceDir != "" {
		referenceFiles, _ = utils.ReadDirFilesFiltered(referenceDir, p.ignoreDirs)
	}

	reviewPrompt := BuildReviewPrompt(originalPrompt, generatedFiles, referenceFiles, evaluationCriteria)

	// Run reviewers sequentially — one Copilot session at a time
	for i, model := range p.models {
		rlg := logging.ReviewLogger(p.promptID, p.configName, model)
		rlg.Debug("Panel reviewer starting", "index", i)
		modelWorkDir, copyErr := copyDirToTemp(workDir, fmt.Sprintf("hyoka-review-%s-*", model))
		if copyErr != nil {
			rlg.Warn("Failed to create workspace copy for reviewer", "error", copyErr)
			modelWorkDir = workDir
		} else {
			defer os.RemoveAll(modelWorkDir)
		}
		result, reviewErr := p.runSingleReview(ctx, model, reviewPrompt, modelWorkDir)
		if result != nil {
			result.Model = model
		}
		if reviewErr != nil {
			rlg.Warn("Panel reviewer failed", "error", reviewErr)
			continue
		}
		rlg.Debug("Panel reviewer complete", "overall_score", result.OverallScore, "max_score", result.MaxScore)
		panel = append(panel, *result)
	}

	if len(panel) == 0 {
		return nil, nil, fmt.Errorf("all reviewers failed")
	}

	// If only one reviewer succeeded, use it as consolidated
	if len(panel) == 1 {
		c := panel[0]
		return panel, &c, nil
	}

	// Consolidate: use the first model to synthesize all reviews
	clg := logging.ConsolidatorLogger(p.promptID, p.configName, p.models[0])
	clg.Info("Starting review consolidation", "panel_size", len(panel))
	consolidated, err = p.consolidate(ctx, originalPrompt, generatedFiles, panel)
	if err != nil {
		// Fallback: use average scores from panel
		clg.Warn("Consolidation failed, falling back to average", "error", err)
		consolidated = averageReview(panel)
	}
	consolidated.Model = "consensus"
	plg.Info("Panel review complete", "panel_size", len(panel), "consensus_score", consolidated.OverallScore, "max_score", consolidated.MaxScore)

	return panel, consolidated, nil
}

// Review implements the Reviewer interface using the panel (for backward compat).
func (p *PanelReviewer) Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error) {
	_, consolidated, err := p.ReviewPanel(ctx, originalPrompt, workDir, referenceDir, evaluationCriteria)
	return consolidated, err
}

// runSingleReview creates a Copilot client, runs a review session, and returns the result.
func (p *PanelReviewer) runSingleReview(ctx context.Context, model string, reviewPrompt string, workDir string) (*ReviewResult, error) {
	lg := logging.ReviewLogger(p.promptID, p.configName, model)
	lg.Debug("Starting single review")
	opts := *p.clientOpts
	client := copilot.NewClient(&opts)

	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting reviewer client for %s: %w", model, err)
	}
	var panelSessionID string
	defer func() {
		// Delete session state before stopping client (#62)
		if panelSessionID != "" {
			deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer deleteCancel()
			if err := client.DeleteSession(deleteCtx, panelSessionID); err != nil {
				lg.Debug("Session delete failed",
					"sessionID", panelSessionID, "error", err)
			}
		}
		done := make(chan struct{})
		go func() { client.Stop(); close(done) }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
		}
	}()

	// Create isolated config directory to prevent user-level skills from
	// leaking into the review session (#21).
	configDir, err := os.MkdirTemp("", "hyoka-config-*")
	if err != nil {
		return nil, fmt.Errorf("creating isolated config dir for %s: %w", model, err)
	}
	defer os.RemoveAll(configDir)

	var assistantContent strings.Builder
	var reviewEvents []ReviewEvent
	var mu sync.Mutex

	var actionCounter int
	var actionLimitHit bool
	reviewCtx, reviewCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer reviewCancel()

	eventHandler := func(event copilot.SessionEvent) {
		mu.Lock()
		defer mu.Unlock()

		// Count actions and enforce limit
		switch event.Type {
		case copilot.SessionEventTypeAssistantReasoning,
			copilot.SessionEventTypeAssistantMessage,
			copilot.SessionEventTypeToolExecutionStart:
			actionCounter++
			if p.maxSessionActions > 0 && actionCounter > p.maxSessionActions && !actionLimitHit {
				actionLimitHit = true
				lg.Warn("Action limit reached, cancelling session",
					"actions", actionCounter, "max_session_actions", p.maxSessionActions)
				reviewCancel()
			}
		}

		// Log review events at debug level for visibility during runs.
		switch event.Type {
		case copilot.SessionEventTypeAssistantTurnStart:
			lg.Debug("Turn started")
		case copilot.SessionEventTypeAssistantTurnEnd:
			lg.Debug("Turn ended")
		case copilot.SessionEventTypeAssistantMessage:
			if event.Data.Content != nil {
				lg.Debug("Assistant message",
					"content_len", len(*event.Data.Content))
			}
		case copilot.SessionEventTypeToolExecutionStart:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			lg.Debug("Tool start", "tool", toolName)
		case copilot.SessionEventTypeToolExecutionComplete:
			toolName := ""
			if event.Data.ToolName != nil {
				toolName = *event.Data.ToolName
			}
			lg.Debug("Tool done", "tool", toolName)
		case copilot.SessionEventTypeAssistantUsage:
			lg.Debug("Token usage")
		}

		if event.Type == copilot.SessionEventTypeAssistantMessage && event.Data.Content != nil {
			assistantContent.WriteString(*event.Data.Content)
		}
		// Capture all events for the report timeline
		evt := ReviewEvent{Type: string(event.Type)}
		if event.Data.ToolName != nil {
			evt.ToolName = *event.Data.ToolName
		}
		if event.Data.Content != nil {
			evt.Content = *event.Data.Content
		}
		if event.Data.Arguments != nil {
			if argsBytes, err := json.Marshal(event.Data.Arguments); err == nil {
				evt.ToolArgs = string(argsBytes)
			}
		}
		if event.Data.Result != nil {
			if event.Data.Result.Content != nil {
				evt.Result = *event.Data.Result.Content
			}
		}
		if event.Data.Error != nil {
			if event.Data.Error.ErrorClass != nil {
				evt.Error = event.Data.Error.ErrorClass.Message
			} else if event.Data.Error.String != nil {
				evt.Error = *event.Data.Error.String
			}
		}
		if event.Data.Duration != nil {
			evt.Duration = *event.Data.Duration
		}
		reviewEvents = append(reviewEvents, evt)
	}

	lg.Info("Starting review session", "skills", len(p.skillDirectories), "work_dir", workDir)
	lg.Debug("Creating review session")
	session, err := client.CreateSession(reviewCtx, &copilot.SessionConfig{
		Model: model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code review judge evaluating another AI agent's work. Actively verify the code: attempt to build it, check if SDK packages are the latest versions, and test any claims. Score each criterion as pass/fail per the rubric. Respond with ONLY valid JSON. No markdown, no explanation.",
		},
		ConfigDir:           configDir,
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    p.skillDirectories,
		OnEvent:             eventHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("creating review session for %s: %w", model, err)
	}
	panelSessionID = session.SessionID

	lg.Debug("Sending review prompt")
	_, err = session.SendAndWait(reviewCtx, copilot.MessageOptions{
		Prompt: reviewPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("review session send for %s: %w", model, err)
	}

	mu.Lock()
	responseText := assistantContent.String()
	capturedEvents := make([]ReviewEvent, len(reviewEvents))
	copy(capturedEvents, reviewEvents)
	mu.Unlock()

	result, err := parseReviewResponse(responseText)
	if err != nil {
		return nil, err
	}
	result.Events = capturedEvents
	return result, nil
}

// consolidate uses the first model to synthesize all individual reviews into a consensus.
func (p *PanelReviewer) consolidate(ctx context.Context, originalPrompt string, generatedFiles map[string]string, panel []ReviewResult) (*ReviewResult, error) {
	consolidatorModel := p.models[0]
	lg := logging.ConsolidatorLogger(p.promptID, p.configName, consolidatorModel)
	lg.Debug("Starting consolidation", "panel_size", len(panel))
	var b strings.Builder
	b.WriteString("You are a senior review consolidator. Multiple independent reviewers have scored the same generated code.\n")
	b.WriteString("Synthesize their feedback into a single consensus review.\n\n")

	b.WriteString("## Original Prompt\n\n")
	b.WriteString(originalPrompt)
	b.WriteString("\n\n")

	b.WriteString("## Individual Reviews\n\n")
	for i, r := range panel {
		reviewJSON, _ := json.MarshalIndent(r, "", "  ")
		fmt.Fprintf(&b, "### Reviewer %d (%s)\n```json\n%s\n```\n\n", i+1, r.Model, string(reviewJSON))
	}

	b.WriteString("## Instructions\n\n")
	b.WriteString("Produce a consensus review using the criteria-based pass/fail system. ")
	b.WriteString("For each criterion, it PASSES if the majority of reviewers marked it as passed. ")
	b.WriteString("Use the union of all criteria across reviewers. ")
	b.WriteString("Combine the best issues and strengths from all reviewers. ")
	b.WriteString("Write a summary that captures the consensus view.\n\n")
	b.WriteString("Respond with ONLY a JSON object in the same format as the individual reviews.\n")

	lg.Debug("Sending consolidation prompt")
	result, err := p.runSingleReview(ctx, consolidatorModel, b.String(), "")
	if err != nil {
		return nil, fmt.Errorf("consolidation failed: %w", err)
	}
	lg.Debug("Consolidation complete", "overall_score", result.OverallScore, "max_score", result.MaxScore)
	return result, nil
}

// averageReview computes average pass rates across a panel as a fallback.
// For each criterion, it passes if the majority of reviewers marked it passed.
func averageReview(panel []ReviewResult) *ReviewResult {
	if len(panel) == 0 {
		return &ReviewResult{Summary: "No reviews to consolidate"}
	}

	// Collect all criteria by name, track pass counts
	type criterionAgg struct {
		passCount int
		total     int
		reasons   []string
	}
	criteriaMap := make(map[string]*criterionAgg)
	var criteriaOrder []string

	for _, r := range panel {
		for _, c := range r.Scores.Criteria {
			agg, exists := criteriaMap[c.Name]
			if !exists {
				agg = &criterionAgg{}
				criteriaMap[c.Name] = agg
				criteriaOrder = append(criteriaOrder, c.Name)
			}
			agg.total++
			if c.Passed {
				agg.passCount++
			}
			if c.Reason != "" {
				agg.reasons = append(agg.reasons, c.Reason)
			}
		}
	}

	// Build consensus criteria — passed if majority passed
	var criteria []CriterionResult
	passedCount := 0
	for _, name := range criteriaOrder {
		agg := criteriaMap[name]
		passed := agg.passCount > agg.total/2 // majority
		reason := fmt.Sprintf("%d/%d reviewers passed", agg.passCount, agg.total)
		criteria = append(criteria, CriterionResult{
			Name:   name,
			Passed: passed,
			Reason: reason,
		})
		if passed {
			passedCount++
		}
	}

	// Merge issues and strengths
	issueSet := make(map[string]bool)
	var issues []string
	strengthSet := make(map[string]bool)
	var strengths []string
	for _, r := range panel {
		for _, iss := range r.Issues {
			if !issueSet[iss] {
				issueSet[iss] = true
				issues = append(issues, iss)
			}
		}
		for _, s := range r.Strengths {
			if !strengthSet[s] {
				strengthSet[s] = true
				strengths = append(strengths, s)
			}
		}
	}

	return &ReviewResult{
		Model: "consensus (average)",
		Scores: ReviewScores{
			Criteria: criteria,
		},
		OverallScore: passedCount,
		MaxScore:     len(criteria),
		Summary:      fmt.Sprintf("Average consensus from %d reviewers: %d/%d criteria passed", len(panel), passedCount, len(criteria)),
		Issues:       issues,
		Strengths:    strengths,
	}
}

func copyDirToTemp(src string, pattern string) (string, error) {
	dst, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.Create(target)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
	if err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	return dst, nil
}
