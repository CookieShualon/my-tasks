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

To regenerate the demo GIF after UI changes, install [VHS](https://github.com/charmbracelet/vhs) and run:

```sh
vhs demo.tape
```

This replaces `demo.gif` in place. The tape is configured at 1300×650 px, font size 22, typing speed 75 ms. The recorded session launches the app, adds two tasks ("Buy groceries", "Read a book"), toggles one done, navigates down and deletes one, then triggers the filter with `/buy`.

After committing the new `demo.gif`, update the image URL in `README.md` to point to the latest commit SHA (e.g. `https://raw.githubusercontent.com/<owner>/my-tasks/<commit-sha>/demo.gif`) to bust GitHub's CDN cache. Using the raw URL pinned to a commit guarantees the updated GIF is shown immediately.

## Project Structure

```
main.go        # Entire application — styles, persistence, model, update, view
go.mod         # Module: my-tasks, requires Go 1.26.1
go.sum         # Dependency checksums
demo.tape      # VHS script that records the demo terminal session
demo.gif       # Output GIF embedded in the README (generated from demo.tape)
```

All code is in `package main`. There are no subdirectories or packages to split across.

## Architecture — Bubble Tea MVU Pattern

The app follows the Elm-like MVU (Model-View-Update) architecture required by Bubble Tea:

| Component | Location | Role |
|-----------|----------|------|
| `model` struct | `main.go:192` | All application state |
| `(m model) Init()` | `main.go:251` | Returns initial `tea.Cmd` (nil here) |
| `(m model) Update()` | `main.go:257` | Handles all messages, returns updated model + cmd |
| `(m model) View()` | `main.go:401` | Dispatches to `listView()` or `addView()` |

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

The in-memory list `model.todos []todo` is the source of truth. `model.list` (the Bubbles list widget) holds a parallel `[]list.Item` slice — both must be kept in sync. After any mutation call `m.list.SetItems(todosToItems(m.todos))`. For a single-item toggle, `m.list.SetItem(idx, item)` is used instead to avoid a full reset.

### model fields

```go
type model struct {
    list      list.Model
    input     textinput.Model
    state     viewState
    todos     []todo
    width     int
    height    int
    statusMsg string
}
```

- `width`/`height` — updated on every `tea.WindowSizeMsg` to keep the list properly sized.
- `statusMsg` — transient feedback shown in the footer (e.g. `"Added \"Buy milk\""`, `"Deleted \"...""`).

## Custom Delegate — `todoDelegate`

The list uses a custom `todoDelegate` (implements `list.ItemDelegate`) instead of the default delegate. It renders each item as a single line:

```
○  Task title        (pending)
✓  ~~Done task~~     (completed, strikethrough + grey)
```

The selected item gets a purple left border (`#7C3AED`). Non-selected items get 2-space left padding to align with the border width.

## Persistence

Tasks auto-save to JSON on every add, toggle, delete, move, and quit.

- **Save path (macOS)**: `~/Library/Application Support/my-tasks/tasks.json`
- **Save path (Linux/other)**: `$XDG_CONFIG_HOME/my-tasks/tasks.json` (falls back to `.` if `os.UserConfigDir()` fails)

Serialization uses `savedTodo` (exported fields for JSON) as an intermediary; `todo` fields are unexported.

## Styling

All styles are package-level `lipgloss.Style` vars defined at the top of `main.go`. The primary accent color is `#7C3AED` (purple). Adaptive colors are used for the status bar to support light/dark terminals.

| Var | Purpose |
|-----|---------|
| `appStyle` | Outer padding (1, 2) for the whole app |
| `titleStyle` | Purple background title bar |
| `statusBarStyle` | Adaptive footer bar showing task count |
| `inputStyle` | Rounded border input box in add view |
| `doneStyle` | Strikethrough + grey for completed task titles |
| `pendingStyle` | Near-white for pending task titles |
| `checkDoneStyle` | Green `✓` checkmark |
| `checkPendingStyle` | Grey `○` circle |
| `helpStyle` | Dim grey for status/help text |

## Views

`View()` at `main.go:401` dispatches to two sub-methods:

- **`listView()`** — renders the `list.Model`, then a footer row with a `done/total` progress bar and the current `statusMsg`.
- **`addView()`** — a simple modal with a `titleStyle` header, bold label, `inputStyle`-wrapped `textinput`, and help hint (`enter  confirm  •  esc  cancel`).

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

## Helper Functions

- **`todosToItems(ts []todo) []list.Item`** — converts the `[]todo` slice into `[]list.Item` for the Bubbles list widget.
- **`findTodoIndex(todos []todo, target todo) int`** — linear search matching on `title` and `done`; returns -1 if not found. Used to map a selected `list.Item` back to its index in `model.todos`.

## Gotchas

- **Dual list sync**: `model.todos` and `model.list` items must always be kept in sync. Use `m.list.SetItems(todosToItems(m.todos))` after mutations; use `m.list.SetItem(idx, item)` for single-item updates (e.g. toggle).
- **`findTodoIndex` limitation**: matches by `title + done`, so two tasks with identical text and state would be ambiguous. Avoid duplicates.
- **No tests**: There are no test files. Manual testing requires building and running the TUI.
- **Single-file app**: All code is in `main.go`. Keep it that way unless the file grows substantially — the Charm pattern works well in one file for small apps.
- **Value receiver on model**: Bubble Tea requires the model to be passed by value. Never use pointer receivers on `model`.
- **Go version**: `go.mod` specifies `go 1.26.1` — ensure your toolchain matches or exceeds this.
- **`saveTodos` is called eagerly**: Every mutating action saves immediately — there is no deferred/buffered save.
