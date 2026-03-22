# My Tasks

A terminal to-do app built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

![demo](https://raw.githubusercontent.com/CookieShualon/my-tasks/ce48f4b084eed2071641e52a47ce742d7d9b4c8f/demo.gif)

## Build

```sh
go build -o my-tasks .
```

## Run

```sh
./my-tasks
```

## Keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `space` | Toggle task done/pending |
| `a` / `n` | Add a new task |
| `d` / `x` | Delete selected task |
| `shift+↑` / `K` | Move task up |
| `shift+↓` / `J` | Move task down |
| `/` | Filter tasks |
| `?` | Toggle full help |
| `q` / `ctrl+c` | Quit |

## Data

Tasks are saved automatically to `~/Library/Application Support/my-tasks/tasks.json` (macOS) or the equivalent config directory on your platform. Changes are written on every add, toggle, delete, and quit.
