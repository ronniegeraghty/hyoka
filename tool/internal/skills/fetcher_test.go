package skills

import (
	"os"
	"testing"

	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
)

func TestFetchEmpty(t *testing.T) {
	dir, err := Fetch(nil)
	if err != nil {
		t.Fatalf("Fetch(nil) returned error: %v", err)
	}
	if dir != "" {
		t.Fatalf("expected empty dir for nil input, got %q", dir)
	}

	dir, err = Fetch([]config.RemoteSkill{})
	if err != nil {
		t.Fatalf("Fetch([]) returned error: %v", err)
	}
	if dir != "" {
		t.Fatalf("expected empty dir for empty input, got %q", dir)
	}
}

func TestCleanupNoop(t *testing.T) {
	// Cleanup with empty string should not panic
	Cleanup("")
}

func TestCleanupRemovesDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "skills-cleanup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	// Write a file so the dir isn't empty
	os.WriteFile(dir+"/test.txt", []byte("hello"), 0644)

	Cleanup(dir)

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected dir %q to be removed, but it still exists", dir)
	}
}
