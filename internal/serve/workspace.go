package serve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"reasonix/internal/config"
)

var workspaceListPath = func() string {
	dir := config.MemoryUserDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "desktop-workspaces.json")
}

// loadWorkspaces returns the list of remembered workspace directories,
// newest first, deduplicated. Returns nil on any error (no file yet, etc.).
func loadWorkspaces() []string {
	p := workspaceListPath()
	if p == "" {
		return nil
	}
	var paths []string
	b, err := os.ReadFile(p)
	if err != nil || json.Unmarshal(b, &paths) != nil {
		return nil
	}
	out := make([]string, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out
}

// rememberWorkspace adds dir to the workspace list (newest first, deduped,
// capped at 12 entries). It is safe to call when the list file doesn't exist yet.
func rememberWorkspace(dir string) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	paths := []string{dir}
	for _, path := range loadWorkspaces() {
		if path != dir {
			paths = append(paths, path)
		}
		if len(paths) >= 12 {
			break
		}
	}
	p := workspaceListPath()
	if p == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	if b, err := json.MarshalIndent(paths, "", "  "); err == nil {
		// Atomic write: temp file + rename so concurrent readers never see
		// a partially written file.
		tmpFile, err := os.CreateTemp(filepath.Dir(p), "workspaces.*.tmp")
		if err == nil {
			tmpPath := tmpFile.Name()
			if _, wErr := tmpFile.Write(b); wErr == nil {
				tmpFile.Close()
				_ = os.Rename(tmpPath, p)
			} else {
				tmpFile.Close()
				os.Remove(tmpPath)
				// Fallback: direct write when temp write fails
				_ = os.WriteFile(p, b, 0o644)
			}
		} else {
			// Fallback: direct write when temp creation fails (e.g. permissions)
			_ = os.WriteFile(p, b, 0o644)
		}
	}
}
