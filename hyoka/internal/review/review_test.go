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

func TestBuildReviewPromptTiered(t *testing.T) {
	prompt := "Write Azure Key Vault code"
	generated := map[string]string{"Main.java": "import com.azure.security.keyvault.secrets.*;"}
	attrCriteria := "1. **Maven Deps** — Correct Maven dependencies.\n2. **Try-With-Resources** — AutoCloseable clients.\n"
	promptCriteria := "- Must use DefaultAzureCredential"

	result := BuildReviewPromptTiered(prompt, generated, nil, attrCriteria, promptCriteria)

	checks := []string{
		"Attribute-Matched Evaluation Criteria",
		"Maven Deps",
		"Try-With-Resources",
		"Prompt-Specific Evaluation Criteria",
		"DefaultAzureCredential",
		"Scoring Rubric",
	}
	for _, check := range checks {
		if !contains(result, check) {
			t.Errorf("tiered review prompt missing %q", check)
		}
	}
}

func TestBuildReviewPromptTieredNoAttrCriteria(t *testing.T) {
	prompt := "Write code"
	generated := map[string]string{"main.go": "package main"}
	promptCriteria := "- Must handle errors"

	result := BuildReviewPromptTiered(prompt, generated, nil, "", promptCriteria)

	if contains(result, "Attribute-Matched") {
		t.Error("should not contain attribute-matched section when empty")
	}
	if !contains(result, "Prompt-Specific Evaluation Criteria") {
		t.Error("should contain prompt-specific criteria")
	}
}

func TestBuildReviewPromptTieredBackwardCompat(t *testing.T) {
	// BuildReviewPrompt (legacy) should still work, delegating to BuildReviewPromptTiered
	prompt := "Write code"
	generated := map[string]string{"main.go": "package main"}
	criteria := "- Must use DefaultAzureCredential"

	result := BuildReviewPrompt(prompt, generated, nil, criteria)

	if !contains(result, "Prompt-Specific Evaluation Criteria") {
		t.Error("legacy BuildReviewPrompt should include prompt-specific criteria")
	}
	if contains(result, "Attribute-Matched") {
		t.Error("legacy BuildReviewPrompt should not include attribute-matched section")
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
	result, err := s.Review(nil, "test prompt", "/tmp/test", "", "", "")
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
