// Package eval provides the core evaluation engine, workspace management, and Copilot interaction logic.
package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workspace manages a temporary directory for an evaluation run.
type Workspace struct {
	BaseDir string
	Dir     string
}

// NewWorkspace creates a new workspace in the OS temp directory.
// The workspace is ephemeral — generated files should be copied out
// before calling Cleanup.
func NewWorkspace(promptID, configName string) (*Workspace, error) {
	prefix := fmt.Sprintf("azsdk-prompt-eval-%s-%s-", promptID, configName)
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return nil, fmt.Errorf("creating temp workspace: %w", err)
	}

	return &Workspace{
		BaseDir: os.TempDir(),
		Dir:     dir,
	}, nil
}

// Cleanup removes the workspace directory.
func (w *Workspace) Cleanup() error {
	return os.RemoveAll(w.Dir)
}

// ListFiles returns all non-hidden files in the workspace, relative to its root.
func (w *Workspace) ListFiles() ([]string, error) {
	return listFiles(w.Dir)
}

// CopyFilesTo copies all non-hidden workspace files into destDir,
// preserving relative paths. Returns the list of files copied.
func (w *Workspace) CopyFilesTo(destDir string) ([]string, error) {
	files, err := w.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("listing workspace files: %w", err)
	}
	if len(files) == 0 {
		return nil, nil
	}

	for _, rel := range files {
		src := filepath.Join(w.Dir, rel)
		dst := filepath.Join(destDir, rel)

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return nil, fmt.Errorf("creating dir for %s: %w", rel, err)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", rel, err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", rel, err)
		}
	}

	return files, nil
}

// listFiles is a helper used by Workspace and CopilotSDKEvaluator.
func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}