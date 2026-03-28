// Package eval provides the core evaluation engine, workspace management, and Copilot interaction logic.
package eval

import (
	"fmt"
	"log/slog"
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
	prefix := fmt.Sprintf("hyoka-%s-%s-", promptID, configName)
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
			slog.Warn("Skipping symlink in starter project", "path", rel)
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

// NewIsolatedConfigDir creates an empty temporary directory to serve as the
// Copilot CLI configuration directory. By pointing ConfigDir at this empty
// directory, user-level skills and settings from ~/.config/github-copilot/
// are excluded from eval sessions. Only skills explicitly listed in the eval
// config's SkillDirectories are loaded (fixes #21).
// The caller must defer os.RemoveAll on the returned path.
func NewIsolatedConfigDir() (string, error) {
	dir, err := os.MkdirTemp("", "hyoka-config-*")
	if err != nil {
		return "", fmt.Errorf("creating isolated config dir: %w", err)
	}
	return dir, nil
}

// NewReviewerWorkspace creates an ephemeral temporary workspace and copies
// all files from sourceDir into it. Reviewers operate on this copy so they
// cannot modify the original generated output (fixes #26).
// The caller must defer os.RemoveAll on the returned path.
func NewReviewerWorkspace(sourceDir string) (string, error) {
	dir, err := os.MkdirTemp("", "hyoka-review-*")
	if err != nil {
		return "", fmt.Errorf("creating reviewer workspace: %w", err)
	}
	if err := copyDir(sourceDir, dir); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("copying files to reviewer workspace: %w", err)
	}
	return dir, nil
}

// ValidateWorkspaceContainment checks whether any new items appeared in dir
// since the pre-eval snapshot. Returns the names of items that escaped the
// workspace boundary. Called after recoverMisplacedFiles as a safety net
// to catch anything recovery could not handle (fixes #26).
func ValidateWorkspaceContainment(dir string, preSnapshot map[string]bool) []string {
	if preSnapshot == nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var escaped []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !preSnapshot[e.Name()] {
			escaped = append(escaped, e.Name())
		}
	}
	return escaped
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

// junkDirs lists directory names that are build/runtime artifacts and should
// be deleted rather than recovered into the workspace.
var junkDirs = map[string]bool{
	"__pycache__":  true,
	"node_modules": true,
	"venv":         true,
	".venv":        true,
	"env":          true,
	".tox":         true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"bin":          true,
	"obj":          true,
}

// snapshotDir returns a set of non-hidden entry names (files AND directories) in
// a directory (non-recursive). Capturing directories lets recoverMisplacedFiles
// detect new directories created during an eval run.
func snapshotDir(dir string) map[string]bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			names[e.Name()] = true
		}
	}
	return names
}

// recoverMisplacedFiles moves files and directories that appeared in dir since
// the snapshot into destDir. Files with recognized code extensions (or no
// extension) are moved; new directories are either moved into the workspace or
// deleted if they match a known junk pattern. Returns the count of recovered
// items (files + directories moved or cleaned up).
func recoverMisplacedFiles(dir string, preSnapshot map[string]bool, destDir string, _ string, debug bool) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	recovered := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if preSnapshot[e.Name()] {
			continue // existed before eval
		}

		src := filepath.Join(dir, e.Name())

		if e.IsDir() {
			// Junk directories → just delete
			if junkDirs[e.Name()] {
				if err := os.RemoveAll(src); err == nil {
					recovered++
					if debug {
						slog.Debug("Deleted junk directory", "path", src)
					}
				}
				continue
			}
			// Other new directories → move into workspace
			dst := filepath.Join(destDir, e.Name())
			if err := os.Rename(src, dst); err != nil {
				// Rename may fail across filesystems; fall back to copy+delete
				if err := copyDir(src, dst); err == nil {
					os.RemoveAll(src)
				} else {
					continue
				}
			}
			recovered++
			if debug {
				slog.Debug("Recovered misplaced directory", "src", src, "dst", dst)
			}
			continue
		}

		// Regular file handling (unchanged logic)
		ext := strings.ToLower(filepath.Ext(e.Name()))
		// Also recover extensionless files like "Dockerfile", "Makefile"
		if !codeFileExts[ext] && ext != "" {
			continue
		}

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
			slog.Debug("Recovered misplaced file", "src", src, "dst", dst)
		}
	}
	return recovered
}
