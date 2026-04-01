package eval

import (
	"os"
	"strings"
	"testing"
)

func TestHyokaBaseEnv(t *testing.T) {
	env := HyokaBaseEnv()
	// Should contain the full process environment plus our marker.
	osEnv := os.Environ()
	if len(env) != len(osEnv)+1 {
		t.Fatalf("expected %d entries (os.Environ + 1), got %d", len(osEnv)+1, len(env))
	}
	// Last entry should be our marker.
	last := env[len(env)-1]
	if last != "HYOKA_SESSION=true" {
		t.Errorf("expected last entry HYOKA_SESSION=true, got %q", last)
	}
	// PATH must be inherited.
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected PATH to be inherited from os.Environ()")
	}
}

func TestHyokaEvalEnv(t *testing.T) {
	env := HyokaEvalEnv("azure-sdk-go", "gpt-4o")
	osEnv := os.Environ()
	// os.Environ + HYOKA_SESSION + HYOKA_PROMPT_ID + HYOKA_CONFIG
	if len(env) != len(osEnv)+3 {
		t.Fatalf("expected %d entries, got %d", len(osEnv)+3, len(env))
	}
	expected := map[string]bool{
		"HYOKA_SESSION=true":           false,
		"HYOKA_PROMPT_ID=azure-sdk-go": false,
		"HYOKA_CONFIG=gpt-4o":          false,
	}
	for _, e := range env {
		if _, ok := expected[e]; ok {
			expected[e] = true
		}
	}
	for k, v := range expected {
		if !v {
			t.Errorf("missing expected env entry: %q", k)
		}
	}
}

func TestHyokaEvalEnvEmpty(t *testing.T) {
	env := HyokaEvalEnv("", "")
	osEnv := os.Environ()
	if len(env) != len(osEnv)+1 {
		t.Fatalf("expected %d entries (session only), got %d", len(osEnv)+1, len(env))
	}
}

func TestHyokaEvalEnvPartial(t *testing.T) {
	env := HyokaEvalEnv("my-prompt", "")
	osEnv := os.Environ()
	if len(env) != len(osEnv)+2 {
		t.Fatalf("expected %d entries, got %d", len(osEnv)+2, len(env))
	}
}

func TestEnvConstants(t *testing.T) {
	if EnvHyokaSession != "HYOKA_SESSION" {
		t.Errorf("unexpected session const: %q", EnvHyokaSession)
	}
	if EnvHyokaPromptID != "HYOKA_PROMPT_ID" {
		t.Errorf("unexpected prompt const: %q", EnvHyokaPromptID)
	}
	if EnvHyokaConfig != "HYOKA_CONFIG" {
		t.Errorf("unexpected config const: %q", EnvHyokaConfig)
	}
}
