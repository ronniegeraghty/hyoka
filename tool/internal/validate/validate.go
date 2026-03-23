// Package validate provides functionality to validate prompt files against schema and rules.
package validate

import (
"fmt"
"sort"
"strings"

"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/prompt"
)

var validServices = map[string]bool{
"storage": true, "key-vault": true, "cosmos-db": true, "event-hubs": true,
"app-configuration": true, "purview": true, "digital-twins": true,
"identity": true, "resource-manager": true, "service-bus": true,
}

var validPlanes = map[string]bool{
"data-plane": true, "management-plane": true,
}

var validLanguages = map[string]bool{
"dotnet": true, "java": true, "js-ts": true, "python": true,
"go": true, "rust": true, "cpp": true,
}

var validCategories = map[string]bool{
"authentication": true, "pagination": true, "polling": true, "retries": true,
"error-handling": true, "crud": true, "batch": true, "streaming": true,
"auth": true, "provisioning": true,
}

var validDifficulties = map[string]bool{
"basic": true, "intermediate": true, "advanced": true,
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
"id": p.ID, "service": p.Service, "plane": p.Plane,
"language": p.Language, "category": p.Category, "difficulty": p.Difficulty,
"description": p.Description, "created": p.Created, "author": p.Author,
}
for field, val := range requiredFields {
if val == "" {
addErr(fmt.Sprintf("missing required field: %s", field))
}
}

// Enum validation
if p.Service != "" && !validServices[p.Service] {
addErr(fmt.Sprintf("invalid service %q; must be one of: %s", p.Service, joinKeys(validServices)))
}
if p.Plane != "" && !validPlanes[p.Plane] {
addErr(fmt.Sprintf("invalid plane %q; must be one of: %s", p.Plane, joinKeys(validPlanes)))
}
if p.Language != "" && !validLanguages[p.Language] {
addErr(fmt.Sprintf("invalid language %q; must be one of: %s", p.Language, joinKeys(validLanguages)))
}
if p.Category != "" && !validCategories[p.Category] {
addErr(fmt.Sprintf("invalid category %q; must be one of: %s", p.Category, joinKeys(validCategories)))
}
if p.Difficulty != "" && !validDifficulties[p.Difficulty] {
addErr(fmt.Sprintf("invalid difficulty %q; must be one of: %s", p.Difficulty, joinKeys(validDifficulties)))
}

// ID naming convention: {service}-{dp|mp}-{language}-
if p.Service != "" && p.Plane != "" && p.Language != "" {
abbrev := planeAbbrev[p.Plane]
if abbrev != "" {
expectedPrefix := fmt.Sprintf("%s-%s-%s-", p.Service, abbrev, p.Language)
if !strings.HasPrefix(p.ID, expectedPrefix) {
addErr(fmt.Sprintf("id %q must start with %q", p.ID, expectedPrefix))
}
}
}

// Must have ## Prompt section with content
if p.PromptText == "" {
addErr("missing or empty ## Prompt section")
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
