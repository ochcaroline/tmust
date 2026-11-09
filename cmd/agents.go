package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
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
		return runAgents()
	},
}

// Brand colours taken from each agent's logo.
var agentIcon = map[string]string{
	"claude":     "\x1b[38;2;217;119;87m◼\x1b[0m",
	"opencode":   "\x1b[38;2;207;206;205m▣\x1b[0m",
	"gh copilot": "\x1b[38;2;88;166;255m◉\x1b[0m",
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}

func runAgents() error {
	agents, err := tmux.Agents()
	if err != nil {
		return fmt.Errorf("no sessions (or tmux server not running)")
	}
	if len(agents) == 0 {
		fmt.Println("no running agents")
		return nil
	}
	byTarget := map[string]tmux.Agent{}
	worktrees := make([]string, len(agents))
	nameW, sessionW := 0, 0
	for i, a := range agents {
		worktrees[i] = filepath.Base(a.Path)
		if branch, err := gitOutput(a.Path, "rev-parse", "--abbrev-ref", "HEAD"); err == nil && strings.TrimSpace(branch) != "" {
			worktrees[i] = strings.TrimSpace(branch)
		}
		nameW = max(nameW, utf8.RuneCountInString(a.Name))
		sessionW = max(sessionW, utf8.RuneCountInString(a.Session))
	}
	var b strings.Builder
	for i, a := range agents {
		byTarget[a.Pane] = a
		fmt.Fprintf(&b, "%s %-*s  %-*s  %s\t%s\n", agentIcon[a.Name], nameW, a.Name, sessionW, a.Session, worktrees[i], a.Pane)
	}
	selected, err := fzf.Pick(b.String(), "agent> ", "--ansi", "--delimiter=\t", "--with-nth=1")
	if err != nil || selected == "" {
		return err
	}
	_, pane, _ := strings.Cut(selected, "\t")
	return tmux.Jump(byTarget[pane])
}
