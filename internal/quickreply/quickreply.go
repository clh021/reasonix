package quickreply

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

// Manager loads and saves QuickReply lists from a TOML file.
// Load and Save are safe for concurrent callers.
type Manager struct {
	mu   sync.Mutex
	path string // absolute path to the TOML file
}

// NewManager creates a Manager that reads from and writes to path.
// The file does not need to exist yet — Load returns an empty list when missing.
func NewManager(path string) *Manager {
	return &Manager{path: path}
}

// Load reads quick replies from the TOML file. Returns an empty (non-nil) slice
// when the file does not exist or cannot be parsed, so callers can always range
// over the result safely.
func (m *Manager) Load() []QuickReply {
	if m.path == "" {
		return []QuickReply{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var s store
	_, err := toml.DecodeFile(m.path, &s)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("quickreply: load failed", "path", m.path, "err", err)
		}
		return []QuickReply{}
	}
	if s.QuickReplies == nil {
		return []QuickReply{}
	}
	return sanitizeReplies(s.QuickReplies)
}

// Save atomically writes the provided quick replies to the TOML file.
// It creates the parent directory if it does not exist.
// Returns an error when the manager path is not configured.
func (m *Manager) Save(replies []QuickReply) error {
	if m.path == "" {
		return errors.New("quickreply: path not configured")
	}
	if replies == nil {
		replies = []QuickReply{}
	}
	s := store{QuickReplies: sanitizeReplies(replies)}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}

	b, err := toml.Marshal(s)
	if err != nil {
		return err
	}

	// Atomic write: temp file + rename so concurrent readers never see
	// a partially written file.
	tmpFile, err := os.CreateTemp(filepath.Dir(m.path), "quickreplies.*.tmp")
	if err != nil {
		// Fallback: direct write when temp creation fails (e.g. permissions).
		return os.WriteFile(m.path, b, 0o644)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, wErr := tmpFile.Write(b); wErr == nil {
		tmpFile.Close()
		if err := os.Rename(tmpPath, m.path); err != nil {
			// Rename failed (e.g. cross-device link); fall back to direct write.
			return os.WriteFile(m.path, b, 0o644)
		}
		return nil
	}
	tmpFile.Close()
	// Fallback: direct write when temp write fails.
	return os.WriteFile(m.path, b, 0o644)
}

func sanitizeReplies(replies []QuickReply) []QuickReply {
	out := make([]QuickReply, 0, len(replies))
	for _, reply := range replies {
		cat := reply.Category
		if !ValidCategoryID(cat) {
			cat = ""
		}
		out = append(out, QuickReply{
			Name:     reply.Name,
			Body:     reply.Body,
			Category: cat,
		})
	}
	return out
}

// DefaultPath returns the conventional path for the quick-replies file under
// the Reasonix home directory.
func DefaultPath(reasonixHome string) string {
	if reasonixHome == "" {
		return ""
	}
	return filepath.Join(reasonixHome, "quick-replies.toml")
}
