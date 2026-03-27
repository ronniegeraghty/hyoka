package validate

import (
"os"
"path/filepath"
"strings"
"testing"
)

const validPrompt = `---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: basic
description: "Authenticate to Azure Blob Storage"
created: "2024-01-15"
author: test
---

# Storage Auth

## Prompt

Write auth code for Azure Blob Storage.
`

const invalidServicePrompt = `---
id: unknown-dp-dotnet-test
service: unknown-service
plane: data-plane
language: dotnet
category: authentication
difficulty: basic
description: "Test"
created: "2024-01-15"
author: test
---

# Test

## Prompt

Some prompt text.
`

const missingFieldsPrompt = `---
id: storage-dp-dotnet-partial
service: storage
plane: data-plane
language: dotnet
---

# Partial

## Prompt

Some prompt text.
`

const badIDPrompt = `---
id: wrong-prefix-dotnet
service: storage
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: "Bad ID"
created: "2024-01-15"
author: test
---

# Bad ID

## Prompt

Some prompt text.
`

const noPromptSection = `---
id: storage-dp-dotnet-noprompt
service: storage
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: "No prompt section"
created: "2024-01-15"
author: test
---

# No Prompt

Just some text but no ## Prompt heading.
`

func writePromptFile(t *testing.T, dir, name, content string) {
t.Helper()
if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
t.Fatal(err)
}
}

func TestValidateAllValid(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "auth.prompt.md", validPrompt)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !result.OK() {
t.Errorf("expected validation to pass, got errors: %v", result.Errors)
}
if result.TotalFiles != 1 {
t.Errorf("expected 1 file, got %d", result.TotalFiles)
}
}

func TestValidateInvalidService(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "bad.prompt.md", invalidServicePrompt)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.OK() {
t.Fatal("expected validation errors")
}

found := false
for _, e := range result.Errors {
if strings.Contains(e.Message, "invalid service") {
found = true
}
}
if !found {
t.Errorf("expected 'invalid service' error, got: %v", result.Errors)
}
}

func TestValidateMissingFields(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "partial.prompt.md", missingFieldsPrompt)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.OK() {
t.Fatal("expected validation errors")
}

// Should report missing: category, difficulty, description, created, author
missingFields := []string{"category", "difficulty", "description", "created", "author"}
for _, field := range missingFields {
found := false
for _, e := range result.Errors {
if strings.Contains(e.Message, field) {
found = true
}
}
if !found {
t.Errorf("expected error about missing field %q", field)
}
}
}

func TestValidateBadID(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "badid.prompt.md", badIDPrompt)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.OK() {
t.Fatal("expected validation errors")
}

found := false
for _, e := range result.Errors {
if strings.Contains(e.Message, "must start with") {
found = true
}
}
if !found {
t.Errorf("expected ID naming error, got: %v", result.Errors)
}
}

func TestValidateNoPromptSection(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "noprompt.prompt.md", noPromptSection)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.OK() {
t.Fatal("expected validation errors")
}

found := false
for _, e := range result.Errors {
if strings.Contains(e.Message, "Prompt section") {
found = true
}
}
if !found {
t.Errorf("expected 'Prompt section' error, got: %v", result.Errors)
}
}

func TestValidateMultipleFiles(t *testing.T) {
dir := t.TempDir()
writePromptFile(t, dir, "good.prompt.md", validPrompt)
writePromptFile(t, dir, "bad.prompt.md", invalidServicePrompt)

result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.TotalFiles != 2 {
t.Errorf("expected 2 files, got %d", result.TotalFiles)
}
if result.OK() {
t.Fatal("expected validation errors from bad file")
}
}

func TestValidateEmptyDir(t *testing.T) {
dir := t.TempDir()
result, err := Validate(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !result.OK() {
t.Errorf("expected no errors for empty dir, got: %v", result.Errors)
}
if result.TotalFiles != 0 {
t.Errorf("expected 0 files, got %d", result.TotalFiles)
}
}

func TestFormatResultSuccess(t *testing.T) {
r := &Result{TotalFiles: 5}
out := FormatResult(r)
if !strings.Contains(out, "✓") || !strings.Contains(out, "5") {
t.Errorf("unexpected output: %s", out)
}
}

func TestFormatResultFailure(t *testing.T) {
r := &Result{
TotalFiles: 1,
Errors: []ValidationError{
{File: "test.prompt.md", Message: "missing field"},
},
}
out := FormatResult(r)
if !strings.Contains(out, "✗") || !strings.Contains(out, "1 error") {
t.Errorf("unexpected output: %s", out)
}
}

func TestValidateNonexistentDir(t *testing.T) {
_, err := Validate("/nonexistent/path")
if err == nil {
t.Fatal("expected error for nonexistent directory")
}
}
