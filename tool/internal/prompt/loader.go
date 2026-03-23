// Package prompt provides functionality to load and parse prompt files.
package prompt

import (
"fmt"
"os"
"path/filepath"
"strings"
)

// LoadPrompts walks the directory tree at root and loads all .prompt.md files.
func LoadPrompts(root string) ([]*Prompt, error) {
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
return nil
})
if err != nil {
return nil, fmt.Errorf("walking prompt directory %s: %w", root, err)
}

return prompts, nil
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
