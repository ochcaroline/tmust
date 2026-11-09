package tmux

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Agent is a recognised agent process running inside a tmux pane.
type Agent struct {
	Name    string
	Session string
	Window  string
	Pane    string
	Path    string
}

type proc struct {
	ppid int
	cmd  string
}

// AgentName returns the agent name for a process command line, or "".
func AgentName(cmdline string) string {
	f := strings.Fields(cmdline)
	if len(f) == 0 {
		return ""
	}
	base := filepath.Base(f[0])
	if (base == "node" || base == "bun") && len(f) > 1 {
		base = filepath.Base(f[1])
	}
	switch base {
	case "claude", "opencode":
		return base
	case "copilot":
		return "gh copilot"
	case "gh":
		if len(f) > 1 && f[1] == "copilot" {
			return "gh copilot"
		}
	}
	return ""
}

// Agents returns every recognised agent process running in any tmux pane.
func Agents() ([]Agent, error) {
	panes, err := exec.Command("tmux", "list-panes", "-a", "-F",
		"#{session_name}\t#{window_id}\t#{pane_id}\t#{pane_pid}\t#{pane_current_path}").Output()
	if err != nil {
		return nil, err
	}
	psOut, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
	if err != nil {
		return nil, err
	}
	procs := map[int]proc{}
	children := map[int][]int{}
	for line := range strings.SplitSeq(string(psOut), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(f[0])
		ppid, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil {
			continue
		}
		procs[pid] = proc{ppid, strings.Join(f[2:], " ")}
		children[ppid] = append(children[ppid], pid)
	}

	var agents []Agent
	for line := range strings.SplitSeq(strings.TrimSpace(string(panes)), "\n") {
		p := strings.Split(line, "\t")
		if len(p) != 5 {
			continue
		}
		root, err := strconv.Atoi(p[3])
		if err != nil {
			continue
		}
		stack := []int{root}
		for len(stack) > 0 {
			pid := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if name := AgentName(procs[pid].cmd); name != "" {
				agents = append(agents, Agent{name, p[0], p[1], p[2], p[4]})
			}
			stack = append(stack, children[pid]...)
		}
	}
	return agents, nil
}

// Jump focuses the agent's window and pane, then switches to its session.
func Jump(a Agent) error {
	if err := exec.Command("tmux", "select-window", "-t", a.Window).Run(); err != nil {
		return err
	}
	if err := exec.Command("tmux", "select-pane", "-t", a.Pane).Run(); err != nil {
		return err
	}
	return Attach(a.Session)
}
