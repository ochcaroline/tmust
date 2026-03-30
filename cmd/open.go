package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ochcaroline/tmust/internal/fzf"
	"github.com/ochcaroline/tmust/internal/tmux"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Pick a directory via zoxide+fzf and attach to its session",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOpen()
	},
}

func runOpen() error {
	zoxideOut, err := exec.Command("zoxide", "query", "--list").Output()
	if err != nil {
		return fmt.Errorf("zoxide error: %w", err)
	}

	dir, err := fzf.Pick(string(zoxideOut), "session> ")
	if err != nil || dir == "" {
		return err
	}

	name := tmux.SanitizeName(filepath.Base(dir))
	return tmux.AttachOrCreate(name, dir)
}
