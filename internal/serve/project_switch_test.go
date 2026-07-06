package serve

import (
	"testing"

	"reasonix/internal/control"
)

type replacementControllerStub struct {
	planMode              bool
	toolApprovalMode      string
	sessionDir            string
	enableInteractiveHits int
}

func (s *replacementControllerStub) EnableInteractiveApproval() { s.enableInteractiveHits++ }
func (s *replacementControllerStub) PlanMode() bool             { return s.planMode }
func (s *replacementControllerStub) SetPlanMode(v bool)         { s.planMode = v }
func (s *replacementControllerStub) ToolApprovalMode() string   { return s.toolApprovalMode }
func (s *replacementControllerStub) SetToolApprovalMode(v string) {
	s.toolApprovalMode = v
}
func (s *replacementControllerStub) SessionDir() string { return s.sessionDir }

func TestPrepareReplacementControllerCarriesRuntimeModes(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	srv := &Server{
		titles:              newTitleCache(oldDir),
		interactiveApproval: true,
	}
	cur := &replacementControllerStub{
		planMode:         true,
		toolApprovalMode: control.ToolApprovalYolo,
		sessionDir:       oldDir,
	}
	next := &replacementControllerStub{
		toolApprovalMode: control.ToolApprovalAsk,
		sessionDir:       newDir,
	}

	srv.prepareReplacementController(cur, next)

	if !next.planMode {
		t.Fatal("plan mode was not carried to replacement controller")
	}
	if got := next.toolApprovalMode; got != control.ToolApprovalYolo {
		t.Fatalf("tool approval mode = %q, want %q", got, control.ToolApprovalYolo)
	}
	if next.enableInteractiveHits != 1 {
		t.Fatalf("interactive approval enabled %d times, want 1", next.enableInteractiveHits)
	}
	if srv.titles == nil || srv.titles.dir != newDir {
		t.Fatalf("title cache dir = %q, want %q", srv.titles.dir, newDir)
	}
}

func TestPrepareReplacementControllerSkipsInteractiveApprovalWhenDisabled(t *testing.T) {
	srv := &Server{titles: newTitleCache(t.TempDir())}
	cur := &replacementControllerStub{
		toolApprovalMode: control.ToolApprovalAuto,
		sessionDir:       t.TempDir(),
	}
	next := &replacementControllerStub{
		sessionDir: t.TempDir(),
	}

	srv.prepareReplacementController(cur, next)

	if next.enableInteractiveHits != 0 {
		t.Fatalf("interactive approval enabled %d times, want 0", next.enableInteractiveHits)
	}
	if got := next.toolApprovalMode; got != control.ToolApprovalAuto {
		t.Fatalf("tool approval mode = %q, want %q", got, control.ToolApprovalAuto)
	}
}
