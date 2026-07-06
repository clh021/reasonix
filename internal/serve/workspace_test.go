package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestWorkspaceListPath(t *testing.T) {
	p := workspaceListPath()
	if p == "" {
		t.Skip("MemoryUserDir is empty — no user config dir available")
	}
	if !filepath.IsAbs(p) {
		t.Errorf("workspaceListPath() = %q, want absolute path", p)
	}
	if filepath.Base(p) != "desktop-workspaces.json" {
		t.Errorf("workspaceListPath() = %q, want .../desktop-workspaces.json", p)
	}
}

func TestRememberWorkspace(t *testing.T) {
	orig := workspaceListPath
	defer func() { workspaceListPath = orig }()

	tmpDir := t.TempDir()
	workspaceListPath = func() string { return filepath.Join(tmpDir, "workspaces.json") }

	// Initially empty
	ws := loadWorkspaces()
	if ws != nil {
		t.Fatalf("loadWorkspaces on empty file = %v, want nil", ws)
	}

	// Remember a path
	rememberWorkspace("/tmp/test-project")
	ws = loadWorkspaces()
	if len(ws) != 1 || ws[0] != "/tmp/test-project" {
		t.Fatalf("after first remember: got %v, want [/tmp/test-project]", ws)
	}

	// Remember another
	rememberWorkspace("/home/user/other-project")
	ws = loadWorkspaces()
	if len(ws) != 2 {
		t.Fatalf("after second remember: len=%d, want 2; got %v", len(ws), ws)
	}
	if ws[0] != "/home/user/other-project" {
		t.Errorf("first entry should be newest; got %q", ws[0])
	}

	// Duplicate — should not add
	rememberWorkspace("/tmp/test-project")
	ws = loadWorkspaces()
	if len(ws) != 2 {
		t.Fatalf("dedup check: len=%d, want 2; got %v", len(ws), ws)
	}

	// Empty path — no-op
	rememberWorkspace("")
	ws = loadWorkspaces()
	if len(ws) != 2 {
		t.Fatalf("empty path add: len=%d, want 2; got %v", len(ws), ws)
	}

	// 12-entry cap
	for i := range 15 {
		rememberWorkspace(filepath.Join("/proj", fmt.Sprintf("test-%d", i)))
	}
	ws = loadWorkspaces()
	if len(ws) > 12 {
		t.Fatalf("cap check: len=%d, want <=12", len(ws))
	}
	if len(ws) != 12 {
		t.Errorf("cap check: len=%d, want exactly 12", len(ws))
	}
	// Newest first: the most recent entry appears at position 0
	if len(ws) > 0 && ws[0] != filepath.Join("/proj", "test-14") {
		t.Errorf("newest first: got %q, want /proj/test-14", ws[0])
	}
}

func TestWorkspacesEndpoint(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/workspaces")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/workspaces status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Workspaces []struct {
			Path string `json:"path"`
			Name string `json:"name"`
		} `json:"workspaces"`
		Current string `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode /workspaces response: %v", err)
	}
	if body.Current == "" {
		t.Logf("/workspaces current is empty (expected in test without real controller)")
	}
}

func TestSwitchProjectHandler_Validation(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()
	url := srv.URL + "/switch-project"

	t.Run("missing path", func(t *testing.T) {
		resp, err := http.Post(url, "application/json", strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("empty body status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		body := `{"path":"/nonexistent/path/that/does/not/exist"}`
		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("non-existent path status = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("path is a file not a directory", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "not-a-dir")
		if err := os.WriteFile(tmpFile, []byte("hi"), 0o644); err != nil {
			t.Fatal(err)
		}
		body := `{"path":"` + tmpFile + `"}`
		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("file path status = %d, want 400", resp.StatusCode)
		}
	})

	// TestSwitchProjectHandler_SamePath is intentionally absent: testing the
	// same-path short-circuit (serve.go:793) requires a controller with a real
	// WorkspaceRoot, which means boot.Build + config resolution — available
	// only in integration tests, not here.
}

// TestSwitchProjectHandler_ValidDir verifies that a valid directory passes
// pre-flight validation and POST /switch-project returns 200 with the new
// workspace root. Requires a configured model in the user config (boot.Build
// fallback); skips gracefully when no config is available (e.g. clean CI).
func TestSwitchProjectHandler_ValidDir(t *testing.T) {
	if _, err := config.Load(); err != nil {
		t.Skipf("no user config available, skipping: %v", err)
	}

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	validDir := t.TempDir()
	body := `{"path":"` + validDir + `"}`
	resp, err := http.Post(srv.URL+"/switch-project", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid dir status = %d, want 200", resp.StatusCode)
	}
	var result struct {
		WorkspaceRoot string `json:"workspaceRoot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.WorkspaceRoot != validDir {
		t.Errorf("workspaceRoot = %q, want %q", result.WorkspaceRoot, validDir)
	}
}
