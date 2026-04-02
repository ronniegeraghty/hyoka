// Package utils provides common utility functions for file system and string manipulation.
package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ReadDirFiles reads all files in a directory recursively, skipping hidden
// files, build artifact directories, and binary/large files (>1 MB per file,
// 10 MB total). Returns a map of relative path -> content.
func ReadDirFiles(dir string) (map[string]string, error) {
	files := make(map[string]string)
	var totalSize int64
	const maxTotalSize = 10 << 20 // 10 MB total cap for review prompt safety
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") && path != dir {
				return filepath.SkipDir
			}
			if IsBuildArtifactDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip binary/large files (limit 1MB)
		if info.Size() > 1<<20 {
			return nil
		}
		// Enforce total size cap
		if totalSize+info.Size() > maxTotalSize {
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
		totalSize += info.Size()
		return nil
	})
	return files, err
}

// IsBuildArtifactDir returns true for well-known build artifact directory
// names that should be excluded from file listings, copies, and review prompts.
func IsBuildArtifactDir(name string) bool {
	switch name {
	case "target", // Rust/Cargo
		"node_modules",  // Node.js/npm
		"__pycache__",   // Python
		".venv", "venv", // Python virtual envs
		"bin", "obj", // .NET
		"build", "dist", // general build output
		"out",              // Java/Gradle
		"vendor",           // Go/PHP
		"packages",         // NuGet
		".gradle",          // Gradle cache
		".cargo",           // Cargo cache
		"debug", "release": // Rust profile subdirs
		return true
	}
	return false
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
