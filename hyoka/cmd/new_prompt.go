package cmd

import (
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/ronniegeraghty/hyoka/internal/validate"
"github.com/spf13/cobra"
)

var validServices = validate.ValidServices
var validLanguages = validate.ValidLanguages
var validPlanes = validate.ValidPlanes
var validCategories = validate.ValidCategories
var validDifficulties = validate.ValidDifficulties

func newPromptCmd() *cobra.Command {
return &cobra.Command{
Use:   "new-prompt",
Short: "Scaffold a new prompt file interactively",
Long:  "Asks for service, language, plane, category, and difficulty, then generates a prompt file with populated frontmatter at the correct path.",
RunE: func(cmd *cobra.Command, args []string) error {
promptsDir := resolvePromptsDir(cmd)

service := askChoice("Service", validServices)
plane := askChoice("Plane", validPlanes)
language := askChoice("Language", validLanguages)
category := askChoice("Category", validCategories)
difficulty := askChoice("Difficulty", validDifficulties)
description := askFreeText("Description (what this prompt tests)")

// Build the prompt ID
planeAbbrev := "dp"
if plane == "management-plane" {
planeAbbrev = "mp"
}

// Ask for a slug to make the ID unique
slug := askFreeText("Short slug for filename (e.g. 'list-blobs')")
slug = strings.ReplaceAll(strings.TrimSpace(slug), " ", "-")
slug = strings.ToLower(slug)

id := fmt.Sprintf("%s-%s-%s-%s", service, planeAbbrev, language, slug)

dir := filepath.Join(promptsDir, service, plane, language)
if err := os.MkdirAll(dir, 0755); err != nil {
return fmt.Errorf("creating directory: %w", err)
}

filename := slug + ".prompt.md"
filePath := filepath.Join(dir, filename)

if _, err := os.Stat(filePath); err == nil {
return fmt.Errorf("file already exists: %s", filePath)
}

today := time.Now().Format("2006-01-02")

content := fmt.Sprintf("---\nid: %s\nservice: %s\nplane: %s\nlanguage: %s\ncategory: %s\ndifficulty: %s\ndescription: >\n  %s\nsdk_package: \"\"\ndoc_url: \"\"\ntags: []\ncreated: %s\nauthor: \"\"\n---\n\n# TODO: Title \u2014 %s (%s)\n\n## Prompt\n\nTODO: Write your prompt here.\n\n## Expected Coverage\n\nThe generated code should demonstrate:\n- TODO: List key aspects to test\n\n## Context\n\nTODO: Why this prompt matters.\n",
id, service, plane, language, category, difficulty, description, today, service, language)

if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
return fmt.Errorf("writing prompt file: %w", err)
}

fmt.Printf("\n\u2705 Created prompt file: %s\n", filePath)
fmt.Printf("   Prompt ID: %s\n", id)
fmt.Println("\nNext steps:")
fmt.Println("  1. Edit the file to add your prompt text")
fmt.Println("  2. Run: go run ./hyoka validate")
return nil
},
}
}

func askChoice(label string, options []string) string {
fmt.Printf("\n%s:\n", label)
for i, opt := range options {
fmt.Printf("  %d) %s\n", i+1, opt)
}
for {
fmt.Printf("Choose [1-%d]: ", len(options))
var choice int
_, err := fmt.Scanln(&choice)
if err == nil && choice >= 1 && choice <= len(options) {
return options[choice-1]
}
fmt.Println("Invalid choice, try again.")
}
}

func askFreeText(label string) string {
fmt.Printf("\n%s: ", label)
var input string
fmt.Scanln(&input)
return strings.TrimSpace(input)
}
