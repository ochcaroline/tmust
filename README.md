# tmust

A minimal tmux session manager that integrates with zoxide.

## Why does it exist?

All tools that I found (like `sesh`) are so bloated and I don't need most of their functionality, but they come with a lot of opinionated workflows.
So I created my own, because I could.

## Dependencies

- [tmux](https://github.com/tmux/tmux)
- [zoxide](https://github.com/ajeetdsouza/zoxide)
- [fzf](https://github.com/junegunn/fzf)

Don't worry - tmust will check if the're installed XD

## Usage

```
tmust                   pick a directory via zoxide+fzf, attach to or create its session
tmust open              same as above
tmust attach            pick from active sessions and attach
tmust attach <name>     attach to a named session directly
tmust ls                list active sessions
tmust kill              pick a session via fzf and kill it
tmust kill <name>       kill a named session directly
```

When tmust is run from inside an existing tmux session it uses `switch-client`
instead of `attach-session`, so you stay inside tmux at all times.

## Recommended tmux configuration

Add these to your `~/.tmux.conf` to make session management smoother.

```tmux
# When a session is destroyed, attach to the most recently used session
# instead of exiting tmux entirely.
set -g detach-on-destroy off

# Renumber windows when one is closed so there are no gaps.
set -g renumber-windows on

# Use a longer history.
set -g history-limit 50000

# Start window and pane indices at 1 (easier to reach on keyboard).
set -g base-index 1
setw -g pane-base-index 1
```

`detach-on-destroy off` is the most important one for tmust: without it,
killing a session drops you back to your shell instead of landing you in
another open session.
