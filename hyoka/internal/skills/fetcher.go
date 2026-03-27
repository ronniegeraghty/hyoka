// Package skills resolves unified Skill entries into local directory paths
// that can be passed to the Copilot SDK session as skill directories.
package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ronniegeraghty/azure-sdk-prompts/hyoka/internal/config"
)

// ResolveSkillDirs takes a list of Skill entries and resolves them to
// absolute directory paths. The baseDir is used as the root for resolving
// relative local paths.
//
//   - type: local  → resolves path (supports glob patterns like "./skills/generator/*")
//   - type: remote → fetches from GitHub repo via "npx skills add", returns the install dir
func ResolveSkillDirs(skills []config.Skill, baseDir string) ([]string, error) {
	var dirs []string
	for _, s := range skills {
		switch s.Type {
		case "local":
			resolved, err := resolveLocal(s.Path, baseDir)
			if err != nil {
				return nil, fmt.Errorf("resolving local skill %q: %w", s.Path, err)
			}
			dirs = append(dirs, resolved...)
		case "remote":
			dir, err := fetchRemote(s, baseDir)
			if err != nil {
				return nil, fmt.Errorf("fetching remote skill %s/%s: %w", s.Repo, s.Name, err)
			}
			dirs = append(dirs, dir)
		default:
			return nil, fmt.Errorf("unknown skill type %q", s.Type)
		}
	}
	return dirs, nil
}

// resolveLocal resolves a local skill path (supports globs) to absolute paths.
func resolveLocal(path, baseDir string) ([]string, error) {
	// Make relative paths absolute based on baseDir
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	// Check for glob characters
	if strings.ContainsAny(path, "*?[") {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", path, err)
		}
		// Filter to directories only
		var dirs []string
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil {
				continue
			}
			if info.IsDir() {
				abs, _ := filepath.Abs(m)
				dirs = append(dirs, abs)
			}
		}
		return dirs, nil
	}

	// Non-glob: try to resolve the path, checking candidate locations
	candidates := []string{
		path,
		filepath.Join(baseDir, path),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(c)
			return []string{abs}, nil
		}
	}

	// Path doesn't exist yet — return absolute form anyway
	abs, _ := filepath.Abs(path)
	return []string{abs}, nil
}

// fetchRemote fetches a remote skill from a GitHub repo using npx skills add.
// Returns the directory where the skill was installed.
func fetchRemote(s config.Skill, baseDir string) (string, error) {
	// Determine install directory: use a skills cache dir under baseDir
	installDir := filepath.Join(baseDir, ".skills-cache", s.Repo)
	if s.Name != "" {
		installDir = filepath.Join(installDir, s.Name)
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("creating skill install dir: %w", err)
	}

	// Use npx skills add to fetch the skill
	args := []string{"skills", "add", s.Repo, "--directory", installDir}
	if s.Name != "" {
		args = append(args, "--name", s.Name)
	}

	fmt.Printf("Fetching remote skill: %s (repo: %s)\n", s.Name, s.Repo)
	cmd := exec.Command("npx", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("npx skills add: %w", err)
	}

	abs, _ := filepath.Abs(installDir)
	return abs, nil
}
