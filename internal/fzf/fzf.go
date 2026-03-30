package fzf

import (
	"os"
	"os/exec"
	"strings"
)

// Pick pipes input into fzf and returns the selected line.
// Returns ("", nil) when the user cancels (exit code 130).
func Pick(input, prompt string) (string, error) {
	cmd := exec.Command("fzf",
		"--height=40%",
		"--reverse",
		"--border",
		"--prompt="+prompt,
	)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return "", nil // user cancelled — not an error
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
