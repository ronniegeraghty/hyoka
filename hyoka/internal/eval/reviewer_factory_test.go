package eval

import (
	"context"
	"io"
	"testing"

	"github.com/ronniegeraghty/hyoka/internal/config"
	"github.com/ronniegeraghty/hyoka/internal/prompt"
	"github.com/ronniegeraghty/hyoka/internal/review"
)

// TestReviewerFactoryPerConfig verifies that each config gets its own reviewer
// instance with the correct model configuration (issue #92).
func TestReviewerFactoryPerConfig(t *testing.T) {
	// Track which configs were used to create reviewers
	configToModels := make(map[string][]string)

	factory := func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error) {
		var models []string
		if cfg.Reviewer != nil {
			models = cfg.Reviewer.Models
		}
		configToModels[cfg.Name] = models
		// Return a stub reviewer
		return &review.StubReviewer{}, nil, nil
	}

	// Create two configs with different reviewer models
	configs := []config.ToolConfig{
		{
			Name: "config-opus",
			Generator: &config.GeneratorConfig{
				Model: "claude-opus-4.6",
			},
			Reviewer: &config.ReviewerConfig{
				Models: []string{"claude-opus-4.6"},
			},
		},
		{
			Name: "config-sonnet",
			Generator: &config.GeneratorConfig{
				Model: "claude-sonnet-4.5",
			},
			Reviewer: &config.ReviewerConfig{
				Models: []string{"claude-sonnet-4.5"},
			},
		},
	}

	// Create a single prompt
	prompts := []*prompt.Prompt{
		{
			ID:                  "test-prompt",
			Language:            "python",
			Service:             "identity",
			Plane:               "data-plane",
			Category:            "auth",
			PromptText:          "Create a test file",
			EvaluationCriteria:  "Must work",
		},
	}

	// Create engine with the factory
	engine := NewEngineWithReviewerFactory(&StubEvaluator{}, factory, quietOpts(EngineOptions{
		OutputDir:    t.TempDir(),
		SkipTests:    true,
		SkipReview:   false, // Enable review to trigger factory calls
		DryRun:       false,
		Workers:      1,
		ProgressMode: "off",
	}))

	// Run evaluation
	_, err := engine.Run(context.Background(), prompts, configs)
	if err != nil {
		t.Fatalf("engine.Run failed: %v", err)
	}

	// Verify factory was called for each config
	if len(configToModels) != 2 {
		t.Errorf("expected factory to be called for 2 configs, got %d", len(configToModels))
	}

	// Verify correct models were associated with each config
	opusModels, hasOpus := configToModels["config-opus"]
	if !hasOpus {
		t.Error("expected factory to be called for config-opus")
	} else if len(opusModels) != 1 || opusModels[0] != "claude-opus-4.6" {
		t.Errorf("expected config-opus to have opus model, got %v", opusModels)
	}

	sonnetModels, hasSonnet := configToModels["config-sonnet"]
	if !hasSonnet {
		t.Error("expected factory to be called for config-sonnet")
	} else if len(sonnetModels) != 1 || sonnetModels[0] != "claude-sonnet-4.5" {
		t.Errorf("expected config-sonnet to have sonnet model, got %v", sonnetModels)
	}
}

// TestReviewerFactoryBackwardCompat verifies that the deprecated
// NewEngineWithReviewer still works (backward compatibility).
func TestReviewerFactoryBackwardCompat(t *testing.T) {
	// Create a stub reviewer
	reviewer := &review.StubReviewer{}

	// Create engine using deprecated method
	engine := NewEngineWithReviewer(&StubEvaluator{}, reviewer, quietOpts(EngineOptions{
		OutputDir:    t.TempDir(),
		SkipTests:    true,
		SkipReview:   false,
		DryRun:       false,
		Workers:      1,
		ProgressMode: "off",
		Stdout:       io.Discard,
	}))

	// Verify engine was created
	if engine == nil {
		t.Fatal("expected engine to be created")
	}

	// Verify factory was set (even though we used the old constructor)
	if engine.reviewerFactory == nil {
		t.Error("expected reviewerFactory to be set for backward compat")
	}
}

// TestReviewerFactoryNilWhenSkipReview verifies that the factory returns
// nil reviewers when SkipReview is true.
func TestReviewerFactoryNilWhenSkipReview(t *testing.T) {
	factoryCalled := false
	factory := func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error) {
		factoryCalled = true
		return &review.StubReviewer{}, nil, nil
	}

	engine := NewEngineWithReviewerFactory(&StubEvaluator{}, factory, quietOpts(EngineOptions{
		OutputDir:    t.TempDir(),
		SkipTests:    true,
		SkipReview:   true, // Skip review
		DryRun:       false,
		Workers:      1,
		ProgressMode: "off",
	}))

	prompts := []*prompt.Prompt{{ID: "test", Language: "python", PromptText: "test"}}
	configs := []config.ToolConfig{{Name: "test-config", Generator: &config.GeneratorConfig{Model: "gpt-4"}}}

	_, err := engine.Run(context.Background(), prompts, configs)
	if err != nil {
		t.Fatalf("engine.Run failed: %v", err)
	}

	// Factory should not be called when SkipReview is true
	if factoryCalled {
		t.Error("expected factory NOT to be called when SkipReview=true")
	}
}
