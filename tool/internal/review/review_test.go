package review

import (
	"testing"
)

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
		if !contains(result, check) {
			t.Errorf("review prompt missing %q", check)
		}
	}
}

func TestBuildReviewPromptNoReference(t *testing.T) {
	prompt := "Write code"
	generated := map[string]string{"main.go": "package main"}

	result := BuildReviewPrompt(prompt, generated, nil, "")

	if !contains(result, "No reference answer provided") {
		t.Error("expected 'No reference answer provided' when no reference given")
	}
}

func TestBuildReviewPromptWithEvaluationCriteria(t *testing.T) {
	prompt := "Write Azure code"
	generated := map[string]string{"main.go": "package main"}
	criteria := "- Must use DefaultAzureCredential\n- Must handle errors"

	result := BuildReviewPrompt(prompt, generated, nil, criteria)

	if !contains(result, "Prompt-Specific Evaluation Criteria") {
		t.Error("expected evaluation criteria section")
	}
	if !contains(result, "DefaultAzureCredential") {
		t.Error("expected criteria content in prompt")
	}
}

func TestParseReviewResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		score   int
	}{
		{
			name:  "clean json with criteria",
			input: `{"scores":{"criteria":[{"name":"Code Builds","passed":true,"reason":"OK"},{"name":"Best Practices","passed":true,"reason":"Good"},{"name":"Error Handling","passed":false,"reason":"Missing"}]},"overall_score":2,"max_score":3,"summary":"Good code","issues":["Missing retry"],"strengths":["Clean"]}`,
			score: 2,
		},
		{
			name:  "wrapped in markdown",
			input: "```json\n" + `{"scores":{"criteria":[{"name":"Code Builds","passed":true}]},"overall_score":1,"max_score":1,"summary":"Good","issues":[],"strengths":[]}` + "\n```",
			score: 1,
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
				t.Errorf("expected overall score %d, got %d", tt.score, result.OverallScore)
			}
		})
	}
}

func TestStubReviewer(t *testing.T) {
	s := &StubReviewer{}
	result, err := s.Review(nil, "test prompt", "/tmp/test", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "Review skipped (stub evaluator)" {
		t.Errorf("unexpected summary: %s", result.Summary)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
