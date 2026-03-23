package verify

import (
	"testing"
)

func TestBuildVerifyPrompt(t *testing.T) {
	prompt := "Write Azure Blob Storage auth code"
	generated := map[string]string{
		"Program.cs": "using Azure.Storage.Blobs;\n// ...",
	}

	result := buildVerifyPrompt(prompt, generated, "Should use DefaultAzureCredential")

	checks := []string{
		"Original Prompt",
		"Expected Coverage",
		"Generated Code",
		"Evaluation Criteria",
		"DefaultAzureCredential",
		"Program.cs",
	}
	for _, check := range checks {
		if !containsStr(result, check) {
			t.Errorf("verify prompt missing %q", check)
		}
	}
}

func TestBuildVerifyPromptNoExpectedCoverage(t *testing.T) {
	prompt := "Write code"
	generated := map[string]string{"main.go": "package main"}

	result := buildVerifyPrompt(prompt, generated, "")

	if containsStr(result, "Expected Coverage") {
		t.Error("should not include Expected Coverage section when empty")
	}
}

func TestParseVerifyResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		pass    bool
	}{
		{
			name:  "pass",
			input: `{"pass": true, "reasoning": "Code looks good", "summary": "All requirements met"}`,
			pass:  true,
		},
		{
			name:  "fail",
			input: `{"pass": false, "reasoning": "Missing auth", "summary": "Missing authentication"}`,
			pass:  false,
		},
		{
			name:  "wrapped in markdown",
			input: "```json\n" + `{"pass": true, "reasoning": "OK", "summary": "OK"}` + "\n```",
			pass:  true,
		},
		{
			name:    "no json",
			input:   "I cannot verify this.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVerifyResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Pass != tt.pass {
				t.Errorf("expected pass=%v, got %v", tt.pass, result.Pass)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
