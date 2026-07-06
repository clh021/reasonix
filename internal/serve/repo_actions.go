package serve

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const repoActionTimeout = 5 * time.Second

type repoActionResponse struct {
	Action    string `json:"action"`
	Workspace string `json:"workspace"`
	RepoRoot  string `json:"repoRoot,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Detached  bool   `json:"detached,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *Server) repoAction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Action) == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(body.Action)
	switch action {
	case "status":
		resp, err := runRepoStatus(r.Context(), s.ctl().WorkspaceRoot())
		if err != nil {
			writeJSON(w, repoActionResponse{
				Action:    action,
				Workspace: s.ctl().WorkspaceRoot(),
				Error:     err.Error(),
			})
			return
		}
		writeJSON(w, resp)
	case "push":
		writeJSON(w, repoActionResponse{
			Action:    action,
			Workspace: s.ctl().WorkspaceRoot(),
			Error:     "git push is not enabled yet; use git status first and keep push in the agent flow for now",
		})
	default:
		http.Error(w, "unsupported action", http.StatusBadRequest)
	}
}

func runRepoStatus(parent context.Context, workspace string) (repoActionResponse, error) {
	ctx, cancel := context.WithTimeout(parent, repoActionTimeout)
	defer cancel()

	root, err := runGitForRepoAction(ctx, workspace, "rev-parse", "--show-toplevel")
	if err != nil {
		return repoActionResponse{}, errors.New("current project is not a git repository")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return repoActionResponse{}, errors.New("failed to resolve git repository root")
	}

	resp := repoActionResponse{
		Action:    "status",
		Workspace: workspace,
		RepoRoot:  root,
	}

	if branch, err := runGitForRepoAction(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD"); err == nil && strings.TrimSpace(branch) != "" {
		resp.Branch = strings.TrimSpace(branch)
	} else if sha, err := runGitForRepoAction(ctx, root, "rev-parse", "--short", "HEAD"); err == nil && strings.TrimSpace(sha) != "" {
		resp.Branch = strings.TrimSpace(sha)
		resp.Detached = true
	} else {
		resp.Branch = "HEAD"
		resp.Detached = true
	}

	statusOut, err := runGitForRepoAction(ctx, root, "status", "--short", "--branch")
	if err != nil {
		return repoActionResponse{}, errors.New("failed to read git status")
	}
	resp.Output = strings.TrimRight(statusOut, "\n")
	if resp.Output == "" {
		resp.Output = "## " + filepath.Base(root)
	}
	return resp, nil
}

func runGitForRepoAction(ctx context.Context, cwd string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
