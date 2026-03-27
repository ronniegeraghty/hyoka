package checkenv

import (
	"testing"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Python 3.12.1", "Python 3.12.1"},
		{"go version go1.26.1 linux/amd64", "go version go1.26.1 linux/amd64"},
		{"v22.5.0\nmore stuff", "v22.5.0"},
		{"", ""},
		{"single line", "single line"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractVersion(tt.input)
			if got != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPrintCheck(t *testing.T) {
	// Smoke test — just ensure no panics
	printCheck(checkResult{name: "Test", ok: true, version: "v1.0"})
	printCheck(checkResult{name: "Test", ok: false, hint: "not found"})
}

func TestCheckFunctionsDoNotPanic(t *testing.T) {
	// These run real commands but should never panic, even if tools are missing
	checks := []func() checkResult{
		checkPython,
		checkDotnet,
		checkGo,
		checkNode,
		checkJava,
		checkRust,
		checkCpp,
		checkCopilotCLI,
		checkCopilotAuth,
		checkNpx,
	}
	for _, fn := range checks {
		result := fn()
		if result.name == "" {
			t.Error("check returned empty name")
		}
	}
}
