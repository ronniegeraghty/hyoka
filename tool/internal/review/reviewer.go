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
	Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string) (*ReviewResult, error)
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
func (r *CopilotReviewer) Review(ctx context.Context, originalPrompt string, workDir string, referenceDir string) (*ReviewResult, error) {
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

	reviewPrompt := BuildReviewPrompt(originalPrompt, generatedFiles, referenceFiles)

	session, err := r.client.CreateSession(ctx, &copilot.SessionConfig{
		Model: r.model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code review judge. Respond with ONLY valid JSON. No markdown, no explanation.",
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
func (s *StubReviewer) Review(_ context.Context, _ string, _ string, _ string) (*ReviewResult, error) {
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
