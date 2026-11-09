package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ochcaroline/tmust/internal/fzf"
	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Pick a running agent (claude, opencode, gh copilot) and jump to it",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if list, _ := cmd.Flags().GetBool("list"); list {
			rows, err := agentRows()
			fmt.Print(rows)
			return err
		}
		return runAgents()
	},
}

// Brand colours taken from each agent's logo.
var agentIcon = map[string]string{
	"claude":     "\x1b[38;2;217;119;87m◼\x1b[0m",
	"opencode":   "\x1b[38;2;207;206;205m▣\x1b[0m",
	"gh copilot": "\x1b[38;2;88;166;255m◉\x1b[0m",
}

var spinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func init() {
	agentsCmd.Flags().Bool("list", false, "print rows and exit")
	agentsCmd.Flags().MarkHidden("list")
	rootCmd.AddCommand(agentsCmd)
}

func runAgents() error {
	rows, err := agentRows()
	if err != nil {
		return err
	}
	if rows == "" {
		fmt.Println("no running agents")
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	selected, err := fzf.Pick(rows, "agent> ", "--ansi", "--delimiter=\t", "--with-nth=1",
		"--bind=load:reload-sync(sleep 0.15; '"+exe+"' agents --list)")
	if err != nil || selected == "" {
		return err
	}
	_, pane, _ := strings.Cut(selected, "\t")
	agents, err := tmux.Agents()
	if err != nil {
		return err
	}
	for _, a := range agents {
		if a.Pane == pane {
			return tmux.Jump(a)
		}
	}
	return fmt.Errorf("agent %s is gone", pane)
}

func agentRows() (string, error) {
	agents, err := tmux.Agents()
	if err != nil {
		return "", fmt.Errorf("no sessions (or tmux server not running)")
	}
	worktrees := make([]string, len(agents))
	nameW, sessionW := 0, 0
	for i, a := range agents {
		worktrees[i] = "-"
		if branch, err := gitOutput(a.Path, "rev-parse", "--abbrev-ref", "HEAD"); err == nil && strings.TrimSpace(branch) != "" {
			worktrees[i] = strings.TrimSpace(branch)
		}
		nameW = max(nameW, utf8.RuneCountInString(a.Name))
		sessionW = max(sessionW, utf8.RuneCountInString(a.Session))
	}
	spin := spinner[time.Now().UnixMilli()/120%int64(len(spinner))]
	var b strings.Builder
	for i, a := range agents {
		state := " "
		if tmux.Working(a) {
			state = spin
		}
		fmt.Fprintf(&b, "%s %s %-*s  %-*s  %s\t%s\n", state, agentIcon[a.Name], nameW, a.Name, sessionW, a.Session, worktrees[i], a.Pane)
	}
	return b.String(), nil
}
