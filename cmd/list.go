package cmd

import (
	"fmt"
	"os"

	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List active tmux sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := tmux.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, "no sessions (or tmux server not running)")
			return nil
		}
		fmt.Print(sessions)
		return nil
	},
}
