package cmd

import (
	"fmt"
	"strings"

	"github.com/ochcaroline/tmust/internal/fzf"
	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var killCurrent bool

var killCmd = &cobra.Command{
	Use:     "kill [session]",
	Aliases: []string{"k"},
	Short:   "Kill a tmux session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if killCurrent {
			return killCurrentSession()
		}
		if len(args) == 1 {
			return killByName(args[0])
		}
		return killInteractive()
	},
}

func init() {
	killCmd.Flags().BoolVar(&killCurrent, "current", false, "kill the current session and switch to the most recent one")
}

// killCurrentSession switches to the last session first, then kills the one we
// just left. The order matters: killing first would drop us out of tmux if
// detach-on-destroy is not set.
func killCurrentSession() error {
	name, err := tmux.CurrentSession()
	if err != nil {
		return err
	}
	if err := tmux.SwitchToLast(); err != nil {
		return fmt.Errorf("no other session to switch to: %w", err)
	}
	return killByName(name)
}

func killByName(name string) error {
	if err := tmux.Kill(name); err != nil {
		return fmt.Errorf("failed to kill session %q: %w", name, err)
	}
	fmt.Printf("killed session: %s\n", name)
	return nil
}

func killInteractive() error {
	sessions, err := tmux.List()
	if err != nil {
		fmt.Println("no sessions (or tmux server not running)")
		return nil
	}

	selected, err := fzf.Pick(sessions, "kill> ")
	if err != nil || selected == "" {
		return err
	}

	// tmux list-sessions format: "<name>: N windows (created ...)"
	name := strings.SplitN(strings.TrimSpace(selected), ":", 2)[0]
	return killByName(name)
}
