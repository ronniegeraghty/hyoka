// Package manifest provides functionality to generate a prompt manifest from the repository.
package manifest

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/ronniegeraghty/hyoka/internal/prompt"
	"gopkg.in/yaml.v3"
)

// PromptEntry represents a single prompt in the manifest.
type PromptEntry struct {
	ID          string   `yaml:"id"`
	Service     string   `yaml:"service"`
	Plane       string   `yaml:"plane"`
	Language    string   `yaml:"language"`
	Category    string   `yaml:"category"`
	Difficulty  string   `yaml:"difficulty"`
	Description string   `yaml:"description"`
	Path        string   `yaml:"path"`
	Created     string   `yaml:"created"`
	Author      string   `yaml:"author"`
	SDKPackage  string   `yaml:"sdk_package,omitempty"`
	DocURL      string   `yaml:"doc_url,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// Manifest represents the full manifest.yaml structure.
type Manifest struct {
	GeneratedAt string        `yaml:"generated_at"`
	PromptCount int           `yaml:"prompt_count"`
	Services    []string      `yaml:"services"`
	Languages   []string      `yaml:"languages"`
	Categories  []string      `yaml:"categories"`
	Prompts     []PromptEntry `yaml:"prompts"`
}

// Generate scans promptsDir for .prompt.md files and returns a Manifest.
// repoRoot is used to compute relative paths (typically the parent of promptsDir).
func Generate(promptsDir string) (*Manifest, error) {
	prompts, err := prompt.LoadPrompts(promptsDir)
	if err != nil {
		return nil, fmt.Errorf("loading prompts: %w", err)
	}

	// repoRoot is the parent of promptsDir so paths become "prompts/..."
	repoRoot := filepath.Dir(promptsDir)

	serviceSet := make(map[string]bool)
	languageSet := make(map[string]bool)
	categorySet := make(map[string]bool)

	var entries []PromptEntry
	for _, p := range prompts {
		relPath, err := filepath.Rel(repoRoot, p.FilePath)
		if err != nil {
			relPath = p.FilePath
		}
		relPath = filepath.ToSlash(relPath)

		entry := PromptEntry{
			ID:          p.ID,
			Service:     p.Service,
			Plane:       p.Plane,
			Language:    p.Language,
			Category:    p.Category,
			Difficulty:  p.Difficulty,
			Description: p.Description,
			Path:        relPath,
			Created:     p.Created,
			Author:      p.Author,
			SDKPackage:  p.SDKPackage,
			DocURL:      p.DocURL,
		}
		if len(p.Tags) > 0 {
			entry.Tags = p.Tags
		}
		entries = append(entries, entry)

		serviceSet[p.Service] = true
		languageSet[p.Language] = true
		categorySet[p.Category] = true
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})

	return &Manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		PromptCount: len(entries),
		Services:    sortedKeys(serviceSet),
		Languages:   sortedKeys(languageSet),
		Categories:  sortedKeys(categorySet),
		Prompts:     entries,
	}, nil
}

// Marshal serializes the manifest to YAML bytes.
func (m *Manifest) Marshal() ([]byte, error) {
	return yaml.Marshal(m)
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
