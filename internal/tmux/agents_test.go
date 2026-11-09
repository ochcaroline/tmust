package tmux

import "testing"

func TestAgentName(t *testing.T) {
	cases := map[string]string{
		"claude --resume":         "claude",
		"/usr/local/bin/opencode": "opencode",
		"node /opt/bin/claude":    "claude",
		"gh copilot suggest":      "gh copilot",
		"gh pr list":              "",
		"-zsh":                    "",
		"nvim .":                  "",
		"":                        "",
	}
	for in, want := range cases {
		if got := AgentName(in); got != want {
			t.Errorf("AgentName(%q) = %q, want %q", in, got, want)
		}
	}
}
