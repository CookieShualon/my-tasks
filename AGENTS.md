# AGENTS.md

## Project Overview

`my-tasks` is a terminal-based to-do app written in Go, built with the Charm TUI stack (Bubble Tea, Bubbles, Lip Gloss). The entire application lives in a single file: `main.go`.

## Commands

```sh
# Build
go build -o my-tasks .

# Run
./my-tasks

# Build and run in one step
go run .

# Update dependencies
go mod tidy
```

No Makefile, no test suite, no linter config present.

## Project Structure

```
main.go        # Entire application — styles, persistence, model, update, view
go.mod         # Module: my-tasks, requires Go 1.26.1
go.sum         # Dependency checksums
```

All code is in `package main`. There are no subdirectories or packages to split across.

## Architecture — Bubble Tea MVU Pattern

The app follows the Elm-like MVU (Model-View-Update) architecture required by Bubble Tea:

| Component | Location | Role |
|-----------|----------|------|
| `model` struct | `main.go:163` | All application state |
| `(m model) Init()` | `main.go:222` | Returns initial `tea.Cmd` (nil here) |
| `(m model) Update()` | `main.go:228` | Handles all messages, returns updated model + cmd |
| `(m model) View()` | `main.go:354` | Renders current state as a string |

**Key rule**: `model` is a value type (not pointer receiver). Every `Update` call returns a new copy.

## State

```go
type viewState int
const (
    viewList viewState = iota  // Normal list browsing
    viewAdd                     // Text input modal for adding a task
)
```

The `model.state` field controls which view is rendered and which key bindings are active.

## Data Model

```go
type todo struct {
    title string
    done  bool
}
```

`todo` implements `list.Item` (from `charmbracelet/bubbles/list`) via `Title()`, `Description()`, and `FilterValue()` methods.

The in-memory list `model.todos []todo` is the source of truth. `model.list` (the Bubbles list widget) holds a parallel `[]list.Item` slice — both must be kept in sync. After any mutation call `m.list.SetItems(todosToItems(m.todos))`.

## Persistence

Tasks auto-save to JSON on every add, toggle, delete, move, and quit.

- **Save path (macOS)**: `~/Library/Application Support/my-tasks/tasks.json`
- **Save path (Linux/other)**: `$XDG_CONFIG_HOME/my-tasks/tasks.json` (falls back to `.` if `os.UserConfigDir()` fails)

Serialization uses `savedTodo` (exported fields for JSON) as an intermediary; `todo` fields are unexported.

## Styling

All styles are package-level `lipgloss.Style` vars defined at the top of `main.go`. The primary accent color is `#7C3AED` (purple). Adaptive colors are used for the status bar to support light/dark terminals.

## Key Bindings (defined in `keyMap`)

| Key | Action |
|-----|--------|
| `space` | Toggle done/pending |
| `a` / `n` | Enter add-task view |
| `d` / `x` / `delete` | Delete selected task |
| `shift+↑` / `K` | Move task up |
| `shift+↓` / `J` | Move task down |
| `q` / `ctrl+c` | Save and quit |
| `/` | Filter (built into Bubbles list) |
| `esc` | Cancel add view |

When the Bubbles list is in **filtering mode** (`list.Filtering`), all custom key handling is bypassed so the filter input receives all keys.

## Gotchas

- **Dual list sync**: `model.todos` and `model.list` items must always be kept in sync. Use `m.list.SetItems(todosToItems(m.todos))` after every mutation.
- **No tests**: There are no test files. Manual testing requires building and running the TUI.
- **Single-file app**: All code is in `main.go`. Keep it that way unless the file grows substantially — the Charm pattern works well in one file for small apps.
- **Value receiver on model**: Bubble Tea requires the model to be passed by value. Never use pointer receivers on `model`.
- **Go version**: `go.mod` specifies `go 1.26.1` — ensure your toolchain matches or exceeds this.
- **`saveTodos` is called eagerly**: Every mutating action saves immediately — there is no deferred/buffered save.
