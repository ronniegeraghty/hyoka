package eval

import (
	"testing"
)

func TestHyokaBaseEnv(t *testing.T) {
	env := HyokaBaseEnv()
	if len(env) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(env))
	}
	if env[0] != "HYOKA_SESSION=true" {
		t.Errorf("expected HYOKA_SESSION=true, got %q", env[0])
	}
}

func TestHyokaEvalEnv(t *testing.T) {
	env := HyokaEvalEnv("azure-sdk-go", "gpt-4o")
	if len(env) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(env))
	}
	expected := map[string]bool{
		"HYOKA_SESSION=true":          true,
		"HYOKA_PROMPT_ID=azure-sdk-go": true,
		"HYOKA_CONFIG=gpt-4o":         true,
	}
	for _, e := range env {
		if !expected[e] {
			t.Errorf("unexpected env entry: %q", e)
		}
	}
}

func TestHyokaEvalEnvEmpty(t *testing.T) {
	env := HyokaEvalEnv("", "")
	if len(env) != 1 {
		t.Fatalf("expected 1 entry (session only), got %d", len(env))
	}
	if env[0] != "HYOKA_SESSION=true" {
		t.Errorf("expected HYOKA_SESSION=true, got %q", env[0])
	}
}

func TestHyokaEvalEnvPartial(t *testing.T) {
	env := HyokaEvalEnv("my-prompt", "")
	if len(env) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(env))
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
