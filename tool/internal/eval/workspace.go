// Package eval provides the core evaluation engine, workspace management, and Copilot interaction logic.
package eval

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Workspace manages a directory for an evaluation run.
type Workspace struct {
	BaseDir string
	Dir     string
	persist bool // if true, Cleanup is a no-op
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

// NewWorkspaceAt creates a persistent workspace at the given directory.
// The directory is created if it doesn't exist. Cleanup is a no-op.
func NewWorkspaceAt(dir string) (*Workspace, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating workspace directory: %w", err)
	}
	return &Workspace{
		BaseDir: filepath.Dir(dir),
		Dir:     dir,
		persist: true,
	}, nil
}

// Cleanup removes the workspace directory (no-op for persistent workspaces).
func (w *Workspace) Cleanup() error {
	if w.persist {
		return nil
	}
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

// copyDir recursively copies src to dst, skipping symlinks to prevent escape.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip symlinks to prevent following links outside the source tree
		if info.Mode()&os.ModeSymlink != 0 {
			rel, _ := filepath.Rel(src, path)
			log.Printf("warning: skipping symlink in starter project: %s", rel)
			return nil
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

// codeFileExts lists extensions that indicate generated code files.
var codeFileExts = map[string]bool{
	".py": true, ".cs": true, ".java": true, ".go": true, ".rs": true,
	".ts": true, ".js": true, ".cpp": true, ".c": true, ".h": true,
	".csproj": true, ".sln": true, ".json": true, ".xml": true,
	".yaml": true, ".yml": true, ".toml": true, ".mod": true,
	".txt": true, ".md": true, ".gradle": true, ".kt": true,
	".swift": true, ".rb": true, ".sh": true, ".bat": true,
	".html": true, ".css": true, ".sum": true, ".lock": true,
	".cfg": true, ".ini": true, ".env": true, ".dockerfile": true,
}

// snapshotDir returns a set of non-hidden file names in a directory (non-recursive).
func snapshotDir(dir string) map[string]bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	files := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			files[e.Name()] = true
		}
	}
	return files
}

// recoverMisplacedFiles moves files that appeared in dir since the snapshot
// into destDir. Only moves files with recognized code extensions.
// Returns the count of recovered files.
func recoverMisplacedFiles(dir string, preSnapshot map[string]bool, destDir string, debugPrefix string, debug bool) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	recovered := 0
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if preSnapshot[e.Name()] {
			continue // existed before eval
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		// Also recover extensionless files like "Dockerfile", "Makefile"
		if !codeFileExts[ext] && ext != "" {
			continue
		}

		src := filepath.Join(dir, e.Name())
		dst := filepath.Join(destDir, e.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			continue
		}
		os.Remove(src) // clean up the misplaced file
		recovered++
		if debug {
			log.Printf("[DEBUG] %s: Recovered misplaced file: %s → %s", debugPrefix, src, dst)
		}
	}
	return recovered
}