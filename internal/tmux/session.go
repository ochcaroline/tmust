package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// List returns the raw output of `tmux list-sessions`.
func List() (string, error) {
	out, err := exec.Command("tmux", "list-sessions").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Exists reports whether a session with the given name is active.
func Exists(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

// Create starts a new detached session named name with working directory dir.
func Create(name, dir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir).Run()
}

// Kill terminates the named session.
func Kill(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

// Attach attaches to the named session.
// When already inside tmux it switches the client instead.
func Attach(name string) error {
	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", name)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FindByDir returns the name of an existing session whose session_path matches
// dir, or ("", nil) if none exists. session_path is the directory that was
// passed to `tmux new-session -c` and does not change as the user navigates.
func FindByDir(dir string) (string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}\t#{session_path}").Output()
	if err != nil {
		return "", nil // no server running — not an error
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 && parts[1] == dir {
			return parts[0], nil
		}
	}
	return "", nil
}

// AttachOrCreate attaches to an existing session or creates one rooted at dir.
// Before creating, it checks whether any session is already rooted at dir and
// reuses it if so — even if its name differs from the derived name.
func AttachOrCreate(name, dir string) error {
	if existing, _ := FindByDir(dir); existing != "" {
		fmt.Printf("reusing session: %s  (%s)\n", existing, dir)
		return Attach(existing)
	}
	if !Exists(name) {
		fmt.Printf("creating session: %s  (%s)\n", name, dir)
		if err := Create(name, dir); err != nil {
			return fmt.Errorf("failed to create session %q: %w", name, err)
		}
	}
	return Attach(name)
}

// CurrentSession returns the name of the tmux session the current client is
// attached to. Returns an error when not inside tmux.
func CurrentSession() (string, error) {
	if os.Getenv("TMUX") == "" {
		return "", fmt.Errorf("not inside a tmux session")
	}
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// SwitchToLast switches the current client to the most recently used session.
func SwitchToLast() error {
	return exec.Command("tmux", "switch-client", "-l").Run()
}

func SanitizeName(s string) string {
	return strings.NewReplacer(".", "_", ":", "_", " ", "_").Replace(s)
}
