package cmd

import (
	"fmt"
	"os"

	"github.com/ochcaroline/tmust/internal/check"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tmust",
	Short: "A minimal tmux session manager powered by zoxide",
	Long: `tmust — a minimal tmux session manager powered by zoxide

Pick a directory from your zoxide history and attach to (or create) a
tmux session for it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOpen()
	},
	// Check that all required tools are present before any command runs.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return check.Dependencies()
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(versionCmd)
}
