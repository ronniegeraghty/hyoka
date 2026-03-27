package manifest

import (
"os"
"path/filepath"
"testing"

"gopkg.in/yaml.v3"
)

const testPrompt1 = `---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: basic
description: "Authenticate to Azure Blob Storage"
sdk_package: Azure.Storage.Blobs
doc_url: https://example.com
tags:
  - blob
  - identity
created: "2024-01-15"
author: test
---

# Storage Auth

## Prompt

Write auth code.
`

const testPrompt2 = `---
id: key-vault-dp-python-crud
service: key-vault
plane: data-plane
language: python
category: crud
difficulty: intermediate
description: "CRUD operations on Key Vault secrets"
created: "2024-02-01"
author: test2
---

# Key Vault CRUD

## Prompt

Write CRUD code.
`

func setupTestDir(t *testing.T) string {
t.Helper()
dir := t.TempDir()
promptsDir := filepath.Join(dir, "prompts")

dir1 := filepath.Join(promptsDir, "storage", "data-plane", "dotnet")
dir2 := filepath.Join(promptsDir, "key-vault", "data-plane", "python")
if err := os.MkdirAll(dir1, 0755); err != nil {
t.Fatal(err)
}
if err := os.MkdirAll(dir2, 0755); err != nil {
t.Fatal(err)
}
if err := os.WriteFile(filepath.Join(dir1, "auth.prompt.md"), []byte(testPrompt1), 0644); err != nil {
t.Fatal(err)
}
if err := os.WriteFile(filepath.Join(dir2, "crud.prompt.md"), []byte(testPrompt2), 0644); err != nil {
t.Fatal(err)
}
return promptsDir
}

func TestGenerate(t *testing.T) {
promptsDir := setupTestDir(t)

m, err := Generate(promptsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if m.PromptCount != 2 {
t.Errorf("expected 2 prompts, got %d", m.PromptCount)
}
if m.GeneratedAt == "" {
t.Error("expected non-empty generated_at")
}

// Check sorted services
if len(m.Services) != 2 {
t.Fatalf("expected 2 services, got %d", len(m.Services))
}
if m.Services[0] != "key-vault" || m.Services[1] != "storage" {
t.Errorf("unexpected services: %v", m.Services)
}

// Check sorted languages
if len(m.Languages) != 2 {
t.Fatalf("expected 2 languages, got %d", len(m.Languages))
}
if m.Languages[0] != "dotnet" || m.Languages[1] != "python" {
t.Errorf("unexpected languages: %v", m.Languages)
}

// Check sorted categories
if len(m.Categories) != 2 {
t.Fatalf("expected 2 categories, got %d", len(m.Categories))
}
if m.Categories[0] != "authentication" || m.Categories[1] != "crud" {
t.Errorf("unexpected categories: %v", m.Categories)
}

// Entries are sorted by ID
if m.Prompts[0].ID != "key-vault-dp-python-crud" {
t.Errorf("expected first prompt to be key-vault-dp-python-crud, got %q", m.Prompts[0].ID)
}
if m.Prompts[1].ID != "storage-dp-dotnet-auth" {
t.Errorf("expected second prompt to be storage-dp-dotnet-auth, got %q", m.Prompts[1].ID)
}

// Check relative paths start with "prompts/"
for _, p := range m.Prompts {
if len(p.Path) < 8 || p.Path[:8] != "prompts/" {
t.Errorf("expected path to start with 'prompts/', got %q", p.Path)
}
}

// Check optional fields
if m.Prompts[1].SDKPackage != "Azure.Storage.Blobs" {
t.Errorf("expected sdk_package 'Azure.Storage.Blobs', got %q", m.Prompts[1].SDKPackage)
}
if m.Prompts[1].DocURL != "https://example.com" {
t.Errorf("expected doc_url, got %q", m.Prompts[1].DocURL)
}
if len(m.Prompts[1].Tags) != 2 {
t.Errorf("expected 2 tags, got %d", len(m.Prompts[1].Tags))
}
// Second prompt has no optional fields
if m.Prompts[0].SDKPackage != "" {
t.Errorf("expected empty sdk_package, got %q", m.Prompts[0].SDKPackage)
}
}

func TestMarshal(t *testing.T) {
promptsDir := setupTestDir(t)

m, err := Generate(promptsDir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

data, err := m.Marshal()
if err != nil {
t.Fatalf("marshal error: %v", err)
}

// Verify it round-trips as valid YAML
var parsed Manifest
if err := yaml.Unmarshal(data, &parsed); err != nil {
t.Fatalf("unmarshal error: %v", err)
}
if parsed.PromptCount != 2 {
t.Errorf("expected 2 prompts after round-trip, got %d", parsed.PromptCount)
}
}

func TestGenerateEmptyDir(t *testing.T) {
dir := t.TempDir()
m, err := Generate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if m.PromptCount != 0 {
t.Errorf("expected 0 prompts, got %d", m.PromptCount)
}
}

func TestGenerateNonexistentDir(t *testing.T) {
_, err := Generate("/nonexistent/path")
if err == nil {
t.Fatal("expected error for nonexistent directory")
}
}
