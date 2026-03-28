package criteria

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadCriteriaSets scans a directory tree for .yaml/.yml files and parses each
// as a CriteriaSet. Files that fail to parse are logged as warnings and skipped.
func LoadCriteriaSets(root string) ([]CriteriaSet, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("Criteria directory does not exist, skipping", "path", root)
			return nil, nil
		}
		return nil, fmt.Errorf("stat criteria directory %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("criteria path is not a directory: %s", root)
	}

	var sets []CriteriaSet
	err = filepath.Walk(root, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(fi.Name()))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			slog.Warn("Failed to read criteria file", "path", path, "error", readErr)
			return nil
		}

		var cs CriteriaSet
		if parseErr := yaml.Unmarshal(data, &cs); parseErr != nil {
			slog.Warn("Failed to parse criteria file", "path", path, "error", parseErr)
			return nil
		}

		if len(cs.Criteria) == 0 {
			slog.Debug("Criteria file has no criteria entries, skipping", "path", path)
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		if rel == "" {
			rel = path
		}
		cs.Source = rel

		slog.Debug("Loaded criteria set", "path", rel, "match", cs.Match, "criteria_count", len(cs.Criteria))
		sets = append(sets, cs)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking criteria directory %s: %w", root, err)
	}

	slog.Info("Criteria sets loaded", "count", len(sets), "root", root)
	return sets, nil
}

// MatchCriteria returns all criteria from the loaded sets whose match rules
// are satisfied by the given prompt attributes. Results are returned in load order.
func MatchCriteria(sets []CriteriaSet, attrs PromptAttributes) []Criterion {
	var matched []Criterion
	for _, s := range sets {
		if s.Match.Matches(attrs) {
			slog.Debug("Criteria set matched", "source", s.Source, "criteria_count", len(s.Criteria))
			matched = append(matched, s.Criteria...)
		}
	}
	return matched
}

// FormatCriteria renders a list of criteria as a numbered markdown list
// suitable for injection into the review prompt.
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
