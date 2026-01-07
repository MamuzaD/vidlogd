package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepo struct {
	Dir    string
	Remote string
}

func NewGitRepo(dir string) *GitRepo {
	return &GitRepo{Dir: dir}
}

func (r *GitRepo) cmd(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	return cmd
}

func (r *GitRepo) run(args ...string) error {
	cmd := r.cmd(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return nil
}

func (r *GitRepo) output(args ...string) (string, error) {
	cmd := r.cmd(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *GitRepo) Init(remote string) error {
	if err := r.run("init", "-b", "main"); err != nil {
		// fallback to default
		if err := r.run("init"); err != nil {
			return err
		}

		// ensure branch "main" exists
		_ = r.run("checkout", "-b", "main")
	}

	if remote != "" {
		r.run("remote", "add", "origin", remote)
		r.Remote = remote
	}
	return nil
}

func (r *GitRepo) Clone(remote string) error {
	dir := r.Dir
	cmd := exec.Command("git", "clone", remote, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %s", strings.TrimSpace(string(out)))
	}
	r.Remote = remote
	return nil
}

func (r *GitRepo) ConfigUser(name, email string) error {
	if err := r.run("config", "user.name", name); err != nil {
		return err
	}
	return r.run("config", "user.email", email)
}

// --- state ---

func (r *GitRepo) Initialized() bool {
	_, err := os.Stat(filepath.Join(r.Dir, ".git"))
	return err == nil
}

func (r *GitRepo) Branch() (string, error) {
	out, err := r.output("branch", "--show-current")
	if err != nil {
		return "", err
	}
	if out == "" {
		out = "(detached)"
	}
	return out, nil
}

func (r *GitRepo) HeadHash() (string, error) {
	out, err := r.output("rev-parse", "--verify", "HEAD")
	if err != nil && strings.Contains(err.Error(), "fatal") {
		return "", nil // empty repo case
	}
	return out, err
}

type GitStatus struct {
	HasChanges   bool
	ChangedFiles []string
	Branch       string
	Clean        bool
}

func (r *GitRepo) Status() (*GitStatus, error) {
	out, err := r.output("status", "--porcelain")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")

	branch, _ := r.Branch()
	hasChanges := strings.TrimSpace(out) != ""
	return &GitStatus{
		HasChanges:   hasChanges,
		ChangedFiles: lines,
		Branch:       branch,
		Clean:        !hasChanges,
	}, nil
}

// --- executions ---
func (r *GitRepo) Fetch() error {
	return r.run("fetch", "origin", "--prune", "--quiet")
}

func (r *GitRepo) Pull() error {
	_ = r.run("config", "pull.rebase", "false")
	out, err := r.output("pull", "--ff-only")
	if err != nil && strings.Contains(out, "non-fast-forward") {
		return ErrManualResolutionRequired
	}
	return err
}

func (r *GitRepo) Push() error {
	out, err := r.output("push", "origin", "HEAD")
	if err != nil && strings.Contains(out, "Permission denied") {
		return ErrNoWritePermission
	}
	return err
}
