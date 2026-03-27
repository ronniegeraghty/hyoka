package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/config"
)

func TestResolveLocal_DirectPath(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	dirs, err := ResolveSkillDirs([]config.Skill{
		{Type: "local", Path: skillDir},
	}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(dirs))
	}
	if dirs[0] != skillDir {
		t.Errorf("expected %q, got %q", skillDir, dirs[0])
	}
}

func TestResolveLocal_RelativePath(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "skills", "generator")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	dirs, err := ResolveSkillDirs([]config.Skill{
		{Type: "local", Path: "skills/generator"},
	}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(dirs))
	}
}

func TestResolveLocal_GlobPattern(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"skill-a", "skill-b", "not-a-skill.txt"} {
		p := filepath.Join(dir, "skills", name)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a file (not a dir) that matches the glob
	f := filepath.Join(dir, "skills", "not-a-skill.txt", "dummy")
	os.Remove(filepath.Join(dir, "skills", "not-a-skill.txt"))
	os.WriteFile(filepath.Join(dir, "skills", "readme.txt"), []byte("hi"), 0644)
	_ = f

	dirs, err := ResolveSkillDirs([]config.Skill{
		{Type: "local", Path: filepath.Join(dir, "skills", "*")},
	}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should get skill-a and skill-b (directories only)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 dirs from glob, got %d: %v", len(dirs), dirs)
	}
}

func TestResolveLocal_EmptySkills(t *testing.T) {
	dirs, err := ResolveSkillDirs(nil, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected 0 dirs, got %d", len(dirs))
	}
}

func TestResolveLocal_InvalidType(t *testing.T) {
	_, err := ResolveSkillDirs([]config.Skill{
		{Type: "unknown", Path: "/some/path"},
	}, ".")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
