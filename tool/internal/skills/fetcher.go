// Package skills provides utilities for fetching skills from remote registries.
package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
)

// Fetch downloads remote skills into a temporary directory and returns
// the path. Each RemoteSkill entry produces one or more npx invocations
// of the form: npx skills add <repo>/<skill> --directory <dir>
//
// The caller must call Cleanup(dir) when the temp directory is no longer needed.
func Fetch(remoteSkills []config.RemoteSkill) (string, error) {
	if len(remoteSkills) == 0 {
		return "", nil
	}

	dir, err := os.MkdirTemp("", "azsdk-remote-skills-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir for remote skills: %w", err)
	}

	for _, rs := range remoteSkills {
		for _, skill := range rs.Skills {
			ref := rs.Repo + "/" + skill
			fmt.Printf("Fetching remote skill: %s → %s\n", ref, filepath.Base(dir))
			cmd := exec.Command("npx", "skills", "add", ref, "--directory", dir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Clean up on failure so we don't leak temp dirs
				os.RemoveAll(dir)
				return "", fmt.Errorf("fetching skill %q: %w", ref, err)
			}
		}
	}

	return dir, nil
}

// Cleanup removes a temporary skill directory created by Fetch.
func Cleanup(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
