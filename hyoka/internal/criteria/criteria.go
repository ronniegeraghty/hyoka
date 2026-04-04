// Package criteria implements a tiered evaluation criteria system.
//
// Criteria YAML files use a "when" map to declare which prompts they apply to.
// All entries in "when" must match the corresponding prompt property
// (case-insensitive). An empty or absent "when" block matches every prompt.
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

// CriteriaSet is a collection of criteria with conditions for when they apply.
type CriteriaSet struct {
When     map[string]string `yaml:"when"`
Criteria []Criterion       `yaml:"criteria"`
Source   string            `yaml:"-"` // source file path
}

// Matches returns true if every entry in When matches the corresponding value
// in properties (case-insensitive). An empty When matches all prompts.
func (cs CriteriaSet) Matches(properties map[string]string) bool {
for k, v := range cs.When {
if !strings.EqualFold(properties[k], v) {
return false
}
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

// MatchingCriteria returns all criteria from sets whose When conditions match
// the given prompt properties.
func MatchingCriteria(sets []CriteriaSet, properties map[string]string) []Criterion {
var matched []Criterion
for _, s := range sets {
if s.Matches(properties) {
matched = append(matched, s.Criteria...)
}
}
return matched
}
