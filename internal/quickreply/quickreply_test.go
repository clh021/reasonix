package quickreply

import (
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
		{Name: "Ack", Body: "On it", AutoSend: true, Icon: "+"},
		{Name: "Draft", Body: "Let me think", AutoSend: false},
	}

	if err := m.Save(want); err != nil {
		t.Fatal(err)
	}
	got := m.Load()
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
