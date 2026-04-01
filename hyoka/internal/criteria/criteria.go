// Package criteria implements a tiered evaluation criteria system.
//
// Three tiers of criteria are merged at eval time:
//   - Tier 1: General criteria from rubric.md (always applied)
//   - Tier 2: Attribute-matched criteria from YAML files (applied when prompt metadata matches)
//   - Tier 3: Prompt-specific criteria from the prompt's ## Evaluation Criteria section
package criteria

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Criterion defines a single evaluation criterion.
type Criterion struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

// MatchCondition defines when a criteria set should apply.
// All non-empty fields must match the prompt's metadata.
type MatchCondition struct {
	Language string `yaml:"language,omitempty"`
	Service  string `yaml:"service,omitempty"`
	Plane    string `yaml:"plane,omitempty"`
	Category string `yaml:"category,omitempty"`
	SDK      string `yaml:"sdk,omitempty"`
}

// CriteriaSet is a collection of criteria with conditions for when they apply.
type CriteriaSet struct {
	Match    MatchCondition `yaml:"match"`
	Criteria []Criterion    `yaml:"criteria"`
	Source   string         `yaml:"-"` // source file path
}

// PromptAttrs holds prompt metadata used for criteria matching.
type PromptAttrs struct {
	Language string
	Service  string
	Plane    string
	Category string
	SDK      string
}

// Matches returns true if all non-empty fields in the condition match the prompt attrs.
func (m MatchCondition) Matches(attrs PromptAttrs) bool {
	if m.Language != "" && !strings.EqualFold(m.Language, attrs.Language) {
		return false
	}
	if m.Service != "" && !strings.EqualFold(m.Service, attrs.Service) {
		return false
	}
	if m.Plane != "" && !strings.EqualFold(m.Plane, attrs.Plane) {
		return false
	}
	if m.Category != "" && !strings.EqualFold(m.Category, attrs.Category) {
		return false
	}
	if m.SDK != "" && !strings.EqualFold(m.SDK, attrs.SDK) {
		return false
	}
	return true
}

// LoadDir loads all criteria YAML files from a directory tree.
func LoadDir(dir string) ([]CriteriaSet, error) {
	var sets []CriteriaSet

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		cs, err := loadFile(path)
		if err != nil {
			slog.Warn("Skipping invalid criteria file", "path", path, "error", err)
			return nil
		}
		cs.Source = path
		sets = append(sets, *cs)
		slog.Debug("Loaded criteria set", "path", path, "criteria_count", len(cs.Criteria))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking criteria directory %s: %w", dir, err)
	}
	return sets, nil
}

func loadFile(path string) (*CriteriaSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cs CriteriaSet
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cs); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(cs.Criteria) == 0 {
		return nil, fmt.Errorf("%s: no criteria defined", path)
	}
	return &cs, nil
}

// MatchingCriteria returns all criteria from sets that match the given prompt attrs.
func MatchingCriteria(sets []CriteriaSet, attrs PromptAttrs) []Criterion {
	var matched []Criterion
	for _, s := range sets {
		if s.Match.Matches(attrs) {
			matched = append(matched, s.Criteria...)
		}
	}
	return matched
}

// FormatCriteria formats a list of criteria as a text block suitable for
// injection into a review prompt.
func FormatCriteria(criteria []Criterion) string {
	if len(criteria) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range criteria {
		fmt.Fprintf(&b, "%d. **%s**", i+1, c.Name)
		if c.Description != "" {
			fmt.Fprintf(&b, " — %s", strings.TrimSpace(c.Description))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// MergeCriteria combines Tier 2 (attribute-matched) criteria text with Tier 3
// (prompt-specific) criteria text. Returns the merged string suitable for
// passing to the reviewer.
func MergeCriteria(tier2 []Criterion, tier3Text string) string {
	parts := make([]string, 0, 2)

	formatted := FormatCriteria(tier2)
	if formatted != "" {
		parts = append(parts, "### Attribute-Matched Criteria\n\n"+formatted)
	}

	tier3Text = strings.TrimSpace(tier3Text)
	if tier3Text != "" {
		parts = append(parts, "### Prompt-Specific Criteria\n\n"+tier3Text)
	}

	return strings.Join(parts, "\n")
}
