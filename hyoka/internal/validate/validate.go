// Package validate provides functionality to validate prompt files against schema and rules.
package validate

import (
"fmt"
"os"
"path/filepath"
"sort"
"strings"

"github.com/ronniegeraghty/hyoka/internal/prompt"
)

// ValidServices is the canonical list of valid service values.
var ValidServices = []string{
"storage", "key-vault", "cosmos-db", "event-hubs",
"app-configuration", "purview", "digital-twins",
"identity", "resource-manager", "service-bus",
}

// ValidPlanes is the canonical list of valid plane values.
var ValidPlanes = []string{"data-plane", "management-plane"}

// ValidLanguages is the canonical list of valid language values.
var ValidLanguages = []string{"dotnet", "java", "js-ts", "python", "go", "rust", "cpp"}

// ValidCategories is the canonical list of valid category values.
var ValidCategories = []string{
"authentication", "pagination", "polling", "retries",
"error-handling", "crud", "batch", "streaming", "auth", "provisioning",
}

// ValidDifficulties is the canonical list of valid difficulty values.
var ValidDifficulties = []string{"basic", "intermediate", "advanced"}

var validServicesMap = toSet(ValidServices)
var validPlanesMap = toSet(ValidPlanes)
var validLanguagesMap = toSet(ValidLanguages)
var validCategoriesMap = toSet(ValidCategories)
var validDifficultiesMap = toSet(ValidDifficulties)

func toSet(ss []string) map[string]bool {
m := make(map[string]bool, len(ss))
for _, s := range ss {
m[s] = true
}
return m
}

// planeAbbrev maps plane values to their ID prefix abbreviation.
var planeAbbrev = map[string]string{
"data-plane":       "dp",
"management-plane": "mp",
}

// ValidationError represents a single validation error for a file.
type ValidationError struct {
File    string
Message string
}

func (e ValidationError) String() string {
return fmt.Sprintf("%s: %s", e.File, e.Message)
}

// Result holds the complete validation results.
type Result struct {
TotalFiles int
Errors     []ValidationError
}

// OK returns true if there are no validation errors.
func (r *Result) OK() bool {
return len(r.Errors) == 0
}

// Validate loads all prompt files from promptsDir and validates them.
func Validate(promptsDir string) (*Result, error) {
prompts, err := prompt.LoadPrompts(promptsDir)
if err != nil {
return nil, fmt.Errorf("loading prompts: %w", err)
}

result := &Result{TotalFiles: len(prompts)}

for _, p := range prompts {
errs := validatePrompt(p)
result.Errors = append(result.Errors, errs...)
}

return result, nil
}

func validatePrompt(p *prompt.Prompt) []ValidationError {
var errs []ValidationError
addErr := func(msg string) {
errs = append(errs, ValidationError{File: p.FilePath, Message: msg})
}

// Required fields (id is already enforced by parser, but check anyway)
requiredFields := map[string]string{
"id": p.ID, "service": p.Service(), "plane": p.Plane(),
"language": p.Language(), "category": p.Category(), "difficulty": p.Difficulty(),
"description": p.Description(), "created": p.Created(), "author": p.Author(),
}
for field, val := range requiredFields {
if val == "" {
addErr(fmt.Sprintf("missing required field: %s", field))
}
}

// Enum validation
if p.Service() != "" && !validServicesMap[p.Service()] {
addErr(fmt.Sprintf("invalid service %q; must be one of: %s", p.Service(), joinKeys(validServicesMap)))
}
if p.Plane() != "" && !validPlanesMap[p.Plane()] {
addErr(fmt.Sprintf("invalid plane %q; must be one of: %s", p.Plane(), joinKeys(validPlanesMap)))
}
if p.Language() != "" && !validLanguagesMap[p.Language()] {
addErr(fmt.Sprintf("invalid language %q; must be one of: %s", p.Language(), joinKeys(validLanguagesMap)))
}
if p.Category() != "" && !validCategoriesMap[p.Category()] {
addErr(fmt.Sprintf("invalid category %q; must be one of: %s", p.Category(), joinKeys(validCategoriesMap)))
}
if p.Difficulty() != "" && !validDifficultiesMap[p.Difficulty()] {
addErr(fmt.Sprintf("invalid difficulty %q; must be one of: %s", p.Difficulty(), joinKeys(validDifficultiesMap)))
}

// ID naming convention: {service}-{dp|mp}-{language}-
if p.Service() != "" && p.Plane() != "" && p.Language() != "" {
abbrev := planeAbbrev[p.Plane()]
if abbrev != "" {
expectedPrefix := fmt.Sprintf("%s-%s-%s-", p.Service(), abbrev, p.Language())
if !strings.HasPrefix(p.ID, expectedPrefix) {
addErr(fmt.Sprintf("id %q must start with %q", p.ID, expectedPrefix))
}
}
}

// Must have ## Prompt section with content
if p.PromptText == "" {
addErr("missing or empty ## Prompt section")
}

// Starter project path validation
if p.StarterProject != "" {
starterDir := p.StarterProject
if !filepath.IsAbs(starterDir) && p.FilePath != "" {
starterDir = filepath.Join(filepath.Dir(p.FilePath), starterDir)
}
info, statErr := os.Stat(starterDir)
if statErr != nil {
addErr(fmt.Sprintf("starter_project %q: path does not exist", p.StarterProject))
} else if !info.IsDir() {
addErr(fmt.Sprintf("starter_project %q: not a directory", p.StarterProject))
} else {
entries, readErr := os.ReadDir(starterDir)
if readErr != nil {
addErr(fmt.Sprintf("starter_project %q: cannot read directory", p.StarterProject))
} else {
hasFiles := false
for _, e := range entries {
if !strings.HasPrefix(e.Name(), ".") {
hasFiles = true
break
}
}
if !hasFiles {
addErr(fmt.Sprintf("starter_project %q: directory is empty", p.StarterProject))
}
}
}
}

return errs
}

// FormatResult returns a human-readable string for the validation result.
func FormatResult(r *Result) string {
if r.OK() {
return fmt.Sprintf("✓ All %d prompt(s) are valid", r.TotalFiles)
}

var b strings.Builder
fmt.Fprintf(&b, "Validation failed with %d error(s):\n", len(r.Errors))
for _, e := range r.Errors {
fmt.Fprintf(&b, "  ✗ %s\n", e.String())
}
return b.String()
}

func joinKeys(m map[string]bool) string {
keys := make([]string, 0, len(m))
for k := range m {
keys = append(keys, k)
}
// Sort for deterministic output
sort.Strings(keys)
return strings.Join(keys, ", ")
}
