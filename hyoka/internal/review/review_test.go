package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// BuildReviewPrompt tests
// ---------------------------------------------------------------------------

func TestBuildReviewPrompt(t *testing.T) {
	prompt := "Write Azure Blob Storage auth code"
	generated := map[string]string{
		"Program.cs": "using Azure.Storage.Blobs;\n// ...",
	}
	reference := map[string]string{
		"Program.cs": "using Azure.Storage.Blobs;\n// reference",
	}

	result := BuildReviewPrompt(prompt, generated, reference, "")

	if result == "" {
		t.Fatal("expected non-empty review prompt")
	}

	checks := []string{
		"Original Prompt",
		"Generated Code",
		"Reference Answer",
		"Scoring Rubric",
		"passed",
		"Program.cs",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("review prompt missing %q", check)
		}
	}
}

func TestBuildReviewPromptNoReference(t *testing.T) {
	prompt := "Write code"
	generated := map[string]string{"main.go": "package main"}

	result := BuildReviewPrompt(prompt, generated, nil, "")

	if !strings.Contains(result, "No reference answer provided") {
		t.Error("expected 'No reference answer provided' when no reference given")
	}
}

func TestBuildReviewPromptEmptyReference(t *testing.T) {
	result := BuildReviewPrompt("prompt", map[string]string{"a.go": "code"}, map[string]string{}, "")
	if !strings.Contains(result, "No reference answer provided") {
		t.Error("empty reference map should show 'No reference answer provided'")
	}
}

func TestBuildReviewPromptWithEvaluationCriteria(t *testing.T) {
	prompt := "Write Azure code"
	generated := map[string]string{"main.go": "package main"}
	criteria := "- Must use DefaultAzureCredential\n- Must handle errors"

	result := BuildReviewPrompt(prompt, generated, nil, criteria)

	if !strings.Contains(result, "Prompt-Specific Evaluation Criteria") {
		t.Error("expected evaluation criteria section")
	}
	if !strings.Contains(result, "DefaultAzureCredential") {
		t.Error("expected criteria content in prompt")
	}
}

func TestBuildReviewPromptNoCriteria(t *testing.T) {
	result := BuildReviewPrompt("prompt", map[string]string{"a.go": "code"}, nil, "")
	// The rubric may mention criteria, but the dynamic section header should
	// not appear before the rubric when no criteria are passed.
	rubricIdx := strings.Index(result, "# Review Scoring Rubric")
	if rubricIdx < 0 {
		rubricIdx = len(result)
	}
	beforeRubric := result[:rubricIdx]
	if strings.Contains(beforeRubric, "## Prompt-Specific Evaluation Criteria") {
		t.Error("should not contain criteria section header before rubric when criteria is empty")
	}
}

func TestBuildReviewPromptMultipleFiles(t *testing.T) {
	generated := map[string]string{
		"main.go":   "package main",
		"helper.go": "package helper",
		"util.go":   "package util",
	}
	reference := map[string]string{
		"ref_main.go": "package main // ref",
		"ref_help.go": "package helper // ref",
	}

	result := BuildReviewPrompt("prompt", generated, reference, "criteria")

	for name := range generated {
		if !strings.Contains(result, name) {
			t.Errorf("prompt missing generated file %q", name)
		}
	}
	for name := range reference {
		if !strings.Contains(result, name) {
			t.Errorf("prompt missing reference file %q", name)
		}
	}
}

func TestBuildReviewPromptEmptyGeneratedFiles(t *testing.T) {
	result := BuildReviewPrompt("prompt", map[string]string{}, nil, "")
	if !strings.Contains(result, "Generated Code") {
		t.Error("should still contain Generated Code header even with empty files")
	}
}

func TestBuildReviewPromptContainsRubric(t *testing.T) {
	result := BuildReviewPrompt("p", map[string]string{"f": "c"}, nil, "")
	if !strings.Contains(result, "Scoring Rubric") {
		t.Error("prompt should contain the embedded rubric")
	}
}

func TestBuildReviewPromptPreservesOriginalPrompt(t *testing.T) {
	original := "Write a Python script that uses azure-identity DefaultAzureCredential"
	result := BuildReviewPrompt(original, map[string]string{"main.py": "pass"}, nil, "")
	if !strings.Contains(result, original) {
		t.Error("prompt should contain the original prompt verbatim")
	}
}

// ---------------------------------------------------------------------------
// parseReviewResponse tests
// ---------------------------------------------------------------------------

func TestParseReviewResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		score     int
		maxScore  int
		criteria  int
		summary   string
		issues    int
		strengths int
	}{
		{
			name:      "clean json with criteria",
			input:     `{"scores":{"criteria":[{"name":"Code Builds","passed":true,"reason":"OK"},{"name":"Best Practices","passed":true,"reason":"Good"},{"name":"Error Handling","passed":false,"reason":"Missing"}]},"overall_score":2,"max_score":3,"summary":"Good code","issues":["Missing retry"],"strengths":["Clean"]}`,
			score:     2,
			maxScore:  3,
			criteria:  3,
			summary:   "Good code",
			issues:    1,
			strengths: 1,
		},
		{
			name:     "wrapped in markdown json fence",
			input:    "```json\n" + `{"scores":{"criteria":[{"name":"Code Builds","passed":true}]},"overall_score":1,"max_score":1,"summary":"Good","issues":[],"strengths":[]}` + "\n```",
			score:    1,
			maxScore: 1,
			criteria: 1,
			summary:  "Good",
		},
		{
			name:     "wrapped in plain markdown fence",
			input:    "```\n" + `{"scores":{"criteria":[{"name":"X","passed":false}]},"overall_score":0,"max_score":1,"summary":"Bad","issues":["everything"],"strengths":[]}` + "\n```",
			score:    0,
			maxScore: 1,
			criteria: 1,
			issues:   1,
		},
		{
			name:    "no json",
			input:   "I cannot review this code because...",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			input:   "   \n\t  \n  ",
			wantErr: true,
		},
		{
			name:     "auto-fill max_score from criteria count",
			input:    `{"scores":{"criteria":[{"name":"A","passed":true},{"name":"B","passed":true},{"name":"C","passed":false}]},"overall_score":0,"max_score":0,"summary":"test","issues":[],"strengths":[]}`,
			score:    2,
			maxScore: 3,
			criteria: 3,
		},
		{
			name:     "auto-fill overall_score from criteria",
			input:    `{"scores":{"criteria":[{"name":"A","passed":true},{"name":"B","passed":false}]},"summary":"test","issues":[],"strengths":[]}`,
			score:    1,
			maxScore: 2,
			criteria: 2,
		},
		{
			name:     "json with surrounding text",
			input:    `Here is my review: {"scores":{"criteria":[{"name":"Build","passed":true}]},"overall_score":1,"max_score":1,"summary":"ok","issues":[],"strengths":[]} End of review.`,
			score:    1,
			maxScore: 1,
			criteria: 1,
		},
		{
			name:      "all criteria passed",
			input:     `{"scores":{"criteria":[{"name":"A","passed":true},{"name":"B","passed":true}]},"overall_score":2,"max_score":2,"summary":"Perfect","issues":[],"strengths":["Great"]}`,
			score:     2,
			maxScore:  2,
			criteria:  2,
			strengths: 1,
		},
		{
			name:     "all criteria failed",
			input:    `{"scores":{"criteria":[{"name":"A","passed":false},{"name":"B","passed":false}]},"overall_score":0,"max_score":2,"summary":"Bad","issues":["A failed","B failed"],"strengths":[]}`,
			score:    0,
			maxScore: 2,
			criteria: 2,
			issues:   2,
		},
		{
			name:    "invalid json structure",
			input:   `{"scores": "not an object"}`,
			wantErr: true,
		},
		{
			name:     "empty criteria list",
			input:    `{"scores":{"criteria":[]},"overall_score":0,"max_score":0,"summary":"Nothing to evaluate","issues":[],"strengths":[]}`,
			score:    0,
			maxScore: 0,
			criteria: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseReviewResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.OverallScore != tt.score {
				t.Errorf("OverallScore = %d, want %d", result.OverallScore, tt.score)
			}
			if tt.maxScore > 0 && result.MaxScore != tt.maxScore {
				t.Errorf("MaxScore = %d, want %d", result.MaxScore, tt.maxScore)
			}
			if tt.criteria > 0 && len(result.Scores.Criteria) != tt.criteria {
				t.Errorf("Criteria count = %d, want %d", len(result.Scores.Criteria), tt.criteria)
			}
			if tt.summary != "" && result.Summary != tt.summary {
				t.Errorf("Summary = %q, want %q", result.Summary, tt.summary)
			}
			if tt.issues > 0 && len(result.Issues) != tt.issues {
				t.Errorf("Issues count = %d, want %d", len(result.Issues), tt.issues)
			}
			if tt.strengths > 0 && len(result.Strengths) != tt.strengths {
				t.Errorf("Strengths count = %d, want %d", len(result.Strengths), tt.strengths)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ReviewScores tests
// ---------------------------------------------------------------------------

func TestReviewScoresPassedCount(t *testing.T) {
	tests := []struct {
		name     string
		criteria []CriterionResult
		want     int
	}{
		{"all passed", []CriterionResult{
			{Name: "A", Passed: true},
			{Name: "B", Passed: true},
		}, 2},
		{"none passed", []CriterionResult{
			{Name: "A", Passed: false},
			{Name: "B", Passed: false},
		}, 0},
		{"mixed", []CriterionResult{
			{Name: "A", Passed: true},
			{Name: "B", Passed: false},
			{Name: "C", Passed: true},
		}, 2},
		{"empty", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ReviewScores{Criteria: tt.criteria}
			if got := s.PassedCount(); got != tt.want {
				t.Errorf("PassedCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestReviewScoresTotalCount(t *testing.T) {
	tests := []struct {
		name     string
		criteria []CriterionResult
		want     int
	}{
		{"three criteria", []CriterionResult{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		}, 3},
		{"empty", nil, 0},
		{"one", []CriterionResult{{Name: "A"}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ReviewScores{Criteria: tt.criteria}
			if got := s.TotalCount(); got != tt.want {
				t.Errorf("TotalCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestReviewScoresAllPassed(t *testing.T) {
	tests := []struct {
		name     string
		criteria []CriterionResult
		want     bool
	}{
		{"all passed", []CriterionResult{
			{Name: "A", Passed: true},
			{Name: "B", Passed: true},
		}, true},
		{"one failed", []CriterionResult{
			{Name: "A", Passed: true},
			{Name: "B", Passed: false},
		}, false},
		{"none passed", []CriterionResult{
			{Name: "A", Passed: false},
		}, false},
		{"empty returns false", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ReviewScores{Criteria: tt.criteria}
			if got := s.AllPassed(); got != tt.want {
				t.Errorf("AllPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StubReviewer tests
// ---------------------------------------------------------------------------

func TestStubReviewer(t *testing.T) {
	s := &StubReviewer{}
	result, err := s.Review(nil, "test prompt", "some-dir", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "Review skipped (stub evaluator)" {
		t.Errorf("unexpected summary: %s", result.Summary)
	}
}

func TestStubReviewerScores(t *testing.T) {
	s := &StubReviewer{}
	result, err := s.Review(nil, "prompt", "dir", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallScore != 1 {
		t.Errorf("OverallScore = %d, want 1", result.OverallScore)
	}
	if result.MaxScore != 1 {
		t.Errorf("MaxScore = %d, want 1", result.MaxScore)
	}
	if len(result.Scores.Criteria) != 1 {
		t.Fatalf("Criteria count = %d, want 1", len(result.Scores.Criteria))
	}
	c := result.Scores.Criteria[0]
	if c.Name != "stub_criterion" {
		t.Errorf("criterion name = %q, want %q", c.Name, "stub_criterion")
	}
	if !c.Passed {
		t.Error("stub criterion should pass")
	}
	if result.Issues == nil {
		t.Error("Issues should not be nil")
	}
	if result.Strengths == nil {
		t.Error("Strengths should not be nil")
	}
}

func TestStubReviewerIgnoresInputs(t *testing.T) {
	s := &StubReviewer{}
	r1, _ := s.Review(nil, "prompt1", "dir1", "ref1", "criteria1")
	r2, _ := s.Review(nil, "prompt2", "dir2", "ref2", "criteria2")

	if r1.Summary != r2.Summary {
		t.Error("stub reviewer should return identical results regardless of inputs")
	}
	if r1.OverallScore != r2.OverallScore {
		t.Error("stub reviewer should return identical scores regardless of inputs")
	}
}

// ---------------------------------------------------------------------------
// NewCopilotReviewer tests
// ---------------------------------------------------------------------------

func TestNewCopilotReviewerDefaultModel(t *testing.T) {
	r := NewCopilotReviewer(nil, "", 50)
	if r.model != "claude-sonnet-4.5" {
		t.Errorf("default model = %q, want %q", r.model, "claude-sonnet-4.5")
	}
}

func TestNewCopilotReviewerCustomModel(t *testing.T) {
	r := NewCopilotReviewer(nil, "gpt-4o", 100)
	if r.model != "gpt-4o" {
		t.Errorf("model = %q, want %q", r.model, "gpt-4o")
	}
	if r.maxSessionActions != 100 {
		t.Errorf("maxSessionActions = %d, want 100", r.maxSessionActions)
	}
}

func TestCopilotReviewerSetSkillDirectories(t *testing.T) {
	r := NewCopilotReviewer(nil, "", 50)
	dirs := []string{"/skills/gen", "/skills/rev"}
	r.SetSkillDirectories(dirs)
	if len(r.skillDirectories) != 2 {
		t.Errorf("skillDirectories count = %d, want 2", len(r.skillDirectories))
	}
	for i, d := range dirs {
		if r.skillDirectories[i] != d {
			t.Errorf("skillDirectories[%d] = %q, want %q", i, r.skillDirectories[i], d)
		}
	}
}

func TestCopilotReviewerSetSessionTimeout(t *testing.T) {
	r := NewCopilotReviewer(nil, "", 50)
	r.SetSessionTimeout(5 * time.Minute)
	if r.sessionTimeout != 5*time.Minute {
		t.Errorf("sessionTimeout = %v, want %v", r.sessionTimeout, 5*time.Minute)
	}
}

// ---------------------------------------------------------------------------
// PanelReviewer construction tests
// ---------------------------------------------------------------------------

func TestNewPanelReviewer(t *testing.T) {
	models := []string{"model-a", "model-b", "model-c"}
	p := NewPanelReviewer(nil, models, 25)

	if len(p.models) != 3 {
		t.Fatalf("model count = %d, want 3", len(p.models))
	}
	if p.maxSessionActions != 25 {
		t.Errorf("maxSessionActions = %d, want 25", p.maxSessionActions)
	}
}

func TestPanelReviewerModels(t *testing.T) {
	models := []string{"a", "b"}
	p := NewPanelReviewer(nil, models, 10)
	got := p.Models()
	if len(got) != len(models) {
		t.Fatalf("Models() returned %d items, want %d", len(got), len(models))
	}
	for i, m := range models {
		if got[i] != m {
			t.Errorf("Models()[%d] = %q, want %q", i, got[i], m)
		}
	}
}

func TestPanelReviewerSetSkillDirectories(t *testing.T) {
	p := NewPanelReviewer(nil, []string{"m"}, 10)
	dirs := []string{"/a", "/b"}
	p.SetSkillDirectories(dirs)
	if len(p.skillDirectories) != 2 {
		t.Errorf("skillDirectories = %d, want 2", len(p.skillDirectories))
	}
}

func TestPanelReviewerSetSessionTimeout(t *testing.T) {
	p := NewPanelReviewer(nil, []string{"m"}, 10)
	p.SetSessionTimeout(3 * time.Minute)
	if p.sessionTimeout != 3*time.Minute {
		t.Errorf("sessionTimeout = %v, want %v", p.sessionTimeout, 3*time.Minute)
	}
}

// ---------------------------------------------------------------------------
// averageReview tests
// ---------------------------------------------------------------------------

func TestAverageReviewEmpty(t *testing.T) {
	result := averageReview(nil)
	if result.Summary != "No reviews to consolidate" {
		t.Errorf("Summary = %q, want %q", result.Summary, "No reviews to consolidate")
	}
}

func TestAverageReviewSingleReviewer(t *testing.T) {
	panel := []ReviewResult{{
		Model: "model-a",
		Scores: ReviewScores{Criteria: []CriterionResult{
			{Name: "Build", Passed: true, Reason: "ok"},
			{Name: "Style", Passed: false, Reason: "messy"},
		}},
		OverallScore: 1,
		MaxScore:     2,
		Summary:      "Decent",
		Issues:       []string{"messy code"},
		Strengths:    []string{"compiles"},
	}}

	result := averageReview(panel)

	if result.OverallScore != 1 {
		t.Errorf("OverallScore = %d, want 1", result.OverallScore)
	}
	if result.MaxScore != 2 {
		t.Errorf("MaxScore = %d, want 2", result.MaxScore)
	}
	// With 1 reviewer: 1/1 > 1/2 = true for Build, 0/1 > 0 = false for Style
	buildPassed := false
	stylePassed := false
	for _, c := range result.Scores.Criteria {
		if c.Name == "Build" {
			buildPassed = c.Passed
		}
		if c.Name == "Style" {
			stylePassed = c.Passed
		}
	}
	if !buildPassed {
		t.Error("Build should pass with 1/1 majority")
	}
	if stylePassed {
		t.Error("Style should fail with 0/1 majority")
	}
}

func TestAverageReviewMajorityVoting(t *testing.T) {
	panel := []ReviewResult{
		{
			Model: "m1",
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "Build", Passed: true},
				{Name: "Style", Passed: true},
				{Name: "Errors", Passed: false},
			}},
			Issues:    []string{"no retries"},
			Strengths: []string{"clean"},
		},
		{
			Model: "m2",
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "Build", Passed: true},
				{Name: "Style", Passed: false},
				{Name: "Errors", Passed: true},
			}},
			Issues:    []string{"inconsistent style"},
			Strengths: []string{"handles errors"},
		},
		{
			Model: "m3",
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "Build", Passed: true},
				{Name: "Style", Passed: false},
				{Name: "Errors", Passed: true},
			}},
			Issues:    []string{"no retries"},
			Strengths: []string{"clean"},
		},
	}

	result := averageReview(panel)

	criteriaMap := map[string]bool{}
	for _, c := range result.Scores.Criteria {
		criteriaMap[c.Name] = c.Passed
	}

	// Build: 3/3 pass → pass
	if !criteriaMap["Build"] {
		t.Error("Build should pass (3/3 majority)")
	}
	// Style: 1/3 pass → fail (1 > 1 is false)
	if criteriaMap["Style"] {
		t.Error("Style should fail (1/3 majority)")
	}
	// Errors: 2/3 pass → pass (2 > 1 is true)
	if !criteriaMap["Errors"] {
		t.Error("Errors should pass (2/3 majority)")
	}

	// Verify correct overall score: Build + Errors = 2 passed
	if result.OverallScore != 2 {
		t.Errorf("OverallScore = %d, want 2", result.OverallScore)
	}
	if result.MaxScore != 3 {
		t.Errorf("MaxScore = %d, want 3", result.MaxScore)
	}
}

func TestAverageReviewDeduplicatesIssuesAndStrengths(t *testing.T) {
	panel := []ReviewResult{
		{
			Scores:    ReviewScores{Criteria: []CriterionResult{{Name: "A", Passed: true}}},
			Issues:    []string{"dup issue", "unique1"},
			Strengths: []string{"dup strength", "unique_s1"},
		},
		{
			Scores:    ReviewScores{Criteria: []CriterionResult{{Name: "A", Passed: true}}},
			Issues:    []string{"dup issue", "unique2"},
			Strengths: []string{"dup strength", "unique_s2"},
		},
	}

	result := averageReview(panel)

	if len(result.Issues) != 3 {
		t.Errorf("Issues count = %d, want 3 (dedup 'dup issue')", len(result.Issues))
	}
	if len(result.Strengths) != 3 {
		t.Errorf("Strengths count = %d, want 3 (dedup 'dup strength')", len(result.Strengths))
	}
}

func TestAverageReviewDisjointCriteria(t *testing.T) {
	panel := []ReviewResult{
		{
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "Build", Passed: true},
			}},
		},
		{
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "Style", Passed: false},
			}},
		},
	}

	result := averageReview(panel)

	if len(result.Scores.Criteria) != 2 {
		t.Errorf("Criteria count = %d, want 2 (union of disjoint sets)", len(result.Scores.Criteria))
	}
	// Build: 1/1 → pass; Style: 0/1 → fail
	criteriaMap := map[string]bool{}
	for _, c := range result.Scores.Criteria {
		criteriaMap[c.Name] = c.Passed
	}
	if !criteriaMap["Build"] {
		t.Error("Build should pass (1/1)")
	}
	if criteriaMap["Style"] {
		t.Error("Style should fail (0/1)")
	}
}

func TestAverageReviewSummaryFormat(t *testing.T) {
	panel := []ReviewResult{
		{
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "A", Passed: true},
			}},
		},
		{
			Scores: ReviewScores{Criteria: []CriterionResult{
				{Name: "A", Passed: true},
			}},
		},
	}

	result := averageReview(panel)

	if !strings.Contains(result.Summary, "2 reviewers") {
		t.Errorf("Summary should mention reviewer count, got: %q", result.Summary)
	}
	if result.Model != "consensus (average)" {
		t.Errorf("Model = %q, want %q", result.Model, "consensus (average)")
	}
}

func TestAverageReviewPreservesCriteriaOrder(t *testing.T) {
	panel := []ReviewResult{{
		Scores: ReviewScores{Criteria: []CriterionResult{
			{Name: "Build", Passed: true},
			{Name: "Style", Passed: true},
			{Name: "Errors", Passed: true},
			{Name: "Docs", Passed: true},
		}},
	}}

	result := averageReview(panel)

	expected := []string{"Build", "Style", "Errors", "Docs"}
	for i, c := range result.Scores.Criteria {
		if c.Name != expected[i] {
			t.Errorf("Criteria[%d].Name = %q, want %q", i, c.Name, expected[i])
		}
	}
}

func TestAverageReviewEvenSplitFailsByCriteria(t *testing.T) {
	// With 2 reviewers, 1 pass + 1 fail → passCount=1, total=2 → 1 > 2/2=1 → false (tie fails)
	panel := []ReviewResult{
		{Scores: ReviewScores{Criteria: []CriterionResult{{Name: "X", Passed: true}}}},
		{Scores: ReviewScores{Criteria: []CriterionResult{{Name: "X", Passed: false}}}},
	}

	result := averageReview(panel)

	if len(result.Scores.Criteria) != 1 {
		t.Fatal("expected 1 criterion")
	}
	if result.Scores.Criteria[0].Passed {
		t.Error("tie (1/2) should fail — majority requires strictly more than half")
	}
}

// ---------------------------------------------------------------------------
// copyDirToTemp tests
// ---------------------------------------------------------------------------

func TestCopyDirToTemp(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.go"), []byte("package main"), 0644)
	sub := filepath.Join(src, "pkg")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "lib.go"), []byte("package pkg"), 0644)

	dst, err := copyDirToTemp(src, "hyoka-test-*")
	if err != nil {
		t.Fatalf("copyDirToTemp failed: %v", err)
	}
	defer os.RemoveAll(dst)

	data, err := os.ReadFile(filepath.Join(dst, "main.go"))
	if err != nil {
		t.Fatalf("failed to read copied main.go: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("main.go content = %q, want %q", string(data), "package main")
	}

	data, err = os.ReadFile(filepath.Join(dst, "pkg", "lib.go"))
	if err != nil {
		t.Fatalf("failed to read copied pkg/lib.go: %v", err)
	}
	if string(data) != "package pkg" {
		t.Errorf("pkg/lib.go content = %q, want %q", string(data), "package pkg")
	}
}

func TestCopyDirToTempSkipsDotDirs(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.go"), []byte("package main"), 0644)
	hidden := filepath.Join(src, ".git")
	os.MkdirAll(hidden, 0755)
	os.WriteFile(filepath.Join(hidden, "config"), []byte("gitconfig"), 0644)

	dst, err := copyDirToTemp(src, "hyoka-test-*")
	if err != nil {
		t.Fatalf("copyDirToTemp failed: %v", err)
	}
	defer os.RemoveAll(dst)

	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Error("hidden .git directory should not be copied")
	}
}

func TestCopyDirToTempSkipsBuildArtifactDirs(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.go"), []byte("package main"), 0644)
	nm := filepath.Join(src, "node_modules")
	os.MkdirAll(nm, 0755)
	os.WriteFile(filepath.Join(nm, "pkg.json"), []byte("{}"), 0644)

	dst, err := copyDirToTemp(src, "hyoka-test-*")
	if err != nil {
		t.Fatalf("copyDirToTemp failed: %v", err)
	}
	defer os.RemoveAll(dst)

	if _, err := os.Stat(filepath.Join(dst, "node_modules")); !os.IsNotExist(err) {
		t.Error("node_modules should be skipped as build artifact dir")
	}
}

func TestCopyDirToTempEmptyDir(t *testing.T) {
	src := t.TempDir()

	dst, err := copyDirToTemp(src, "hyoka-test-*")
	if err != nil {
		t.Fatalf("copyDirToTemp failed: %v", err)
	}
	defer os.RemoveAll(dst)

	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty directory, got %d entries", len(entries))
	}
}

// ---------------------------------------------------------------------------
// ReviewResult / ReviewEvent structural tests
// ---------------------------------------------------------------------------

func TestReviewResultZeroValue(t *testing.T) {
	var r ReviewResult
	if r.OverallScore != 0 {
		t.Errorf("zero-value OverallScore = %d", r.OverallScore)
	}
	if r.Scores.PassedCount() != 0 {
		t.Error("zero-value PassedCount should be 0")
	}
	if r.Scores.AllPassed() {
		t.Error("zero-value AllPassed should be false")
	}
}

func TestReviewEventFields(t *testing.T) {
	evt := ReviewEvent{
		Type:     "tool_execution_complete",
		ToolName: "read_file",
		ToolArgs: `{"path": "main.go"}`,
		Content:  "file content here",
		Result:   "success",
		Error:    "",
		Duration: 123.45,
	}
	if evt.Type != "tool_execution_complete" {
		t.Error("Type mismatch")
	}
	if evt.Duration != 123.45 {
		t.Errorf("Duration = %f, want 123.45", evt.Duration)
	}
}

// ---------------------------------------------------------------------------
// Reviewer interface compliance
// ---------------------------------------------------------------------------

func TestStubReviewerImplementsReviewer(t *testing.T) {
	var _ Reviewer = &StubReviewer{}
}

func TestPanelReviewerImplementsReviewer(t *testing.T) {
	var _ Reviewer = &PanelReviewer{}
}

func TestCopilotReviewerImplementsReviewer(t *testing.T) {
	var _ Reviewer = &CopilotReviewer{}
}
