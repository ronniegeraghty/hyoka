package build

// LanguageConfig holds build commands and file extensions for a language.
type LanguageConfig struct {
Name       string
Extensions []string
BuildCmd   string
BuildArgs  []string
}

// SupportedLanguages returns the known language configurations.
func SupportedLanguages() []LanguageConfig {
return []LanguageConfig{
{Name: "dotnet", Extensions: []string{".cs", ".csproj", ".sln"}, BuildCmd: "dotnet", BuildArgs: []string{"build"}},
{Name: "java-maven", Extensions: []string{".java"}, BuildCmd: "mvn", BuildArgs: []string{"compile", "-q"}},
{Name: "java-gradle", Extensions: []string{".java"}, BuildCmd: "gradle", BuildArgs: []string{"compileJava"}},
{Name: "python", Extensions: []string{".py"}, BuildCmd: "python3", BuildArgs: []string{"-m", "py_compile"}},
{Name: "go", Extensions: []string{".go"}, BuildCmd: "go", BuildArgs: []string{"build", "./..."}},
{Name: "typescript", Extensions: []string{".ts", ".tsx"}, BuildCmd: "npx", BuildArgs: []string{"tsc", "--noEmit"}},
{Name: "javascript", Extensions: []string{".js", ".mjs"}, BuildCmd: "node", BuildArgs: []string{"--check"}},
{Name: "rust", Extensions: []string{".rs"}, BuildCmd: "cargo", BuildArgs: []string{"build"}},
{Name: "cpp", Extensions: []string{".cpp", ".cc", ".h", ".hpp"}, BuildCmd: "cmake", BuildArgs: []string{"-B", "build"}},
}
}

// DetectLanguage determines the language from the prompt metadata or file extensions.
func DetectLanguage(language string) *LanguageConfig {
for _, lc := range SupportedLanguages() {
if lc.Name == language {
return &lc
}
}
// Map common aliases
aliases := map[string]string{
"csharp":     "dotnet",
"c#":         "dotnet",
"java":       "java-maven",
"py":         "python",
"golang":     "go",
"ts":         "typescript",
"js":         "javascript",
"node":       "javascript",
"c++":        "cpp",
	"js-ts":    "typescript",
}
if canonical, ok := aliases[language]; ok {
for _, lc := range SupportedLanguages() {
if lc.Name == canonical {
return &lc
}
}
}
return nil
}
