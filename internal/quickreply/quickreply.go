package quickreply

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"

	"reasonix/internal/config"
)

const projectConfigTemplate = `# Project Quick Reply configuration.
# base_dirs are matched against the current project root using the longest
# matching prefix. The relative path becomes the quick-reply project ID.
base_dirs = ["~/Projects"]
`

type projectLocation struct {
	projectDir string
	replyPath  string
}

type fileBackup struct {
	path   string
	data   []byte
	exists bool
}

// Manager loads and saves public and per-project quick replies.
// Load and Save are safe for concurrent callers.
type Manager struct {
	mu          sync.Mutex
	publicPath  string
	projectRoot string
}

// NewManager creates a Manager for the provided public and project storage roots.
func NewManager(publicPath, projectRoot string) *Manager {
	return &Manager{
		publicPath:  strings.TrimSpace(publicPath),
		projectRoot: strings.TrimSpace(projectRoot),
	}
}

// Load reads public replies and, when workspaceRoot is configured, merges in
// the current project's replies.
func (m *Manager) Load(workspaceRoot string) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	project, projectReplies, err := m.loadProjectRepliesLocked(workspaceRoot)
	if err != nil {
		return Snapshot{}, err
	}
	publicReplies := loadRepliesFile(m.publicPath, ScopePublic)
	return Snapshot{
		Project: project,
		Replies: append(publicReplies, projectReplies...),
	}, nil
}

// Save replaces the public/project replies for the current workspace and
// returns the updated merged snapshot.
func (m *Manager) Save(workspaceRoot string, expectedProjectRoot string, replies []QuickReply) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	publicReplies := make([]QuickReply, 0, len(replies))
	projectReplies := make([]QuickReply, 0, len(replies))
	for _, reply := range sanitizeReplies(replies, ScopePublic) {
		switch normalizeScope(reply.Scope, ScopePublic) {
		case ScopeProject:
			projectReplies = append(projectReplies, QuickReply{
				Name:     reply.Name,
				Body:     reply.Body,
				Category: reply.Category,
				Scope:    ScopeProject,
			})
		default:
			publicReplies = append(publicReplies, QuickReply{
				Name:     reply.Name,
				Body:     reply.Body,
				Category: reply.Category,
				Scope:    ScopePublic,
			})
		}
	}
	projectRequired := len(projectReplies) > 0 || strings.TrimSpace(expectedProjectRoot) != ""
	var (
		project Project
		loc     projectLocation
		err     error
	)
	if projectRequired {
		project, err = m.resolveProjectLocked(workspaceRoot)
		if err != nil {
			if projectRequired {
				return Snapshot{}, err
			}
			return Snapshot{Replies: publicReplies}, nil
		}
		if expected := cleanPath(expectedProjectRoot); expected != "" && !samePath(expected, project.Root) {
			return Snapshot{}, fmt.Errorf("%w: loaded %q but current project is %q", ErrProjectContextChanged, expected, project.Root)
		}
		loc = m.projectLocation(project.ID)
		if err := ensureProjectRootLocked(m.projectRoot); err != nil {
			return Snapshot{}, err
		}
	}

	publicBackup, err := snapshotFile(m.publicPath)
	if err != nil {
		return Snapshot{}, err
	}
	var projectBackup fileBackup
	if loc.replyPath != "" {
		projectBackup, err = snapshotFile(loc.replyPath)
		if err != nil {
			return Snapshot{}, err
		}
		if err := saveRepliesFile(loc.replyPath, projectReplies, ScopeProject); err != nil {
			return Snapshot{}, err
		}
	}
	if err := saveRepliesFile(m.publicPath, publicReplies, ScopePublic); err != nil {
		restoreErrs := make([]string, 0, 2)
		if restoreErr := publicBackup.restore(); restoreErr != nil {
			restoreErrs = append(restoreErrs, "public replies: "+restoreErr.Error())
		}
		if loc.replyPath != "" {
			if restoreErr := projectBackup.restore(); restoreErr != nil {
				restoreErrs = append(restoreErrs, "project replies: "+restoreErr.Error())
			}
		}
		if len(restoreErrs) > 0 {
			return Snapshot{}, fmt.Errorf("save public quick replies: %w (rollback %s)", err, strings.Join(restoreErrs, "; "))
		}
		return Snapshot{}, err
	}
	if loc.replyPath == "" {
		return Snapshot{Replies: publicReplies}, nil
	}
	return Snapshot{Project: project, Replies: append(publicReplies, projectReplies...)}, nil
}

func (m *Manager) loadProjectRepliesLocked(workspaceRoot string) (Project, []QuickReply, error) {
	project, err := m.resolveProjectLocked(workspaceRoot)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotConfigured) {
			return Project{}, []QuickReply{}, nil
		}
		return Project{}, nil, err
	}
	loc := m.projectLocation(project.ID)
	return project, loadRepliesFile(loc.replyPath, ScopeProject), nil
}

var ErrWorkspaceNotConfigured = errors.New("quickreply: workspace root not configured")
var ErrProjectContextChanged = errors.New("quickreply: current project changed")

func (m *Manager) resolveProjectLocked(workspaceRoot string) (Project, error) {
	root := cleanPath(workspaceRoot)
	if root == "" {
		return Project{}, ErrWorkspaceNotConfigured
	}
	if strings.TrimSpace(m.projectRoot) == "" {
		return Project{}, errors.New("quickreply: project storage root not configured")
	}
	if err := ensureProjectRootLocked(m.projectRoot); err != nil {
		return Project{}, err
	}
	baseDirs := loadBaseDirs(filepath.Join(m.projectRoot, "config.toml"))
	baseDir := longestBaseMatch(root, baseDirs)
	id := ""
	if baseDir != "" {
		if rel, err := filepath.Rel(baseDir, root); err == nil {
			id = filepath.ToSlash(rel)
		}
	}
	if id == "" || id == "." {
		id = externalProjectID(root)
	}
	name := filepath.Base(root)
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = id
	}
	return Project{
		ID:      id,
		Name:    name,
		Root:    root,
		BaseDir: baseDir,
	}, nil
}

func (m *Manager) projectLocation(projectID string) projectLocation {
	projectDir := filepath.Join(m.projectRoot, "projects", filepath.FromSlash(projectID))
	return projectLocation{
		projectDir: projectDir,
		replyPath:  filepath.Join(projectDir, "quick-replies.toml"),
	}
}

func ensureProjectRootLocked(root string) error {
	if strings.TrimSpace(root) == "" {
		return errors.New("quickreply: project storage root not configured")
	}
	if err := os.MkdirAll(filepath.Join(root, "projects"), 0o755); err != nil {
		return err
	}
	return ensureTextFile(filepath.Join(root, "config.toml"), projectConfigTemplate)
}

func loadBaseDirs(path string) []string {
	cfg := configFile{}
	if _, err := toml.DecodeFile(path, &cfg); err != nil || len(cfg.BaseDirs) == 0 {
		return defaultBaseDirs()
	}
	out := make([]string, 0, len(cfg.BaseDirs))
	seen := map[string]bool{}
	for _, dir := range cfg.BaseDirs {
		if cleaned := cleanPath(dir); cleaned != "" && !seen[cleaned] {
			seen[cleaned] = true
			out = append(out, cleaned)
		}
	}
	if len(out) == 0 {
		return defaultBaseDirs()
	}
	return out
}

func loadRepliesFile(path string, defaultScope string) []QuickReply {
	if strings.TrimSpace(path) == "" {
		return []QuickReply{}
	}
	var s store
	_, err := toml.DecodeFile(path, &s)
	if err != nil {
		if !os.IsNotExist(err) {
			return []QuickReply{}
		}
		return []QuickReply{}
	}
	if s.QuickReplies == nil {
		return []QuickReply{}
	}
	return sanitizeReplies(s.QuickReplies, defaultScope)
}

func saveRepliesFile(path string, replies []QuickReply, defaultScope string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("quickreply: path not configured")
	}
	if replies == nil {
		replies = []QuickReply{}
	}
	s := store{QuickReplies: sanitizeReplies(replies, defaultScope)}
	b, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	return atomicWrite(path, b)
}

func sanitizeReplies(replies []QuickReply, defaultScope string) []QuickReply {
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
			Scope:    normalizeScope(reply.Scope, defaultScope),
		})
	}
	return out
}

func normalizeScope(scope string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case ScopeProject:
		return ScopeProject
	case ScopePublic:
		return ScopePublic
	default:
		if strings.TrimSpace(fallback) == ScopeProject {
			return ScopeProject
		}
		return ScopePublic
	}
}

func cleanPath(path string) string {
	path = strings.TrimSpace(config.ExpandVars(path))
	if path == "" {
		return ""
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if path == "~" {
				path = home
			} else {
				path = filepath.Join(home, path[2:])
			}
		}
	}
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Clean(path)
}

func longestBaseMatch(root string, baseDirs []string) string {
	best := ""
	for _, base := range baseDirs {
		base = cleanPath(base)
		if base == "" {
			continue
		}
		rel, err := filepath.Rel(base, root)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if len(base) > len(best) {
			best = base
		}
	}
	return best
}

func externalProjectID(root string) string {
	cleaned := filepath.Clean(root)
	vol := filepath.VolumeName(cleaned)
	if vol != "" {
		cleaned = strings.TrimPrefix(cleaned, vol)
		vol = strings.TrimRight(strings.ReplaceAll(vol, ":", ""), string(filepath.Separator))
	}
	parts := strings.FieldsFunc(cleaned, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return "external/project"
	}
	if vol != "" {
		parts = append([]string{vol}, parts...)
	}
	return filepath.ToSlash(filepath.Join(append([]string{"external"}, parts...)...))
}

func defaultBaseDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{filepath.Join(home, "Projects")}
}

func ensureTextFile(path, contents string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return atomicWrite(path, []byte(contents))
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return os.WriteFile(path, data, 0o644)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return os.WriteFile(path, data, 0o644)
	}
	if err := tmpFile.Close(); err != nil {
		return os.WriteFile(path, data, 0o644)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return os.WriteFile(path, data, 0o644)
	}
	return nil
}

func snapshotFile(path string) (fileBackup, error) {
	if strings.TrimSpace(path) == "" {
		return fileBackup{}, errors.New("quickreply: path not configured")
	}
	data, err := os.ReadFile(path)
	if err == nil {
		return fileBackup{path: path, data: data, exists: true}, nil
	}
	if os.IsNotExist(err) {
		return fileBackup{path: path, exists: false}, nil
	}
	return fileBackup{}, err
}

func (b fileBackup) restore() error {
	if strings.TrimSpace(b.path) == "" {
		return nil
	}
	if !b.exists {
		if err := os.Remove(b.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return atomicWrite(b.path, b.data)
}

func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// DefaultPath returns the conventional path for the shared quick-replies file
// under the Reasonix home directory.
func DefaultPath(reasonixHome string) string {
	if strings.TrimSpace(reasonixHome) == "" {
		return ""
	}
	return filepath.Join(reasonixHome, "quick-replies.toml")
}

// DefaultProjectPath returns the conventional root for project-scoped
// quick-replies under the Reasonix home directory.
func DefaultProjectPath(reasonixHome string) string {
	if strings.TrimSpace(reasonixHome) == "" {
		return ""
	}
	return filepath.Join(reasonixHome, "project-quick-replies")
}
