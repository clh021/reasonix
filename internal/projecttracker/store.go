package projecttracker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"

	"reasonix/internal/config"
)

var (
	statusOrder   = []string{"wishlist", "planned", "in_progress", "custom_active", "upstream_covered", "dropped"}
	priorityOrder = []string{"high", "medium", "low"}
	statusRank    = indexMap(statusOrder)
	priorityRank  = indexMap(priorityOrder)
)

const trackerReadme = `# Project Tracker 存储目录说明

该目录用于存放 Reasonix 的按项目隔离的特性追踪数据（Project Feature Tracker）。

默认位置建议为：

    ~/.reasonix/project-tracker

如果设置了 REASONIX_HOME，则对应位置为：

    $REASONIX_HOME/project-tracker

推荐纳入版本管理：

- config.toml
- projects/**/project.toml
- projects/**/features.toml

推荐排除版本管理：

- projects/**/state.toml
`

const trackerGitignore = `projects/**/state.toml
`

const trackerConfigTemplate = `# Project Feature Tracker configuration.
# storage_dir = "~/.reasonix/project-tracker"
# base_dirs are matched against the current project root using the longest
# matching prefix. The relative path becomes the project tracker ID.
base_dirs = ["~/Projects"]
`

// Store persists per-project feature lists under the tracker root.
type Store struct {
	mu   sync.Mutex
	root string
	now  func() time.Time
}

// DefaultPath returns the tracker root under REASONIX_HOME / ~/.reasonix.
func DefaultPath(reasonixHome string) string {
	if strings.TrimSpace(reasonixHome) == "" {
		return ""
	}
	return filepath.Join(reasonixHome, "project-tracker")
}

// NewStore creates a project-tracker store rooted at path.
func NewStore(path string) *Store {
	return &Store{root: strings.TrimSpace(path), now: time.Now}
}

// Load returns the current project's tracker snapshot, creating the backing
// files on first use.
func (s *Store) Load(workspaceRoot string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(workspaceRoot)
}

// Save replaces the current project's features and returns the updated snapshot.
func (s *Store) Save(workspaceRoot string, features []Feature) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, loc, state, err := s.loadPreparedLocked(workspaceRoot)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Features = sanitizeFeatures(features)
	if err := atomicEncodeTOML(loc.featuresPath, featureFile{Features: snap.Features}); err != nil {
		return Snapshot{}, err
	}
	now := s.now().Format(time.RFC3339)
	if state.CreatedAt == "" {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	state.LastOpenedAt = now
	if err := atomicEncodeTOML(loc.statePath, state); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

type projectLocations struct {
	projectDir   string
	projectPath  string
	featuresPath string
	statePath    string
}

func (s *Store) loadLocked(workspaceRoot string) (Snapshot, error) {
	snap, _, state, err := s.loadPreparedLocked(workspaceRoot)
	if err != nil {
		return Snapshot{}, err
	}
	now := s.now().Format(time.RFC3339)
	if state.CreatedAt == "" {
		state.CreatedAt = now
	}
	if state.UpdatedAt == "" {
		state.UpdatedAt = state.CreatedAt
	}
	state.LastOpenedAt = now
	if err := atomicEncodeTOML(s.projectLocationsForSnapshot(snap).statePath, state); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

func (s *Store) loadPreparedLocked(workspaceRoot string) (Snapshot, projectLocations, State, error) {
	root := cleanPath(workspaceRoot)
	if root == "" {
		return Snapshot{}, projectLocations{}, State{}, errors.New("project tracker: workspace root not configured")
	}
	if s.root == "" {
		return Snapshot{}, projectLocations{}, State{}, errors.New("project tracker: storage root not configured")
	}
	if err := s.ensureTrackerRootLocked(); err != nil {
		return Snapshot{}, projectLocations{}, State{}, err
	}
	baseDirs := s.loadBaseDirsLocked()
	project := resolveProject(root, baseDirs)
	loc := s.projectLocations(project.ID)
	if err := os.MkdirAll(loc.projectDir, 0o755); err != nil {
		return Snapshot{}, projectLocations{}, State{}, err
	}
	if err := atomicEncodeTOML(loc.projectPath, project); err != nil {
		return Snapshot{}, projectLocations{}, State{}, err
	}
	ff := featureFile{}
	if _, err := toml.DecodeFile(loc.featuresPath, &ff); err != nil {
		if os.IsNotExist(err) {
			if err := atomicEncodeTOML(loc.featuresPath, featureFile{Features: []Feature{}}); err != nil {
				return Snapshot{}, projectLocations{}, State{}, err
			}
		} else {
			return Snapshot{}, projectLocations{}, State{}, err
		}
	}
	if len(ff.Features) == 0 {
		ff.Features = []Feature{}
	} else {
		ff.Features = sanitizeFeatures(ff.Features)
	}
	state := State{}
	if _, err := toml.DecodeFile(loc.statePath, &state); err != nil {
		if os.IsNotExist(err) {
			now := s.now().Format(time.RFC3339)
			state = State{CreatedAt: now, UpdatedAt: now, LastOpenedAt: now}
			if err := atomicEncodeTOML(loc.statePath, state); err != nil {
				return Snapshot{}, projectLocations{}, State{}, err
			}
		} else {
			return Snapshot{}, projectLocations{}, State{}, err
		}
	}
	return Snapshot{Project: project, Features: ff.Features}, loc, state, nil
}

func (s *Store) projectLocationsForSnapshot(snap Snapshot) projectLocations {
	return s.projectLocations(snap.Project.ID)
}

func (s *Store) projectLocations(projectID string) projectLocations {
	projectDir := filepath.Join(s.root, "projects", filepath.FromSlash(projectID))
	return projectLocations{
		projectDir:   projectDir,
		projectPath:  filepath.Join(projectDir, "project.toml"),
		featuresPath: filepath.Join(projectDir, "features.toml"),
		statePath:    filepath.Join(projectDir, "state.toml"),
	}
}

func (s *Store) ensureTrackerRootLocked() error {
	if err := os.MkdirAll(filepath.Join(s.root, "projects"), 0o755); err != nil {
		return err
	}
	if err := ensureTextFile(filepath.Join(s.root, "README.md"), trackerReadme); err != nil {
		return err
	}
	if err := ensureTextFile(filepath.Join(s.root, ".gitignore"), trackerGitignore); err != nil {
		return err
	}
	return ensureTextFile(filepath.Join(s.root, "config.toml"), trackerConfigTemplate)
}

func (s *Store) loadBaseDirsLocked() []string {
	cfgPath := filepath.Join(s.root, "config.toml")
	cfg := configFile{}
	if _, err := toml.DecodeFile(cfgPath, &cfg); err != nil || len(cfg.BaseDirs) == 0 {
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

func resolveProject(workspaceRoot string, baseDirs []string) Project {
	root := cleanPath(workspaceRoot)
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
	}
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

func sanitizeFeatures(features []Feature) []Feature {
	out := make([]Feature, 0, len(features))
	used := map[string]int{}
	for _, feature := range features {
		title := strings.TrimSpace(feature.Title)
		if title == "" {
			continue
		}
		id := slugify(strings.TrimSpace(feature.ID))
		if id == "" {
			id = slugify(title)
		}
		if id == "" {
			id = "feature"
		}
		count := used[id]
		used[id] = count + 1
		if count > 0 {
			id = fmt.Sprintf("%s-%d", id, count+1)
		}
		status := strings.ToLower(strings.TrimSpace(feature.Status))
		if _, ok := statusRank[status]; !ok {
			status = "wishlist"
		}
		priority := strings.ToLower(strings.TrimSpace(feature.Priority))
		if _, ok := priorityRank[priority]; !ok {
			priority = "medium"
		}
		out = append(out, Feature{
			ID:       id,
			Title:    title,
			Status:   status,
			Priority: priority,
			Summary:  strings.TrimSpace(feature.Summary),
			Notes:    strings.TrimSpace(feature.Notes),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if statusRank[a.Status] != statusRank[b.Status] {
			return statusRank[a.Status] < statusRank[b.Status]
		}
		if priorityRank[a.Priority] != priorityRank[b.Priority] {
			return priorityRank[a.Priority] < priorityRank[b.Priority]
		}
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
	return out
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || unicode.IsSpace(r) || r == '/' || r == '.':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
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

func atomicEncodeTOML(path string, v any) error {
	b, err := toml.Marshal(v)
	if err != nil {
		return err
	}
	return atomicWrite(path, b)
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

func indexMap(items []string) map[string]int {
	out := make(map[string]int, len(items))
	for i, item := range items {
		out[item] = i
	}
	return out
}
