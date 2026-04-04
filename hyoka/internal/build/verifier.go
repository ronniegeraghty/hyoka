// Package build provides functionality to verify build commands for various languages.
package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronniegeraghty/hyoka/internal/eval"
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
		slog.Warn("Unsupported language for build verification", "language", language)
		return &BuildResult{
			Language: language,
			Command:  "",
			ExitCode: -1,
			Stderr:   fmt.Sprintf("unsupported language: %s", language),
			Success:  false,
		}, nil
	}

	commands := buildCommands(lc, workDir)
	slog.Info("Starting build verification", "language", lc.Name, "steps", len(commands))

	start := time.Now()

	var stdout, stderr strings.Builder
	for _, step := range commands {
		cmd := exec.CommandContext(ctx, step.Cmd, step.Args...)
		cmd.Dir = workDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			duration := time.Since(start)
			result := &BuildResult{
				Language: lc.Name,
				Command:  step.String(),
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Duration: duration,
			}
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			} else {
				result.ExitCode = -1
				result.Stderr = err.Error()
			}
			result.Success = false
			slog.Info("Build verification failed", "language", lc.Name, "step", step.String(), "exit_code", result.ExitCode, "duration", duration)
			return result, nil
		}
	}

	duration := time.Since(start)
	last := commands[len(commands)-1]
	result := &BuildResult{
		Language: lc.Name,
		Command:  last.String(),
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Success:  true,
	}
	slog.Info("Build verification passed", "language", lc.Name, "duration", duration)
	return result, nil
}

// buildStep describes a single command to execute during build verification.
type buildStep struct {
	Cmd  string
	Args []string
}

func (s buildStep) String() string {
	return s.Cmd + " " + strings.Join(s.Args, " ")
}

// buildCommands returns the ordered list of commands to execute for the given language.
func buildCommands(lc *LanguageConfig, workDir string) []buildStep {
	switch lc.Name {
	case "dotnet":
		return []buildStep{
			{Cmd: "dotnet", Args: []string{"restore"}},
			{Cmd: "dotnet", Args: []string{"build"}},
		}
	case "python":
		var pyFiles []string
		if err := filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if d.IsDir() && eval.DefaultIgnoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(path, ".py") {
				pyFiles = append(pyFiles, path)
			}
			return nil
		}); err != nil {
			slog.Warn("Failed to walk directory for Python files", "dir", workDir, "error", err)
		}
		if len(pyFiles) == 0 {
			return []buildStep{{Cmd: "python3", Args: []string{"-m", "py_compile", os.DevNull}}}
		}
		args := []string{"-m", "py_compile"}
		args = append(args, pyFiles...)
		return []buildStep{{Cmd: "python3", Args: args}}
	case "javascript":
		var jsFiles []string
		if err := filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if d.IsDir() && eval.DefaultIgnoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			if !d.IsDir() && (strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".mjs")) {
				jsFiles = append(jsFiles, path)
			}
			return nil
		}); err != nil {
			slog.Warn("Failed to walk directory for JavaScript files", "dir", workDir, "error", err)
		}
		if len(jsFiles) == 0 {
			return []buildStep{{Cmd: "node", Args: []string{"--check", os.DevNull}}}
		}
		args := []string{"--check"}
		args = append(args, jsFiles...)
		return []buildStep{{Cmd: "node", Args: args}}
	default:
		return []buildStep{{Cmd: lc.BuildCmd, Args: lc.BuildArgs}}
	}
}
