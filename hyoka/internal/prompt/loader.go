// Package prompt provides functionality to load and parse prompt files.
package prompt

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// LoadPrompts walks the directory tree at root and loads all .prompt.md files.
// Returns an error if the directory contains zero valid prompts, along with
// near-miss suggestions for files that almost match the naming pattern.
func LoadPrompts(root string) ([]*Prompt, error) {
	slog.Debug("Scanning for prompts", "root", root)
	var prompts []*Prompt

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".prompt.md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		p, err := ParsePromptFile(data, path)
		if err != nil {
			return err
		}

		prompts = append(prompts, p)
		slog.Debug("Loaded prompt", "id", p.ID, "path", path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking prompt directory %s: %w", root, err)
	}

	if len(prompts) == 0 {
		nearMisses := ScanNearMisses(root)
		msg := fmt.Sprintf("no prompts found in %s", root)
		if len(nearMisses) > 0 {
			msg += "\n\nDid you mean one of these?"
			for _, nm := range nearMisses {
				suggestion := suggestFix(nm)
				if suggestion != "" {
					msg += fmt.Sprintf("\n  %s → %s", nm, suggestion)
				} else {
					msg += fmt.Sprintf("\n  %s (does not match *.prompt.md pattern)", nm)
				}
			}
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return prompts, nil
}

// ScanNearMisses finds files in dir that look like prompts but don't match
// the *.prompt.md naming convention. It detects patterns such as:
//   - *-prompt.md (hyphenated instead of dotted)
//   - *.prompt.txt (wrong extension)
//   - *.md files containing YAML frontmatter (--- delimiters)
func ScanNearMisses(dir string) []string {
	var nearMisses []string
	seen := make(map[string]bool)

	if walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := info.Name()

		// Skip files that already match the correct pattern
		if strings.HasSuffix(name, ".prompt.md") {
			return nil
		}

		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			slog.Warn("Failed to compute relative path", "dir", dir, "path", path, "error", relErr)
		}
		if rel == "" {
			rel = name
		}

		// Pattern: *-prompt.md (hyphen instead of dot before "prompt")
		if strings.HasSuffix(name, "-prompt.md") {
			if !seen[rel] {
				seen[rel] = true
				nearMisses = append(nearMisses, rel)
			}
			return nil
		}

		// Pattern: *.prompt.txt or *.prompt.* (right base, wrong extension)
		if strings.Contains(name, ".prompt.") && !strings.HasSuffix(name, ".prompt.md") {
			if !seen[rel] {
				seen[rel] = true
				nearMisses = append(nearMisses, rel)
			}
			return nil
		}

		// Pattern: *.md files that contain YAML frontmatter
		if strings.HasSuffix(name, ".md") && info.Size() > 0 && info.Size() < 1<<20 {
			data, readErr := os.ReadFile(path)
			if readErr == nil && strings.HasPrefix(string(data), "---") {
				if !seen[rel] {
					seen[rel] = true
					nearMisses = append(nearMisses, rel)
				}
			}
			return nil
		}

		return nil
	}); walkErr != nil {
		slog.Warn("Failed to walk directory for near-miss prompts", "dir", dir, "error", walkErr)
	}

	return nearMisses
}

// suggestFix returns a corrected filename for a near-miss, or "" if no fix is obvious.
func suggestFix(name string) string {
	base := filepath.Base(name)
	dir := filepath.Dir(name)

	// *-prompt.md → *.prompt.md
	if strings.HasSuffix(base, "-prompt.md") {
		fixed := strings.TrimSuffix(base, "-prompt.md") + ".prompt.md"
		if dir == "." {
			return fixed
		}
		return filepath.ToSlash(filepath.Join(dir, fixed))
	}

	// *.prompt.txt or *.prompt.* → *.prompt.md
	if idx := strings.Index(base, ".prompt."); idx >= 0 {
		fixed := base[:idx] + ".prompt.md"
		if dir == "." {
			return fixed
		}
		return filepath.ToSlash(filepath.Join(dir, fixed))
	}

	return ""
}

// FilterPrompts returns only the prompts matching the given filter.
func FilterPrompts(prompts []*Prompt, f Filter) []*Prompt {
	var result []*Prompt
	for _, p := range prompts {
		if p.Matches(f) {
			result = append(result, p)
		}
	}
	return result
}
