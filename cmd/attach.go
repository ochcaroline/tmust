package cmd

import (
	"fmt"
	"strings"

	"github.com/ochcaroline/tmust/internal/fzf"
	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:     "attach [session]",
	Aliases: []string{"a"},
	Short:   "Attach to an existing tmux session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return tmux.Attach(args[0])
		}
		return attachInteractive()
	},
}

func attachInteractive() error {
	sessions, err := tmux.List()
	if err != nil {
		fmt.Println("no sessions (or tmux server not running)")
		return nil
	}

	selected, err := fzf.Pick(sessions, "attach> ")
	if err != nil || selected == "" {
		return err
	}

	// tmux list-sessions format: "<name>: N windows (created ...)"
	name := strings.SplitN(strings.TrimSpace(selected), ":", 2)[0]
	return tmux.Attach(name)
}
