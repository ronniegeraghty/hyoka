// Package review provides code review functionality using Copilot.
package review

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/utils"
)

// Reviewer runs LLM-as-judge code reviews via a separate Copilot session.
type Reviewer interface {
	Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error)
}

// CopilotReviewer uses a Copilot session to perform code reviews.
type CopilotReviewer struct {
	client           *copilot.Client
	model            string
	skillDirectories []string
}

// NewCopilotReviewer creates a reviewer backed by the given Copilot client.
func NewCopilotReviewer(client *copilot.Client, model string) *CopilotReviewer {
	if model == "" {
		model = "claude-sonnet-4.5"
	}
	return &CopilotReviewer{client: client, model: model}
}

// SetSkillDirectories configures skill directories for the review session.
func (r *CopilotReviewer) SetSkillDirectories(dirs []string) {
	r.skillDirectories = dirs
}

// Review creates a separate Copilot session, sends the review prompt, and parses results.
func (r *CopilotReviewer) Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error) {
	generatedFiles, err := utils.ReadDirFiles(workDir)
	if err != nil {
		return nil, fmt.Errorf("reading generated files: %w", err)
	}
	if len(generatedFiles) == 0 {
		return nil, fmt.Errorf("no generated files found in %s", workDir)
	}

	var referenceFiles map[string]string
	if referenceDir != "" {
		referenceFiles, err = utils.ReadDirFiles(referenceDir)
		if err != nil {
			// Non-fatal: proceed without reference
			referenceFiles = nil
		}
	}

	reviewPrompt := BuildReviewPrompt(originalPrompt, generatedFiles, referenceFiles, evaluationCriteria)

	session, err := r.client.CreateSession(ctx, &copilot.SessionConfig{
		Model: r.model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code review judge evaluating another AI agent's work. The agent was given a prompt and asked to produce code. Score the generated code using the rubric and any prompt-specific evaluation criteria provided. Respond with ONLY valid JSON. No markdown, no explanation.",
		},
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    r.skillDirectories,
	})
	if err != nil {
		return nil, fmt.Errorf("creating review session: %w", err)
	}
	// SDK's Disconnect() can block if the CLI subprocess is stuck.
	// Timeout and let the owning client's Stop handle final cleanup.
	defer func() {
		done := make(chan struct{})
		go func() { session.Disconnect(); close(done) }()
		select {
		case <-done:
		case <-time.After(15 * time.Second):
		}
	}()

	// Capture the assistant's response and all session events
	var assistantContent strings.Builder
	var reviewEvents []ReviewEvent
	var mu sync.Mutex
	unsub := session.On(func(event copilot.SessionEvent) {
		mu.Lock()
		defer mu.Unlock()

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
	})
	defer unsub()

	_, err = session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: reviewPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("review session send: %w", err)
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

// StubReviewer returns placeholder review results for testing.
type StubReviewer struct{}

// Review returns a stub review result.
func (s *StubReviewer) Review(_ context.Context, _ string, _ string, _ string, _ string) (*ReviewResult, error) {
	return &ReviewResult{
		Scores: ReviewScores{
			Correctness:   0,
			Completeness:  0,
			BestPractices: 0,
			ErrorHandling: 0,
			PackageUsage:  0,
			CodeQuality:   0,
		},
		OverallScore: 0,
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
	return &result, nil
}

// PanelReviewer runs multiple reviewers in parallel and consolidates results.
type PanelReviewer struct {
	clientOpts       *copilot.ClientOptions
	models           []string // first model is the consolidator
	skillDirectories []string
	debug            bool
}

// NewPanelReviewer creates a panel reviewer that runs multiple models concurrently.
// The first model in the list is used as the consolidator.
func NewPanelReviewer(clientOpts *copilot.ClientOptions, models []string, debug bool) *PanelReviewer {
	return &PanelReviewer{
		clientOpts: clientOpts,
		models:     models,
		debug:      debug,
	}
}

// SetSkillDirectories configures skill directories for all review sessions.
func (p *PanelReviewer) SetSkillDirectories(dirs []string) {
	p.skillDirectories = dirs
}

// ReviewPanel runs all reviewer models in parallel and returns individual results
// plus a consolidated result. The consolidated result is produced by the first
// eligible model in the list, which receives all other reviewers' outputs.
// generatorModel is excluded from the reviewer panel to avoid self-review.
func (p *PanelReviewer) ReviewPanel(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string, generatorModel string) (panel []ReviewResult, consolidated *ReviewResult, err error) {
	// Filter out the generator model from reviewers
	var activeModels []string
	for _, m := range p.models {
		if m != generatorModel {
			activeModels = append(activeModels, m)
		}
	}
	if len(activeModels) == 0 {
		return nil, nil, fmt.Errorf("no reviewer models available after excluding generator model %q", generatorModel)
	}

	generatedFiles, err := utils.ReadDirFiles(workDir)
	if err != nil || len(generatedFiles) == 0 {
		return nil, nil, fmt.Errorf("no generated files to review in %s", workDir)
	}

	var referenceFiles map[string]string
	if referenceDir != "" {
		referenceFiles, _ = utils.ReadDirFiles(referenceDir)
	}

	reviewPrompt := BuildReviewPrompt(originalPrompt, generatedFiles, referenceFiles, evaluationCriteria)

	// Run all reviewers in parallel
	type reviewOutput struct {
		index  int
		model  string
		result *ReviewResult
		err    error
	}

	results := make(chan reviewOutput, len(activeModels))
	var wg sync.WaitGroup

	for i, model := range activeModels {
		wg.Add(1)
		go func(idx int, m string) {
			defer wg.Done()
			result, reviewErr := p.runSingleReview(ctx, m, reviewPrompt, workDir)
			if result != nil {
				result.Model = m
			}
			results <- reviewOutput{index: idx, model: m, result: result, err: reviewErr}
		}(i, model)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	ordered := make([]*ReviewResult, len(activeModels))
	for out := range results {
		if out.err != nil && p.debug {
			fmt.Printf("[DEBUG] reviewer %s failed: %v\n", out.model, out.err)
		}
		ordered[out.index] = out.result
	}

	// Build panel (non-nil results only)
	for _, r := range ordered {
		if r != nil {
			panel = append(panel, *r)
		}
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
	consolidated, err = p.consolidate(ctx, originalPrompt, generatedFiles, panel)
	if err != nil {
		// Fallback: use median scores from panel
		consolidated = medianReview(panel)
	}
	consolidated.Model = "consensus"

	return panel, consolidated, nil
}

// Review implements the Reviewer interface using the panel (for backward compat).
func (p *PanelReviewer) Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string, evaluationCriteria string) (*ReviewResult, error) {
	_, consolidated, err := p.ReviewPanel(ctx, originalPrompt, workDir, referenceDir, evaluationCriteria, "")
	return consolidated, err
}

// runSingleReview creates a Copilot client, runs a review session, and returns the result.
func (p *PanelReviewer) runSingleReview(ctx context.Context, model string, reviewPrompt string, workDir string) (*ReviewResult, error) {
	opts := *p.clientOpts
	client := copilot.NewClient(&opts)

	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting reviewer client for %s: %w", model, err)
	}
	defer func() {
		done := make(chan struct{})
		go func() { client.Stop(); close(done) }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
		}
	}()

	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		Model: model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code review judge evaluating another AI agent's work. Score the generated code using the rubric and any prompt-specific evaluation criteria provided. Respond with ONLY valid JSON. No markdown, no explanation.",
		},
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    p.skillDirectories,
	})
	if err != nil {
		return nil, fmt.Errorf("creating review session for %s: %w", model, err)
	}

	var assistantContent strings.Builder
	var mu sync.Mutex
	unsub := session.On(func(event copilot.SessionEvent) {
		mu.Lock()
		defer mu.Unlock()
		if event.Type == copilot.SessionEventTypeAssistantMessage && event.Data.Content != nil {
			assistantContent.WriteString(*event.Data.Content)
		}
	})
	defer unsub()

	_, err = session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: reviewPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("review session send for %s: %w", model, err)
	}

	mu.Lock()
	responseText := assistantContent.String()
	mu.Unlock()

	return parseReviewResponse(responseText)
}

// consolidate uses the first model to synthesize all individual reviews into a consensus.
func (p *PanelReviewer) consolidate(ctx context.Context, originalPrompt string, generatedFiles map[string]string, panel []ReviewResult) (*ReviewResult, error) {
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
	b.WriteString("Produce a consensus review. Use the median of individual scores for each dimension. ")
	b.WriteString("Combine the best issues and strengths from all reviewers. ")
	b.WriteString("Write a summary that captures the consensus view.\n\n")
	b.WriteString("Respond with ONLY a JSON object in the same format as the individual reviews.\n")

	consolidatorModel := p.models[0]
	result, err := p.runSingleReview(ctx, consolidatorModel, b.String(), "")
	if err != nil {
		return nil, fmt.Errorf("consolidation failed: %w", err)
	}
	return result, nil
}

// medianReview computes median scores across a panel as a fallback.
func medianReview(panel []ReviewResult) *ReviewResult {
	if len(panel) == 0 {
		return &ReviewResult{Summary: "No reviews to consolidate"}
	}

	median := func(extract func(ReviewScores) int) int {
		vals := make([]int, 0, len(panel))
		for _, r := range panel {
			vals = append(vals, extract(r.Scores))
		}
		// Simple sort for small N
		for i := range vals {
			for j := i + 1; j < len(vals); j++ {
				if vals[j] < vals[i] {
					vals[i], vals[j] = vals[j], vals[i]
				}
			}
		}
		return vals[len(vals)/2]
	}

	scores := ReviewScores{
		Correctness:   median(func(s ReviewScores) int { return s.Correctness }),
		Completeness:  median(func(s ReviewScores) int { return s.Completeness }),
		BestPractices: median(func(s ReviewScores) int { return s.BestPractices }),
		ErrorHandling: median(func(s ReviewScores) int { return s.ErrorHandling }),
		PackageUsage:  median(func(s ReviewScores) int { return s.PackageUsage }),
		CodeQuality:   median(func(s ReviewScores) int { return s.CodeQuality }),
	}

	// Median overall
	overalls := make([]int, 0, len(panel))
	for _, r := range panel {
		overalls = append(overalls, r.OverallScore)
	}
	for i := range overalls {
		for j := i + 1; j < len(overalls); j++ {
			if overalls[j] < overalls[i] {
				overalls[i], overalls[j] = overalls[j], overalls[i]
			}
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
		Model:        "consensus (median)",
		Scores:       scores,
		OverallScore: overalls[len(overalls)/2],
		Summary:      fmt.Sprintf("Median consensus from %d reviewers", len(panel)),
		Issues:       issues,
		Strengths:    strengths,
	}
}
