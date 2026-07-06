package quickreply

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestManagerLoadMissingFileReturnsEmptySlice(t *testing.T) {
	m := NewManager(filepath.Join(t.TempDir(), "quick-replies.toml"))
	got := m.Load()
	if got == nil {
		t.Fatal("Load returned nil slice")
	}
	if len(got) != 0 {
		t.Fatalf("Load length = %d, want 0", len(got))
	}
}

func TestManagerSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quick-replies.toml")
	m := NewManager(path)
	want := []QuickReply{
		{Name: "Ack", Body: "On it"},
		{Name: "Draft", Body: "Let me think"},
	}

	if err := m.Save(want); err != nil {
		t.Fatal(err)
	}
	got := m.Load()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load = %#v, want %#v", got, want)
	}
}

func TestManagerLoadMigratesLegacyFieldsByIgnoringThem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quick-replies.toml")
	if err := os.WriteFile(path, []byte(`
[[quick_replies]]
name = "Ack"
body = "On it"
auto_send = true
icon = "+"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(path)
	got := m.Load()
	want := []QuickReply{{Name: "Ack", Body: "On it"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load = %#v, want %#v", got, want)
	}
}

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/tmp/reasonix")
	want := filepath.Join("/tmp/reasonix", "quick-replies.toml")
	if got != want {
		t.Fatalf("DefaultPath = %q, want %q", got, want)
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

func TestManagerSaveLoadRoundTripWithCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quick-replies.toml")
	m := NewManager(path)
	want := []QuickReply{
		{Name: "Req check", Body: "Have we confirmed requirements?", Category: "requirements"},
		{Name: "Bug bash", Body: "Let's do a bug bash", Category: "testing"},
		{Name: "No category", Body: "General note"},
	}

	if err := m.Save(want); err != nil {
		t.Fatal(err)
	}
	got := m.Load()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load = %#v, want %#v", got, want)
	}
}

func TestSanitizeRejectsInvalidCategory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quick-replies.toml")
	m := NewManager(path)
	replies := []QuickReply{
		{Name: "Valid", Body: "ok", Category: "requirements"},
		{Name: "Invalid cat", Body: "bad", Category: "hacker_attack"},
		{Name: "Empty cat", Body: "fine"},
	}

	if err := m.Save(replies); err != nil {
		t.Fatal(err)
	}
	got := m.Load()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Category != "requirements" {
		t.Errorf("got[0].Category = %q, want 'requirements'", got[0].Category)
	}
	if got[1].Category != "" {
		t.Errorf("got[1].Category = %q, want '' (rejected)", got[1].Category)
	}
	if got[2].Category != "" {
		t.Errorf("got[2].Category = %q, want ''", got[2].Category)
	}
}
