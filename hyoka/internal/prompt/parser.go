package prompt

import (
"bytes"
"fmt"
"regexp"
"strings"

"gopkg.in/yaml.v3"
)

var promptSectionRe = regexp.MustCompile(`(?m)^## Prompt\s*\n`)
var evaluationCriteriaRe = regexp.MustCompile(`(?m)^## Evaluation Criteria\s*\n`)

// rawFrontmatter is an intermediate struct used to decode frontmatter.
// It accepts both the old flat format and the new nested properties format.
type rawFrontmatter struct {
ID              string            `yaml:"id"`
Tags            []string          `yaml:"tags"`
ProjectContext  map[string]string `yaml:"project_context"`
StarterProject  string            `yaml:"starter_project"`
ReferenceAnswer string            `yaml:"reference_answer"`
Timeout         int               `yaml:"timeout"`
ExpectedPkgs    []string          `yaml:"expected_packages"`
ExpectedTools   []string          `yaml:"expected_tools"`

// New nested format
Properties map[string]string `yaml:"properties"`

// Old flat format fields — populated only when properties: is absent
Service     string `yaml:"service"`
Plane       string `yaml:"plane"`
Language    string `yaml:"language"`
Category    string `yaml:"category"`
Difficulty  string `yaml:"difficulty"`
Description string `yaml:"description"`
SDKPackage  string `yaml:"sdk_package"`
DocURL      string `yaml:"doc_url"`
Created     string `yaml:"created"`
Author      string `yaml:"author"`

// YAML-only prompt fields
PromptTextField         string `yaml:"prompt_text"`
EvaluationCriteriaField string `yaml:"evaluation_criteria"`
}

// rawToPrompt converts a decoded rawFrontmatter into a Prompt,
// populating Properties from flat fields when the nested format is absent.
func rawToPrompt(raw *rawFrontmatter) *Prompt {
p := &Prompt{
ID:              raw.ID,
Tags:            raw.Tags,
ProjectContext:  raw.ProjectContext,
StarterProject:  raw.StarterProject,
ReferenceAnswer: raw.ReferenceAnswer,
Timeout:         raw.Timeout,
ExpectedPkgs:    raw.ExpectedPkgs,
ExpectedTools:   raw.ExpectedTools,
}

if raw.Properties != nil {
p.Properties = raw.Properties
} else {
p.Properties = make(map[string]string)
setIfNonEmpty := func(k, v string) {
if v != "" {
p.Properties[k] = v
}
}
setIfNonEmpty("service", raw.Service)
setIfNonEmpty("plane", raw.Plane)
setIfNonEmpty("language", raw.Language)
setIfNonEmpty("category", raw.Category)
setIfNonEmpty("difficulty", raw.Difficulty)
setIfNonEmpty("description", raw.Description)
setIfNonEmpty("sdk_package", raw.SDKPackage)
setIfNonEmpty("doc_url", raw.DocURL)
setIfNonEmpty("created", raw.Created)
setIfNonEmpty("author", raw.Author)
}
return p
}

// ParsePromptFile parses a .prompt.md file's content into a Prompt struct.
// It supports both the new nested properties format and the old flat format.
// For .prompt.yaml/.prompt.yml files, use ParsePromptYAML instead.
func ParsePromptFile(content []byte, filePath string) (*Prompt, error) {
text := string(content)

if !strings.HasPrefix(text, "---") {
return nil, fmt.Errorf("file does not start with frontmatter delimiter: %s", filePath)
}
parts := strings.SplitN(text[3:], "---", 2)
if len(parts) < 2 {
return nil, fmt.Errorf("missing closing frontmatter delimiter: %s", filePath)
}
frontmatter := strings.TrimSpace(parts[0])
body := parts[1]

var raw rawFrontmatter
dec := yaml.NewDecoder(bytes.NewReader([]byte(frontmatter)))
if err := dec.Decode(&raw); err != nil {
return nil, fmt.Errorf("parsing frontmatter in %s: %w", filePath, err)
}

p := rawToPrompt(&raw)

loc := promptSectionRe.FindStringIndex(body)
if loc != nil {
promptBody := body[loc[1]:]
nextHeading := regexp.MustCompile(`(?m)^## `)
nextLoc := nextHeading.FindStringIndex(promptBody)
if nextLoc != nil {
promptBody = promptBody[:nextLoc[0]]
}
p.PromptText = strings.TrimSpace(promptBody)
}

// Extract ## Evaluation Criteria section
covLoc := evaluationCriteriaRe.FindStringIndex(body)
if covLoc != nil {
covBody := body[covLoc[1]:]
nextHeading := regexp.MustCompile(`(?m)^## `)
nextLoc := nextHeading.FindStringIndex(covBody)
if nextLoc != nil {
covBody = covBody[:nextLoc[0]]
}
p.EvaluationCriteria = strings.TrimSpace(covBody)
}

p.FilePath = filePath

if p.ID == "" {
return nil, fmt.Errorf("prompt missing required 'id' field: %s", filePath)
}

return p, nil
}

// ParsePromptYAML parses a pure YAML prompt file (.prompt.yaml or .prompt.yml)
// into a Prompt struct. All fields including prompt_text and evaluation_criteria
// are expressed as top-level YAML keys.
// It supports both the new nested properties format and the old flat format.
func ParsePromptYAML(content []byte, filePath string) (*Prompt, error) {
var raw rawFrontmatter
dec := yaml.NewDecoder(bytes.NewReader(content))
if err := dec.Decode(&raw); err != nil {
return nil, fmt.Errorf("parsing YAML prompt %s: %w", filePath, err)
}

p := rawToPrompt(&raw)
p.PromptText = raw.PromptTextField
p.EvaluationCriteria = raw.EvaluationCriteriaField
p.FilePath = filePath

if p.ID == "" {
return nil, fmt.Errorf("prompt missing required 'id' field: %s", filePath)
}

return p, nil
}
