# My Tasks

A minimal, keyboard-driven terminal to-do app built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

![demo](https://raw.githubusercontent.com/CookieShualon/my-tasks/d1d204cbba8a54bc895d49c4c2276e8a0dd23cd2/demo.gif)

## Features

- Add, delete, and reorder tasks
- Toggle tasks between pending and done
- Fuzzy filter to quickly find tasks
- Progress indicator showing completed vs. total tasks
- Tasks persist automatically across sessions

## Requirements

- [Go](https://go.dev/) 1.21 or later

## Installation

### Build from source

```sh
git clone https://github.com/CookieShualon/my-tasks.git
cd my-tasks
go build -o my-tasks .
./my-tasks
```

### Run without installing

```sh
go run .
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

Tasks are saved automatically to `~/Library/Application Support/my-tasks/tasks.json` on macOS, or the equivalent [`os.UserConfigDir()`](https://pkg.go.dev/os#UserConfigDir) path on other platforms. Changes are written on every add, toggle, delete, and quit.
