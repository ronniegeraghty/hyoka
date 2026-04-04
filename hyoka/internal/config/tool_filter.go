package config

import "fmt"

// ToolEntry represents a tool with optional property-based conditions.
// When the When map is empty, the tool is unconditionally included.
// When the When map has entries, all key-value pairs must match the
// prompt's properties for the tool to be included.
type ToolEntry struct {
Name string            `yaml:"name" json:"name"`
When map[string]string `yaml:"when,omitempty" json:"when,omitempty"`
}

// ResolveTools evaluates tool entries against prompt properties and returns
// the names of tools whose conditions are satisfied. An empty entries slice
// returns nil (meaning "all default tools").
func ResolveTools(entries []ToolEntry, properties map[string]string) []string {
if len(entries) == 0 {
return nil
}
var resolved []string
for _, e := range entries {
if matchesWhen(e.When, properties) {
resolved = append(resolved, e.Name)
}
}
return resolved
}

// matchesWhen returns true when every key-value pair in when matches the
// properties map. An empty when map always matches.
func matchesWhen(when map[string]string, props map[string]string) bool {
for k, v := range when {
if props[k] != v {
return false
}
}
return true
}

// validateToolEntry checks that a ToolEntry has valid fields.
func validateToolEntry(entry ToolEntry, configName string, idx int) error {
if entry.Name == "" {
return fmt.Errorf("config %q: tools[%d] missing name", configName, idx)
}
return nil
}
