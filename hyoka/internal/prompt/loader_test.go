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
