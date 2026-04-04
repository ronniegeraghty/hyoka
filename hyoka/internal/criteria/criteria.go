// Package criteria implements a grader-config-based evaluation criteria system.
//
// Criteria YAML files define grader configs with:
//   - when: map[string]string conditions (all must match for the config to apply)
//   - graders: weighted evaluation rubrics with prompts
//
// At eval time, matching grader configs are collected and merged with any
// prompt-specific criteria to form the final evaluation rubric.
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

// GraderEntry defines a single grader with its evaluation prompt and weight.
type GraderEntry struct {
	Name   string  `yaml:"name" json:"name"`
	Weight float64 `yaml:"weight" json:"weight"`
	Prompt string  `yaml:"prompt" json:"prompt"`
}

// GraderConfig is a collection of graders with conditions for when they apply.
// When the When map is empty, the config is unconditionally included.
// When the When map has entries, all key-value pairs must match the prompt's
// properties for the graders to be included.
type GraderConfig struct {
	When    map[string]string `yaml:"when,omitempty" json:"when,omitempty"`
	Graders []GraderEntry     `yaml:"graders" json:"graders"`
	Source  string            `yaml:"-" json:"-"`
}

// matchesWhen returns true when every key-value pair in when matches the
// properties map (case-insensitive values). An empty when map always matches.
func matchesWhen(when map[string]string, props map[string]string) bool {
	for k, v := range when {
		if !strings.EqualFold(props[k], v) {
			return false
		}
	}
	return true
}

// LoadDir loads all grader config YAML files from a directory tree.
func LoadDir(dir string) ([]GraderConfig, error) {
	var configs []GraderConfig

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

		gc, err := loadFile(path)
		if err != nil {
			slog.Warn("Skipping invalid grader config file", "path", path, "error", err)
			return nil
		}
		gc.Source = path
		configs = append(configs, *gc)
		slog.Debug("Loaded grader config", "path", path, "grader_count", len(gc.Graders))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking criteria directory %s: %w", dir, err)
	}
	return configs, nil
}

func loadFile(path string) (*GraderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gc GraderConfig
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&gc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(gc.Graders) == 0 {
		return nil, fmt.Errorf("%s: no graders defined", path)
	}
	return &gc, nil
}

// MatchingGraders returns all grader entries from configs whose when-conditions
// match the given prompt properties.
func MatchingGraders(configs []GraderConfig, props map[string]string) []GraderEntry {
	var matched []GraderEntry
	for _, gc := range configs {
		if matchesWhen(gc.When, props) {
			matched = append(matched, gc.Graders...)
		}
	}
	return matched
}

// FormatGraders formats a list of grader entries as a text block suitable for
// injection into a review prompt.
func FormatGraders(graders []GraderEntry) string {
	if len(graders) == 0 {
		return ""
	}
	var b strings.Builder
	for i, g := range graders {
		fmt.Fprintf(&b, "%d. **%s**", i+1, g.Name)
		if g.Prompt != "" {
			fmt.Fprintf(&b, " — %s", strings.TrimSpace(g.Prompt))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// MergeCriteria combines attribute-matched grader entries with prompt-specific
// criteria text. Returns the merged string suitable for passing to the reviewer.
func MergeCriteria(graders []GraderEntry, promptCriteria string) string {
	parts := make([]string, 0, 2)

	formatted := FormatGraders(graders)
	if formatted != "" {
		parts = append(parts, "### Attribute-Matched Criteria\n\n"+formatted)
	}

	promptCriteria = strings.TrimSpace(promptCriteria)
	if promptCriteria != "" {
		parts = append(parts, "### Prompt-Specific Criteria\n\n"+promptCriteria)
	}

	return strings.Join(parts, "\n")
}
