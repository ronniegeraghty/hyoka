package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

const testPromptContent = `---
id: storage-auth-dotnet
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: beginner
description: "Authenticate to Azure Blob Storage using DefaultAzureCredential"
sdk_package: Azure.Storage.Blobs
doc_url: https://learn.microsoft.com/en-us/dotnet/api/azure.storage.blobs
tags:
  - authentication
  - blob
  - identity
created: "2024-01-15"
author: test
expected_packages:
  - Azure.Storage.Blobs
  - Azure.Identity
expected_tools:
  - create_file
  - run_terminal_command
---

# Storage Authentication (.NET)

## Prompt

Write a C# console application that authenticates to Azure Blob Storage
using DefaultAzureCredential and lists all containers in the storage account.

## Notes

This is a beginner-level prompt for testing.
`

func TestParsePromptFile(t *testing.T) {
	p, err := ParsePromptFile([]byte(testPromptContent), "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.ID != "storage-auth-dotnet" {
		t.Errorf("expected ID 'storage-auth-dotnet', got %q", p.ID)
	}
	if p.Service != "storage" {
		t.Errorf("expected service 'storage', got %q", p.Service)
	}
	if p.Plane != "data-plane" {
		t.Errorf("expected plane 'data-plane', got %q", p.Plane)
	}
	if p.Language != "dotnet" {
		t.Errorf("expected language 'dotnet', got %q", p.Language)
	}
	if p.Category != "authentication" {
		t.Errorf("expected category 'authentication', got %q", p.Category)
	}
	if len(p.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(p.Tags))
	}
	if len(p.ExpectedPkgs) != 2 {
		if len(p.ExpectedTools) != 2 {
			t.Errorf("expected 2 expected_tools, got %d", len(p.ExpectedTools))
		}
		t.Errorf("expected 2 expected_packages, got %d", len(p.ExpectedPkgs))
		if len(p.ExpectedTools) != 2 {
			t.Errorf("expected 2 expected_tools, got %d", len(p.ExpectedTools))
		}
	}
	if len(p.ExpectedTools) != 2 {
		t.Errorf("expected 2 expected_tools, got %d", len(p.ExpectedTools))
	}
	if p.PromptText == "" {
		t.Error("expected non-empty prompt text")
	}
	if p.FilePath != "test.prompt.md" {
		t.Errorf("expected file path 'test.prompt.md', got %q", p.FilePath)
	}
}

func TestParsePromptFileMissingFrontmatter(t *testing.T) {
	_, err := ParsePromptFile([]byte("no frontmatter here"), "bad.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParsePromptFileMissingID(t *testing.T) {
	content := []byte("---\nservice: storage\n---\n## Prompt\nHello\n")
	_, err := ParsePromptFile(content, "no-id.prompt.md")
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestParsePromptFileMissingClosingDelimiter(t *testing.T) {
	content := []byte("---\nid: test\nno closing delimiter")
	_, err := ParsePromptFile(content, "bad.prompt.md")
	if err == nil {
		t.Fatal("expected error for missing closing delimiter")
	}
}

func TestLoadPrompts(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "storage", "data-plane", "dotnet")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subDir, "auth.prompt.md"), []byte(testPromptContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	// Write a non-prompt file that should be ignored
	if err := os.WriteFile(filepath.Join(subDir, "readme.md"), []byte("# Readme"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	prompts, err := LoadPrompts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].ID != "storage-auth-dotnet" {
		t.Errorf("expected ID 'storage-auth-dotnet', got %q", prompts[0].ID)
	}
}

func TestFilterPrompts(t *testing.T) {
	prompts := []*Prompt{
		{ID: "p1", Service: "storage", Language: "dotnet", Plane: "data-plane", Category: "authentication", Tags: []string{"blob", "identity"}},
		{ID: "p2", Service: "keyvault", Language: "python", Plane: "data-plane", Category: "encryption", Tags: []string{"keys"}},
		{ID: "p3", Service: "storage", Language: "java", Plane: "management-plane", Category: "authentication", Tags: []string{"blob"}},
	}

	tests := []struct {
		name     string
		filter   Filter
		expected int
	}{
		{"no filter", Filter{}, 3},
		{"by service", Filter{Service: "storage"}, 2},
		{"by language", Filter{Language: "dotnet"}, 1},
		{"by plane", Filter{Plane: "data-plane"}, 2},
		{"by category", Filter{Category: "authentication"}, 2},
		{"by tags", Filter{Tags: []string{"blob"}}, 2},
		{"by multiple tags", Filter{Tags: []string{"blob", "identity"}}, 1},
		{"by prompt ID", Filter{PromptID: "p2"}, 1},
		{"combined filters", Filter{Service: "storage", Language: "dotnet"}, 1},
		{"no match", Filter{Service: "cosmos"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterPrompts(prompts, tt.filter)
			if len(result) != tt.expected {
				t.Errorf("expected %d results, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestLoadPromptsNonexistentDir(t *testing.T) {
	_, err := LoadPrompts("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestScanNearMissesHyphenated(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth-prompt.md"), []byte("---\nid: x\n---\n"), 0644)
	misses := ScanNearMisses(dir)
	if len(misses) != 1 {
		t.Fatalf("expected 1 near miss, got %d", len(misses))
	}
	if misses[0] != "auth-prompt.md" {
		t.Errorf("expected 'auth-prompt.md', got %q", misses[0])
	}
}

func TestScanNearMissesWrongExtension(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth.prompt.txt"), []byte("content"), 0644)
	misses := ScanNearMisses(dir)
	if len(misses) != 1 {
		t.Fatalf("expected 1 near miss, got %d", len(misses))
	}
}

func TestScanNearMissesMdWithFrontmatter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("---\nid: x\n---\ncontent"), 0644)
	misses := ScanNearMisses(dir)
	if len(misses) != 1 {
		t.Fatalf("expected 1 near miss, got %d", len(misses))
	}
}

func TestScanNearMissesIgnoresCorrectFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth.prompt.md"), []byte("---\nid: x\n---\n"), 0644)
	misses := ScanNearMisses(dir)
	if len(misses) != 0 {
		t.Errorf("expected 0 near misses for correct file, got %d", len(misses))
	}
}

func TestScanNearMissesEmpty(t *testing.T) {
	dir := t.TempDir()
	misses := ScanNearMisses(dir)
	if len(misses) != 0 {
		t.Errorf("expected 0 near misses for empty dir, got %d", len(misses))
	}
}

func TestSuggestFix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth-prompt.md", "auth.prompt.md"},
		{"auth.prompt.txt", "auth.prompt.md"},
		{"sub/auth-prompt.md", "sub/auth.prompt.md"},
		{"readme.md", ""},
	}
	for _, tt := range tests {
		got := suggestFix(tt.input)
		if got != tt.expected {
			t.Errorf("suggestFix(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestLoadPromptsZeroPromptsWithNearMisses(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth-prompt.md"), []byte("---\nid: x\n---\n## Prompt\nhello\n"), 0644)
	_, err := LoadPrompts(dir)
	if err == nil {
		t.Fatal("expected error for zero prompts")
	}
	errMsg := err.Error()
	if !filepath.IsAbs(dir) {
		t.Skip()
	}
	if len(errMsg) == 0 {
		t.Fatal("expected non-empty error message")
	}
	// Should contain near-miss suggestions
	if !contains(errMsg, "Did you mean") {
		t.Errorf("expected near-miss suggestion in error, got %q", errMsg)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

const testYAMLPromptContent = `id: storage-auth-dotnet
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: beginner
description: "Authenticate to Azure Blob Storage using DefaultAzureCredential"
sdk_package: Azure.Storage.Blobs
doc_url: https://learn.microsoft.com/en-us/dotnet/api/azure.storage.blobs
tags:
  - authentication
  - blob
  - identity
created: "2024-01-15"
author: test
expected_packages:
  - Azure.Storage.Blobs
  - Azure.Identity
expected_tools:
  - create_file
  - run_terminal_command
prompt_text: |
  Write a C# console application that authenticates to Azure Blob Storage
  using DefaultAzureCredential and lists all containers in the storage account.
evaluation_criteria: |
  Must use DefaultAzureCredential from Azure.Identity.
`

func TestParsePromptYAML(t *testing.T) {
p, err := ParsePromptYAML([]byte(testYAMLPromptContent), "test.prompt.yaml")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if p.ID != "storage-auth-dotnet" {
t.Errorf("expected ID 'storage-auth-dotnet', got %q", p.ID)
}
if p.Service != "storage" {
t.Errorf("expected service 'storage', got %q", p.Service)
}
if p.Plane != "data-plane" {
t.Errorf("expected plane 'data-plane', got %q", p.Plane)
}
if p.Language != "dotnet" {
t.Errorf("expected language 'dotnet', got %q", p.Language)
}
if p.Category != "authentication" {
t.Errorf("expected category 'authentication', got %q", p.Category)
}
if p.Difficulty != "beginner" {
t.Errorf("expected difficulty 'beginner', got %q", p.Difficulty)
}
if p.SDKPackage != "Azure.Storage.Blobs" {
t.Errorf("expected sdk_package 'Azure.Storage.Blobs', got %q", p.SDKPackage)
}
if len(p.Tags) != 3 {
t.Errorf("expected 3 tags, got %d", len(p.Tags))
}
if len(p.ExpectedPkgs) != 2 {
t.Errorf("expected 2 expected_packages, got %d", len(p.ExpectedPkgs))
}
if len(p.ExpectedTools) != 2 {
t.Errorf("expected 2 expected_tools, got %d", len(p.ExpectedTools))
}
if p.PromptText == "" {
t.Error("expected non-empty prompt text")
}
if p.EvaluationCriteria == "" {
t.Error("expected non-empty evaluation criteria")
}
if p.FilePath != "test.prompt.yaml" {
t.Errorf("expected file path 'test.prompt.yaml', got %q", p.FilePath)
}
}

func TestParsePromptYAMLMissingID(t *testing.T) {
content := []byte("service: storage\nprompt_text: hello\n")
_, err := ParsePromptYAML(content, "no-id.prompt.yaml")
if err == nil {
t.Fatal("expected error for missing ID")
}
}

func TestParsePromptYAMLInvalidField(t *testing.T) {
content := []byte("id: test\nunknown_field: bad\nprompt_text: hello\n")
_, err := ParsePromptYAML(content, "bad.prompt.yaml")
if err == nil {
t.Fatal("expected error for unknown field with KnownFields(true)")
}
}

func TestLoadPromptsYAML(t *testing.T) {
dir := t.TempDir()
subDir := filepath.Join(dir, "storage", "data-plane", "dotnet")
if err := os.MkdirAll(subDir, 0755); err != nil {
t.Fatalf("failed to create dir: %v", err)
}

if err := os.WriteFile(filepath.Join(subDir, "auth.prompt.yaml"), []byte(testYAMLPromptContent), 0644); err != nil {
t.Fatalf("failed to write file: %v", err)
}

prompts, err := LoadPrompts(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(prompts) != 1 {
t.Fatalf("expected 1 prompt, got %d", len(prompts))
}
if prompts[0].ID != "storage-auth-dotnet" {
t.Errorf("expected ID 'storage-auth-dotnet', got %q", prompts[0].ID)
}
}

func TestLoadPromptsMixed(t *testing.T) {
dir := t.TempDir()
if err := os.WriteFile(filepath.Join(dir, "md.prompt.md"), []byte(testPromptContent), 0644); err != nil {
t.Fatalf("failed to write md file: %v", err)
}
if err := os.WriteFile(filepath.Join(dir, "yaml.prompt.yml"), []byte(testYAMLPromptContent), 0644); err != nil {
t.Fatalf("failed to write yaml file: %v", err)
}

prompts, err := LoadPrompts(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(prompts) != 2 {
t.Fatalf("expected 2 prompts (md + yaml), got %d", len(prompts))
}
}

func TestScanNearMissesIgnoresYAMLPrompts(t *testing.T) {
dir := t.TempDir()
os.WriteFile(filepath.Join(dir, "auth.prompt.yaml"), []byte("id: x\nprompt_text: hello\n"), 0644)
misses := ScanNearMisses(dir)
if len(misses) != 0 {
t.Errorf("expected 0 near misses for .prompt.yaml file, got %d", len(misses))
}
}

