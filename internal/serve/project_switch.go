package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"reasonix/internal/agent"
	"reasonix/internal/boot"
	"reasonix/internal/config"
)

// switchProject rebuilds the controller for a different workspace/project root.
// Unlike switchModel, it does NOT carry the conversation — switching project
// starts a fresh session in the new project's context. The heavy work runs
// off s.mu; bindMu serializes it against the other session-binding entry
// points so the controller and session lease move together.
func (s *Server) switchProject(ctx context.Context, root string) error {
	s.bindMu.Lock()
	defer s.bindMu.Unlock()

	cur := s.ctl()
	if cur.Running() {
		return fmt.Errorf("cannot switch project while a turn is running")
	}
	// Snapshot the current session so it's saved in the old project's session dir.
	if err := cur.Snapshot(); err != nil {
		slog.Warn("serve: snapshot before project switch", "err", err)
	}
	modelRef := cur.Label()

	sessionDir := config.ProjectSessionDir(root)
	newCtrl, err := boot.Build(ctx, boot.Options{
		Model:         modelRef,
		Sink:          s.bc,
		Stderr:        os.Stderr,
		WorkspaceRoot: root,
		SessionDir:    sessionDir,
	})
	if err != nil {
		return fmt.Errorf("switch project: %w", err)
	}
	// Start a fresh session in the new project.
	newCtrl.SetSessionPath(agent.NewSessionPath(newCtrl.SessionDir(), newCtrl.Label()))

	s.mu.Lock()
	if s.ctrl != cur {
		s.mu.Unlock()
		newCtrl.Close()
		return fmt.Errorf("switch project: session changed during switch")
	}
	s.prepareReplacementController(cur, newCtrl)
	s.ctrl = newCtrl
	s.mu.Unlock()

	if err := s.rebindSessionLease(newCtrl.SessionPath()); err != nil {
		slog.Warn("serve: session lease after project switch", "err", err)
	}
	rememberWorkspace(root)
	cur.Close()
	s.initTitleProvider()
	return nil
}

type replacementController interface {
	EnableInteractiveApproval()
	PlanMode() bool
	SetPlanMode(bool)
	ToolApprovalMode() string
	SetToolApprovalMode(string)
	SessionDir() string
}

func (s *Server) prepareReplacementController(cur, newCtrl replacementController) {
	if s.interactiveApproval {
		newCtrl.EnableInteractiveApproval()
	}
	newCtrl.SetPlanMode(cur.PlanMode())
	newCtrl.SetToolApprovalMode(cur.ToolApprovalMode())
	s.titles = newTitleCache(newCtrl.SessionDir())
}

func (s *Server) enableInteractiveApproval() {
	s.mu.Lock()
	s.interactiveApproval = true
	ctrl := s.ctrl
	s.mu.Unlock()
	ctrl.EnableInteractiveApproval()
}

// workspaces returns the list of remembered workspace directories.
func (s *Server) workspaces(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	cur := s.ctrl.WorkspaceRoot()
	s.mu.RUnlock()

	type wsEntry struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	paths := loadWorkspaces()
	out := make([]wsEntry, 0, len(paths))
	for _, p := range paths {
		out = append(out, wsEntry{Path: p, Name: filepath.Base(p)})
	}
	writeJSON(w, map[string]any{
		"workspaces": out,
		"current":    cur,
	})
}

// switchProjectHandler handles POST /switch-project: receives a workspace root
// path and replaces the current controller with one serving that project.
func (s *Server) switchProjectHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	absPath, err := filepath.Abs(body.Path)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		http.Error(w, "path must be an existing directory", http.StatusBadRequest)
		return
	}
	if absPath == s.ctl().WorkspaceRoot() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := s.switchProject(r.Context(), absPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"workspaceRoot": absPath})
}
