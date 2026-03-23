// Package build provides functionality to verify build commands for various languages.
package build

import (
"context"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
"time"
)

// BuildResult holds the outcome of a build verification.
type BuildResult struct {
Language string        `json:"language"`
Command  string        `json:"command"`
ExitCode int           `json:"exit_code"`
Stdout   string        `json:"stdout"`
Stderr   string        `json:"stderr"`
Duration time.Duration `json:"duration_ms"`
Success  bool          `json:"success"`
}

// Verify runs the appropriate build command for the given language in workDir.
func Verify(ctx context.Context, language string, workDir string) (*BuildResult, error) {
lc := DetectLanguage(language)
if lc == nil {
return &BuildResult{
Language: language,
Command:  "",
ExitCode: -1,
Stderr:   fmt.Sprintf("unsupported language: %s", language),
Success:  false,
}, nil
}

cmdStr, args := buildCommand(lc, workDir)

start := time.Now()

cmd := exec.CommandContext(ctx, cmdStr, args...)
cmd.Dir = workDir

var stdout, stderr strings.Builder
cmd.Stdout = &stdout
cmd.Stderr = &stderr

err := cmd.Run()
duration := time.Since(start)

result := &BuildResult{
Language: lc.Name,
Command:  fmt.Sprintf("%s %s", cmdStr, strings.Join(args, " ")),
Stdout:   stdout.String(),
Stderr:   stderr.String(),
Duration: duration,
}

if err != nil {
if exitErr, ok := err.(*exec.ExitError); ok {
result.ExitCode = exitErr.ExitCode()
} else {
result.ExitCode = -1
result.Stderr = err.Error()
}
result.Success = false
} else {
result.ExitCode = 0
result.Success = true
}

return result, nil
}

// buildCommand returns the command and args for the given language config.
func buildCommand(lc *LanguageConfig, workDir string) (string, []string) {
switch lc.Name {
case "dotnet":
return "sh", []string{"-c", "dotnet restore && dotnet build"}
case "python":
// Find all .py files
var pyFiles []string
filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
if err == nil && !info.IsDir() && strings.HasSuffix(path, ".py") {
pyFiles = append(pyFiles, path)
}
return nil
})
if len(pyFiles) == 0 {
return "python3", []string{"-m", "py_compile", "/dev/null"}
}
args := []string{"-m", "py_compile"}
args = append(args, pyFiles...)
return "python3", args
case "javascript":
var jsFiles []string
filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
if err == nil && !info.IsDir() && (strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".mjs")) {
jsFiles = append(jsFiles, path)
}
return nil
})
if len(jsFiles) == 0 {
return "node", []string{"--check", "/dev/null"}
}
args := []string{"--check"}
args = append(args, jsFiles...)
return "node", args
default:
return lc.BuildCmd, lc.BuildArgs
}
}
