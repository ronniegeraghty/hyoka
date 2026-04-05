package config

import "fmt"

// ToolEntry represents a tool with optional property-based conditions.
//
// Entries with neither When nor ExcludeWhen are unconditional (always included).
// When all key-value pairs in When match the prompt, the tool is available.
// When all key-value pairs in ExcludeWhen match the prompt, the tool is excluded.
// Setting both When and ExcludeWhen on the same entry is a validation error.
type ToolEntry struct {
	Name        string            `yaml:"name"                   json:"name"`
	When        map[string]string `yaml:"when,omitempty"         json:"when,omitempty"`
	ExcludeWhen map[string]string `yaml:"exclude_when,omitempty" json:"exclude_when,omitempty"`
}

// PromptProperties holds the subset of prompt metadata used for tool filtering.
type PromptProperties struct {
	Language   string
	Service    string
	Plane      string
	Category   string
	Difficulty string
}

// validPropertyKeys is the set of prompt property names recognized in
// When/ExcludeWhen conditions.
var validPropertyKeys = map[string]bool{
	"language":   true,
	"service":    true,
	"plane":      true,
	"category":   true,
	"difficulty": true,
}

// propertyValue returns the prompt property value for a given key.
func (p PromptProperties) propertyValue(key string) string {
	switch key {
	case "language":
		return p.Language
	case "service":
		return p.Service
	case "plane":
		return p.Plane
	case "category":
		return p.Category
	case "difficulty":
		return p.Difficulty
	default:
		return ""
	}
}

// matchesAll returns true if every key-value pair in conditions matches the
// corresponding prompt property. An empty or nil map always matches.
func matchesAll(conditions map[string]string, props PromptProperties) bool {
	for key, want := range conditions {
		if props.propertyValue(key) != want {
			return false
		}
	}
	return true
}

// ResolveTools evaluates conditional tool entries against prompt properties and
// merges the result with legacy flat lists. Returns the final available and
// excluded tool slices.
//
// Resolution order:
//  1. Start with copies of the legacy available_tools and excluded_tools.
//  2. Evaluate each ToolEntry: unconditional entries and entries whose When
//     conditions match are appended to available; entries whose ExcludeWhen
//     conditions match are appended to excluded.
//  3. Deduplicate both lists.
//  4. If a tool appears in both, excluded wins.
//  5. Nil semantics are preserved: a nil available list means "all defaults".
func ResolveTools(gen *GeneratorConfig, props PromptProperties) (available, excluded []string) {
	if gen == nil {
		return nil, nil
	}

	// No conditional tools defined → return legacy lists as-is.
	if len(gen.Tools) == 0 {
		return gen.AvailableTools, gen.ExcludedTools
	}

	// Start with copies of legacy lists.
	if len(gen.AvailableTools) > 0 {
		available = make([]string, len(gen.AvailableTools))
		copy(available, gen.AvailableTools)
	}
	if len(gen.ExcludedTools) > 0 {
		excluded = make([]string, len(gen.ExcludedTools))
		copy(excluded, gen.ExcludedTools)
	}

	for _, entry := range gen.Tools {
		switch {
		case len(entry.ExcludeWhen) > 0:
			if matchesAll(entry.ExcludeWhen, props) {
				excluded = append(excluded, entry.Name)
			}
		case len(entry.When) > 0:
			if matchesAll(entry.When, props) {
				available = append(available, entry.Name)
			}
		default:
			// Unconditional entry.
			available = append(available, entry.Name)
		}
	}

	available = dedup(available)
	excluded = dedup(excluded)

	// Excluded wins: remove any tool that appears in both lists.
	if len(available) > 0 && len(excluded) > 0 {
		excludeSet := make(map[string]bool, len(excluded))
		for _, name := range excluded {
			excludeSet[name] = true
		}
		filtered := available[:0]
		for _, name := range available {
			if !excludeSet[name] {
				filtered = append(filtered, name)
			}
		}
		available = filtered
	}

	return available, excluded
}

// dedup removes duplicate strings while preserving order. Returns nil for nil input.
func dedup(ss []string) []string {
	if ss == nil {
		return nil
	}
	seen := make(map[string]bool, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// validateToolEntry checks that a ToolEntry has valid fields.
func validateToolEntry(entry ToolEntry, configName string, idx int) error {
	if entry.Name == "" {
		return fmt.Errorf("config %q: tools[%d] missing name", configName, idx)
	}
	if len(entry.When) > 0 && len(entry.ExcludeWhen) > 0 {
		return fmt.Errorf("config %q: tools[%d] (%s) has both when and exclude_when", configName, idx, entry.Name)
	}
	for key, val := range entry.When {
		if !validPropertyKeys[key] {
			return fmt.Errorf("config %q: tools[%d] (%s) when: unrecognized property %q", configName, idx, entry.Name, key)
		}
		if val == "" {
			return fmt.Errorf("config %q: tools[%d] (%s) when[%s]: value must not be empty", configName, idx, entry.Name, key)
		}
	}
	for key, val := range entry.ExcludeWhen {
		if !validPropertyKeys[key] {
			return fmt.Errorf("config %q: tools[%d] (%s) exclude_when: unrecognized property %q", configName, idx, entry.Name, key)
		}
		if val == "" {
			return fmt.Errorf("config %q: tools[%d] (%s) exclude_when[%s]: value must not be empty", configName, idx, entry.Name, key)
		}
	}
	return nil
}
