package quickreply

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestManagerLoadMissingFileReturnsEmptySlice(t *testing.T) {
	m := NewManager(filepath.Join(t.TempDir(), "quick-replies.toml"), filepath.Join(t.TempDir(), "project-quick-replies"))
	got, err := m.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Replies == nil {
		t.Fatal("Load returned nil slice")
	}
	if len(got.Replies) != 0 {
		t.Fatalf("Load length = %d, want 0", len(got.Replies))
	}
}

func TestManagerSaveLoadRoundTripPublicOnly(t *testing.T) {
	home := t.TempDir()
	m := NewManager(filepath.Join(home, "quick-replies.toml"), filepath.Join(home, "project-quick-replies"))
	want := []QuickReply{
		{Name: "Ack", Body: "On it", Scope: ScopePublic},
		{Name: "Draft", Body: "Let me think", Scope: ScopePublic},
	}

	if _, err := m.Save("", "", want); err != nil {
		t.Fatal(err)
	}
	got, err := m.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("Load = %#v, want %#v", got.Replies, want)
	}
}

func TestManagerLoadMigratesLegacyFieldsByIgnoringThem(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "quick-replies.toml")
	if err := os.WriteFile(path, []byte(`
[[quick_replies]]
name = "Ack"
body = "On it"
auto_send = true
icon = "+"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(path, filepath.Join(home, "project-quick-replies"))
	got, err := m.Load("")
	if err != nil {
		t.Fatal(err)
	}
	want := []QuickReply{{Name: "Ack", Body: "On it", Scope: ScopePublic}}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("Load = %#v, want %#v", got.Replies, want)
	}
}

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/tmp/reasonix")
	want := filepath.Join("/tmp/reasonix", "quick-replies.toml")
	if got != want {
		t.Fatalf("DefaultPath = %q, want %q", got, want)
	}
}

func TestDefaultProjectPath(t *testing.T) {
	got := DefaultProjectPath("/tmp/reasonix")
	want := filepath.Join("/tmp/reasonix", "project-quick-replies")
	if got != want {
		t.Fatalf("DefaultProjectPath = %q, want %q", got, want)
	}
}

func TestDefaultCategoriesReturnsFivePhases(t *testing.T) {
	got := DefaultCategories()
	if len(got) != 5 {
		t.Fatalf("DefaultCategories length = %d, want 5", len(got))
	}
	ids := make(map[string]bool)
	for _, c := range got {
		if c.ID == "" {
			t.Error("category with empty ID")
		}
		if c.NameZh == "" {
			t.Errorf("category %q missing NameZh", c.ID)
		}
		if c.NameEn == "" {
			t.Errorf("category %q missing NameEn", c.ID)
		}
		if ids[c.ID] {
			t.Errorf("duplicate category ID: %q", c.ID)
		}
		ids[c.ID] = true
	}
}

func TestValidCategoryID(t *testing.T) {
	if !ValidCategoryID("requirements") {
		t.Error("ValidCategoryID('requirements') = false, want true")
	}
	if !ValidCategoryID("design") {
		t.Error("ValidCategoryID('design') = false, want true")
	}
	if !ValidCategoryID("development") {
		t.Error("ValidCategoryID('development') = false, want true")
	}
	if !ValidCategoryID("testing") {
		t.Error("ValidCategoryID('testing') = false, want true")
	}
	if !ValidCategoryID("deployment") {
		t.Error("ValidCategoryID('deployment') = false, want true")
	}
	if ValidCategoryID("bogus") {
		t.Error("ValidCategoryID('bogus') = true, want false")
	}
	if !ValidCategoryID("") {
		t.Error("ValidCategoryID('') = false, want true (empty is allowed)")
	}
}

func TestManagerSaveLoadRoundTripWithCategoriesAndScopes(t *testing.T) {
	home := t.TempDir()
	workspaceRoot := filepath.Join(home, "Projects", "reasonix")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(home, "project-quick-replies")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(filepath.Join(home, "Projects"))+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(filepath.Join(home, "quick-replies.toml"), projectRoot)
	want := []QuickReply{
		{Name: "Req check", Body: "Have we confirmed requirements?", Category: "requirements", Scope: ScopePublic},
		{Name: "Bug bash", Body: "Let's do a bug bash", Category: "testing", Scope: ScopeProject},
		{Name: "No category", Body: "General note", Scope: ScopeProject},
	}

	if _, err := m.Save(workspaceRoot, workspaceRoot, want); err != nil {
		t.Fatal(err)
	}
	got, err := m.Load(workspaceRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("Load = %#v, want %#v", got.Replies, want)
	}
	if got.Project.ID != "reasonix" {
		t.Fatalf("Project.ID = %q, want reasonix", got.Project.ID)
	}
}

func TestSanitizeRejectsInvalidCategoryAndScope(t *testing.T) {
	home := t.TempDir()
	m := NewManager(filepath.Join(home, "quick-replies.toml"), filepath.Join(home, "project-quick-replies"))
	replies := []QuickReply{
		{Name: "Valid", Body: "ok", Category: "requirements", Scope: ScopePublic},
		{Name: "Invalid cat", Body: "bad", Category: "hacker_attack", Scope: ScopePublic},
		{Name: "Invalid scope", Body: "fallback", Scope: "shared"},
	}

	if _, err := m.Save("", "", replies); err != nil {
		t.Fatal(err)
	}
	got, err := m.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Replies) != 3 {
		t.Fatalf("len = %d, want 3", len(got.Replies))
	}
	if got.Replies[0].Category != "requirements" {
		t.Errorf("got[0].Category = %q, want 'requirements'", got.Replies[0].Category)
	}
	if got.Replies[1].Category != "" {
		t.Errorf("got[1].Category = %q, want ''", got.Replies[1].Category)
	}
	if got.Replies[1].Scope != ScopePublic {
		t.Errorf("got[1].Scope = %q, want %q", got.Replies[1].Scope, ScopePublic)
	}
	if got.Replies[2].Scope != ScopePublic {
		t.Errorf("got[2].Scope = %q, want %q", got.Replies[2].Scope, ScopePublic)
	}
}

func TestManagerProjectIsolation(t *testing.T) {
	home := t.TempDir()
	projectsBase := filepath.Join(home, "Projects")
	workspaceA := filepath.Join(projectsBase, "reasonix")
	workspaceB := filepath.Join(projectsBase, "yak")
	for _, dir := range []string{workspaceA, workspaceB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	projectRoot := filepath.Join(home, "project-quick-replies")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(projectsBase)+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(filepath.Join(home, "quick-replies.toml"), projectRoot)
	if _, err := m.Save(workspaceA, workspaceA, []QuickReply{
		{Name: "Shared", Body: "For everyone", Scope: ScopePublic},
		{Name: "Only A", Body: "For project A", Scope: ScopeProject},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Save(workspaceB, workspaceB, []QuickReply{
		{Name: "Shared", Body: "For everyone", Scope: ScopePublic},
		{Name: "Only B", Body: "For project B", Scope: ScopeProject},
	}); err != nil {
		t.Fatal(err)
	}

	gotA, err := m.Load(workspaceA)
	if err != nil {
		t.Fatal(err)
	}
	wantA := []QuickReply{
		{Name: "Shared", Body: "For everyone", Scope: ScopePublic},
		{Name: "Only A", Body: "For project A", Scope: ScopeProject},
	}
	if !reflect.DeepEqual(gotA.Replies, wantA) {
		t.Fatalf("project A replies = %#v, want %#v", gotA.Replies, wantA)
	}

	gotB, err := m.Load(workspaceB)
	if err != nil {
		t.Fatal(err)
	}
	wantB := []QuickReply{
		{Name: "Shared", Body: "For everyone", Scope: ScopePublic},
		{Name: "Only B", Body: "For project B", Scope: ScopeProject},
	}
	if !reflect.DeepEqual(gotB.Replies, wantB) {
		t.Fatalf("project B replies = %#v, want %#v", gotB.Replies, wantB)
	}
}

func TestManagerSaveRejectsStaleProjectContextWithoutWritingPublicReplies(t *testing.T) {
	home := t.TempDir()
	projectsBase := filepath.Join(home, "Projects")
	workspaceA := filepath.Join(projectsBase, "reasonix")
	workspaceB := filepath.Join(projectsBase, "yak")
	for _, dir := range []string{workspaceA, workspaceB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	projectRoot := filepath.Join(home, "project-quick-replies")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(projectsBase)+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(filepath.Join(home, "quick-replies.toml"), projectRoot)
	if _, err := m.Save(workspaceA, workspaceA, []QuickReply{
		{Name: "Shared", Body: "Before", Scope: ScopePublic},
		{Name: "Only A", Body: "For project A", Scope: ScopeProject},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := m.Save(workspaceB, workspaceA, []QuickReply{
		{Name: "Shared", Body: "After", Scope: ScopePublic},
		{Name: "Only A", Body: "For project A", Scope: ScopeProject},
	})
	if !errors.Is(err, ErrProjectContextChanged) {
		t.Fatalf("Save stale context err = %v, want %v", err, ErrProjectContextChanged)
	}

	got, err := m.Load(workspaceA)
	if err != nil {
		t.Fatal(err)
	}
	want := []QuickReply{
		{Name: "Shared", Body: "Before", Scope: ScopePublic},
		{Name: "Only A", Body: "For project A", Scope: ScopeProject},
	}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("replies after rejected save = %#v, want %#v", got.Replies, want)
	}
}

func TestManagerSaveDoesNotPartiallyWriteWhenProjectScopeNeedsWorkspace(t *testing.T) {
	home := t.TempDir()
	m := NewManager(filepath.Join(home, "quick-replies.toml"), filepath.Join(home, "project-quick-replies"))
	if _, err := m.Save("", "", []QuickReply{{Name: "Shared", Body: "Before", Scope: ScopePublic}}); err != nil {
		t.Fatal(err)
	}

	_, err := m.Save("", "", []QuickReply{
		{Name: "Shared", Body: "After", Scope: ScopePublic},
		{Name: "Project", Body: "Needs workspace", Scope: ScopeProject},
	})
	if !errors.Is(err, ErrWorkspaceNotConfigured) {
		t.Fatalf("Save missing workspace err = %v, want %v", err, ErrWorkspaceNotConfigured)
	}

	got, err := m.Load("")
	if err != nil {
		t.Fatal(err)
	}
	want := []QuickReply{{Name: "Shared", Body: "Before", Scope: ScopePublic}}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("public replies after rejected save = %#v, want %#v", got.Replies, want)
	}
}

func TestManagerLegacyPublicSaveDoesNotClearProjectReplies(t *testing.T) {
	home := t.TempDir()
	projectsBase := filepath.Join(home, "Projects")
	workspace := filepath.Join(projectsBase, "reasonix")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(home, "project-quick-replies")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "config.toml"), []byte(`base_dirs = ["`+filepath.ToSlash(projectsBase)+`"]`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(filepath.Join(home, "quick-replies.toml"), projectRoot)
	if _, err := m.Save(workspace, workspace, []QuickReply{
		{Name: "Shared", Body: "Before", Scope: ScopePublic},
		{Name: "Project", Body: "Keep me", Scope: ScopeProject},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Save(workspace, "", []QuickReply{{Name: "Shared", Body: "After", Scope: ScopePublic}}); err != nil {
		t.Fatal(err)
	}

	got, err := m.Load(workspace)
	if err != nil {
		t.Fatal(err)
	}
	want := []QuickReply{
		{Name: "Shared", Body: "After", Scope: ScopePublic},
		{Name: "Project", Body: "Keep me", Scope: ScopeProject},
	}
	if !reflect.DeepEqual(got.Replies, want) {
		t.Fatalf("replies after legacy public save = %#v, want %#v", got.Replies, want)
	}
}
