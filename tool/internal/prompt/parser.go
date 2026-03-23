package prompt

import (
"fmt"
"regexp"
"strings"

"gopkg.in/yaml.v3"
)

var promptSectionRe = regexp.MustCompile(`(?m)^## Prompt\s*\n`)
var evaluationCriteriaRe = regexp.MustCompile(`(?m)^## Evaluation Criteria\s*\n`)

// ParsePromptFile parses a .prompt.md file's content into a Prompt struct.
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

var p Prompt
if err := yaml.Unmarshal([]byte(frontmatter), &p); err != nil {
return nil, fmt.Errorf("parsing frontmatter in %s: %w", filePath, err)
}

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

return &p, nil
}
