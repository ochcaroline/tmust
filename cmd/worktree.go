package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ochcaroline/tmust/internal/fzf"
	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:     "worktree",
	Aliases: []string{"wt"},
	Short:   "Create or end a worktree tmux session",
}

var worktreeNewCmd = &cobra.Command{
	Use:     "new [branch]",
	Aliases: []string{"create"},
	Short:   "Create a worktree and its tmux session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createWorktree(args)
	},
}

var worktreeEndCmd = &cobra.Command{
	Use:     "end [session]",
	Aliases: []string{"kill"},
	Short:   "End a worktree tmux session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return endWorktree(args)
	},
}

func init() {
	worktreeCmd.AddCommand(worktreeNewCmd, worktreeEndCmd)
	rootCmd.AddCommand(worktreeCmd)
}

func createWorktree(args []string) error {
	repo, err := pickRepo()
	if err != nil || repo == "" {
		return err
	}
	branch := ""
	if len(args) == 1 {
		branch = args[0]
	} else {
		fmt.Print("branch> ")
		if _, err := fmt.Scanln(&branch); err != nil {
			return err
		}
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return fmt.Errorf("branch cannot be empty")
	}
	if err := runGit(repo, "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("invalid branch %q", branch)
	}
	name := tmux.SanitizeName(filepath.Base(repo) + "-" + branch)
	if tmux.Exists(name) {
		return fmt.Errorf("tmux session already exists: %s", name)
	}

	worktree := filepath.Join(repo, ".worktrees", filepath.FromSlash(branch))
	if _, err := os.Stat(worktree); err == nil {
		return fmt.Errorf("worktree already exists: %s", worktree)
	}
	if err := os.MkdirAll(filepath.Dir(worktree), 0o755); err != nil {
		return err
	}

	argsGit := []string{"worktree", "add"}
	if ref, ok := branchRef(repo, branch); ok {
		if strings.HasPrefix(ref, "refs/remotes/") {
			argsGit = append(argsGit, "-b", branch, worktree, ref)
		} else {
			argsGit = append(argsGit, worktree, ref)
		}
	} else {
		argsGit = append(argsGit, "-b", branch, worktree, "HEAD")
	}
	if err := runGit(repo, argsGit...); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".env")); err == nil {
		if err := copyFile(filepath.Join(repo, ".env"), filepath.Join(worktree, ".env")); err != nil {
			return fmt.Errorf("copy .env: %w", err)
		}
	}

	if err := tmux.CreateWorktreeSession(name, worktree); err != nil {
		return err
	}
	return tmux.Attach(name)
}

func endWorktree(args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	} else {
		sessions, err := tmux.List()
		if err != nil {
			return fmt.Errorf("no sessions (or tmux server not running)")
		}
		selected, err := fzf.Pick(sessions, "end worktree> ")
		if err != nil || selected == "" {
			return err
		}
		name = strings.SplitN(strings.TrimSpace(selected), ":", 2)[0]
	}
	dir, err := tmux.SessionDir(name)
	if err != nil {
		return fmt.Errorf("find session worktree: %w", err)
	}
	repo, err := gitCommonRepo(dir)
	if err != nil {
		return fmt.Errorf("find worktree repository: %w", err)
	}
	dirty, err := gitOutput(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("check worktree: %w", err)
	}
	removeArgs := []string{"worktree", "remove"}
	if strings.TrimSpace(dirty) != "" {
		fmt.Printf("worktree has uncommitted changes:\n%sRemove it anyway? [y/N] ", dirty)
		answer, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
		if readErr != nil && len(answer) == 0 {
			return readErr
		}
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
			return fmt.Errorf("worktree was not removed")
		}
		removeArgs = append(removeArgs, "--force")
	}
	if err := runGit(repo, append(removeArgs, dir)...); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	if err := tmux.Kill(name); err != nil {
		return fmt.Errorf("failed to end session %q: %w", name, err)
	}
	fmt.Printf("ended session: %s\n", name)
	return nil
}

func gitCommonRepo(worktree string) (string, error) {
	common, err := gitOutput(worktree, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(common)), nil
}

func gitOutput(repo string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).Output()
	return string(out), err
}

func pickRepo() (string, error) {
	out, err := exec.Command("zoxide", "query", "--list").Output()
	if err != nil {
		return "", fmt.Errorf("zoxide error: %w", err)
	}
	dir, err := fzf.Pick(string(out), "repo> ")
	if err != nil {
		return "", err
	}
	return filepath.Abs(strings.TrimSpace(dir))
}

func branchRef(repo, branch string) (string, bool) {
	if runGit(repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch) == nil {
		return branch, true
	}
	out, err := exec.Command("git", "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/remotes").Output()
	if err != nil {
		return "", false
	}
	for ref := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if strings.HasSuffix(ref, "/"+branch) {
			return "refs/remotes/" + ref, true
		}
	}
	return "", false
}

func runGit(repo string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.ReadFrom(in)
	return err
}
