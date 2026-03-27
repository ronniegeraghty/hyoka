package build

import (
"testing"
)

func TestDetectLanguage(t *testing.T) {
tests := []struct {
input    string
expected string
isNil    bool
}{
{"dotnet", "dotnet", false},
{"go", "go", false},
{"python", "python", false},
{"typescript", "typescript", false},
{"javascript", "javascript", false},
{"rust", "rust", false},
{"java-maven", "java-maven", false},
{"cpp", "cpp", false},
// Aliases
{"csharp", "dotnet", false},
{"c#", "dotnet", false},
{"py", "python", false},
{"golang", "go", false},
{"ts", "typescript", false},
{"js", "javascript", false},
{"c++", "cpp", false},
// Unknown
{"unknown-lang", "", true},
}

for _, tt := range tests {
t.Run(tt.input, func(t *testing.T) {
lc := DetectLanguage(tt.input)
if tt.isNil {
if lc != nil {
t.Errorf("expected nil for %q, got %+v", tt.input, lc)
}
return
}
if lc == nil {
t.Fatalf("expected non-nil for %q", tt.input)
}
if lc.Name != tt.expected {
t.Errorf("expected name %q, got %q", tt.expected, lc.Name)
}
})
}
}

func TestSupportedLanguages(t *testing.T) {
langs := SupportedLanguages()
if len(langs) == 0 {
t.Fatal("expected at least one supported language")
}
names := make(map[string]bool)
for _, l := range langs {
if l.Name == "" {
t.Error("language config has empty name")
}
if l.BuildCmd == "" {
t.Errorf("language %q has empty build command", l.Name)
}
if len(l.Extensions) == 0 {
t.Errorf("language %q has no extensions", l.Name)
}
names[l.Name] = true
}
// Check some expected languages exist
for _, expected := range []string{"dotnet", "go", "python", "typescript", "rust"} {
if !names[expected] {
t.Errorf("expected language %q not found", expected)
}
}
}
