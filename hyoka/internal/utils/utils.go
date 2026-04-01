// Package utils provides common utility functions for file system and string manipulation.
package utils

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ReadDirFiles reads all files in a directory (non-recursive, skipping hidden/binary).
// Returns a map of relative path -> content.
func ReadDirFiles(dir string) (map[string]string, error) {
	return ReadDirFilesFiltered(dir, nil)
}

// ReadDirFilesFiltered reads all files in a directory, skipping hidden files,
// binary/large files (>1MB), and any directories whose name appears in skipDirs.
// Pass nil for skipDirs to skip no directories (same behavior as ReadDirFiles).
func ReadDirFilesFiltered(dir string, skipDirs map[string]bool) (map[string]string, error) {
	files := make(map[string]string)
	skipped := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") && path != dir {
				return filepath.SkipDir
			}
			if len(skipDirs) > 0 && skipDirs[name] && path != dir {
				skipped++
				return filepath.SkipDir
			}
			return nil
		}
		// Skip binary/large files (limit 1MB)
		if info.Size() > 1<<20 {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(filepath.Base(rel), ".") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		files[rel] = string(data)
		return nil
	})
	if skipped > 0 {
		slog.Debug("ReadDirFilesFiltered skipped directories", "dir", dir, "skipped_dirs", skipped)
	}
	return files, err
}

// ExtractJSON finds the first JSON object in the text, stripping markdown code fences.
func ExtractJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.LastIndex(text, "```"); idx >= 0 {
			text = text[:idx]
		}
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.LastIndex(text, "```"); idx >= 0 {
			text = text[:idx]
		}
	}
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return ""
}
