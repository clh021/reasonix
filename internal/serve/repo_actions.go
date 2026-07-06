package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	repoStatusTimeout = 5 * time.Second
	repoPushTimeout   = 30 * time.Second
)

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
		resp, err := runRepoPush(r.Context(), s.ctl().WorkspaceRoot())
		if err != nil {
			writeJSON(w, resp)
			return
		}
		writeJSON(w, resp)
	default:
		http.Error(w, "unsupported action", http.StatusBadRequest)
	}
}

func runRepoStatus(parent context.Context, workspace string) (repoActionResponse, error) {
	ctx, cancel := context.WithTimeout(parent, repoStatusTimeout)
	defer cancel()

	resp, err := resolveRepoContext(ctx, "status", workspace)
	if err != nil {
		return repoActionResponse{}, err
	}

	statusOut, err := runGitForRepoAction(ctx, resp.RepoRoot, "status", "--short", "--branch")
	if err != nil {
		return repoActionResponse{}, errors.New("failed to read git status")
	}
	resp.Output = strings.TrimRight(statusOut, "\n")
	if resp.Output == "" {
		resp.Output = "## " + filepath.Base(resp.RepoRoot)
	}
	return resp, nil
}

func runRepoPush(parent context.Context, workspace string) (repoActionResponse, error) {
	ctx, cancel := context.WithTimeout(parent, repoPushTimeout)
	defer cancel()

	resp, err := resolveRepoContext(ctx, "push", workspace)
	if err != nil {
		return repoActionResponse{
			Action:    "push",
			Workspace: workspace,
			Error:     err.Error(),
		}, err
	}

	out, err := runGitForRepoAction(ctx, resp.RepoRoot, "push")
	resp.Output = strings.TrimRight(out, "\n")
	if err != nil {
		if resp.Output == "" {
			resp.Error = "git push failed"
		} else {
			resp.Error = fmt.Sprintf("git push failed: %s", firstLine(resp.Output))
		}
		return resp, err
	}
	if resp.Output == "" {
		resp.Output = "git push completed"
	}
	return resp, nil
}

func resolveRepoContext(ctx context.Context, action string, workspace string) (repoActionResponse, error) {
	root, err := runGitForRepoAction(ctx, workspace, "rev-parse", "--show-toplevel")
	if err != nil {
		return repoActionResponse{}, errors.New("current project is not a git repository")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return repoActionResponse{}, errors.New("failed to resolve git repository root")
	}

	resp := repoActionResponse{
		Action:    action,
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
	return resp, nil
}

func runGitForRepoAction(ctx context.Context, cwd string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = append(os.Environ(),
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_TERMINAL_PROMPT=0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func firstLine(s string) string {
	line := strings.TrimSpace(s)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}
