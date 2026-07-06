package projecttracker

// Feature is a single tracked capability for a project.
type Feature struct {
	ID       string `json:"id" toml:"id"`
	Title    string `json:"title" toml:"title"`
	Status   string `json:"status" toml:"status"`
	Priority string `json:"priority" toml:"priority"`
	Summary  string `json:"summary,omitempty" toml:"summary,omitempty"`
	Notes    string `json:"notes,omitempty" toml:"notes,omitempty"`
}

// Project identifies the current project as resolved by the tracker store.
type Project struct {
	ID      string `json:"id" toml:"id"`
	Name    string `json:"name" toml:"name"`
	Root    string `json:"root" toml:"root"`
	BaseDir string `json:"baseDir,omitempty" toml:"base_dir,omitempty"`
}

// State stores high-churn runtime metadata separately from versioned files.
type State struct {
	CreatedAt    string `toml:"created_at"`
	UpdatedAt    string `toml:"updated_at"`
	LastOpenedAt string `toml:"last_opened_at"`
}

type configFile struct {
	BaseDirs []string `toml:"base_dirs"`
}

type featureFile struct {
	Features []Feature `toml:"features"`
}

// Snapshot is the JSON payload served to the web UI.
type Snapshot struct {
	Project  Project   `json:"project"`
	Features []Feature `json:"features"`
}
