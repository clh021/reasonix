package projecttracker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/tmp/reasonix")
	want := filepath.Join("/tmp/reasonix", "project-tracker")
	if got != want {
		t.Fatalf("DefaultPath = %q, want %q", got, want)
	}
}

func TestStoreLoadCreatesTrackerFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "tracker")
	store := NewStore(root)
	store.now = func() time.Time { return time.Date(2026, 7, 6, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)) }
	workspace := filepath.Join(t.TempDir(), "Projects", "reasonix")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(filepath.Dir(workspace))+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	snap, err := store.Load(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Project.ID != "reasonix" {
		t.Fatalf("project id = %q, want reasonix", snap.Project.ID)
	}
	for _, path := range []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, ".gitignore"),
		filepath.Join(root, "config.toml"),
		filepath.Join(root, "projects", "reasonix", "project.toml"),
		filepath.Join(root, "projects", "reasonix", "features.toml"),
		filepath.Join(root, "projects", "reasonix", "state.toml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	root := filepath.Join(t.TempDir(), "tracker")
	store := NewStore(root)
	workspace := filepath.Join(t.TempDir(), "Projects", "reasonix", "plugins", "example")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(filepath.Dir(filepath.Dir(filepath.Dir(workspace))))+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	features := []Feature{
		{Title: "Active", Status: "custom_active", Priority: "medium"},
		{Title: "Wishlist", Status: "wishlist", Priority: "high", Notes: "first"},
		{Title: "No Status"},
	}
	if _, err := store.Save(workspace, features); err != nil {
		t.Fatal(err)
	}
	snap, err := store.Load(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Project.ID != "reasonix/plugins/example" {
		t.Fatalf("project id = %q, want reasonix/plugins/example", snap.Project.ID)
	}
	if len(snap.Features) != 3 {
		t.Fatalf("feature count = %d, want 3", len(snap.Features))
	}
	if snap.Features[0].Status != "wishlist" {
		t.Fatalf("first feature status = %q, want wishlist sorted first", snap.Features[0].Status)
	}
	if snap.Features[1].Status != "wishlist" {
		t.Fatalf("second feature should default to wishlist, got %q", snap.Features[1].Status)
	}
}

func TestStoreFallbackExternalProjectID(t *testing.T) {
	root := filepath.Join(t.TempDir(), "tracker")
	store := NewStore(root)
	workspace := filepath.Join(t.TempDir(), "elsewhere", "demo")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	snap, err := store.Load(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(snap.Project.ID, "external/") {
		t.Fatalf("project id = %q, want external/* fallback", snap.Project.ID)
	}
}
