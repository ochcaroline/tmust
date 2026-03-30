package check

import (
	"fmt"
	"os/exec"
	"strings"
)

var required = []string{"tmux", "fzf", "zoxide"}

// Dependencies checks that all required external tools are available on PATH.
func Dependencies() error {
	var missing []string
	for _, dep := range required {
		if _, err := exec.LookPath(dep); err != nil {
			missing = append(missing, dep)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s\nInstall them and make sure they are on your PATH", strings.Join(missing, ", "))
	}
	return nil
}
