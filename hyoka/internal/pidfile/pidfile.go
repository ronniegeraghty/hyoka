// Package pidfile manages PID files that track SDK-spawned Copilot processes.
//
// When hyoka starts a Copilot process, it writes a PID file to a known
// directory. On normal shutdown the file is removed. If hyoka crashes, the
// PID file remains and the clean command can discover orphaned processes
// by reading stale PID files.
package pidfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Info describes a tracked Copilot process.
type Info struct {
	PID      int    `json:"pid"`
	PromptID string `json:"prompt_id,omitempty"`
	Config   string `json:"config,omitempty"`
}

// DirFn returns the PID file directory. Package-level variable so tests
// can override it.
var DirFn = defaultDir

func defaultDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "copilot-cli", "hyoka-pids")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".copilot", "hyoka-pids")
}

// Write creates a PID file for the given process.
func Write(info Info) error {
	dir := DirFn()
	if dir == "" {
		return fmt.Errorf("cannot determine PID file directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.json", info.PID)), data, 0o644)
}

// Remove deletes the PID file for a given process.
func Remove(pid int) {
	dir := DirFn()
	if dir != "" {
		os.Remove(filepath.Join(dir, fmt.Sprintf("%d.json", pid)))
	}
}

// ReadAlive reads all PID files and returns entries whose processes are
// still running. Stale PID files (dead processes) are removed automatically.
func ReadAlive() ([]Info, error) {
	dir := DirFn()
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var alive []Info
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var info Info
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		if info.PID > 0 && isProcessAlive(info.PID) {
			alive = append(alive, info)
		} else {
			// Stale PID file — clean it up.
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return alive, nil
}

// CleanAll removes the entire PID file directory.
func CleanAll() error {
	dir := DirFn()
	if dir == "" {
		return nil
	}
	return os.RemoveAll(dir)
}
