package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	doneStyle    = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("#6B7280"))
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))

	checkDone    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✓")
	checkPending = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("○")

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
)

// ── Persistence ───────────────────────────────────────────────────────────────

type savedTodo struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func dataPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "my-tasks", "tasks.json")
}

func loadTodos() []todo {
	data, err := os.ReadFile(dataPath())
	if err != nil {
		// No save file yet — return empty list
		return []todo{}
	}
	var saved []savedTodo
	if err := json.Unmarshal(data, &saved); err != nil {
		return []todo{}
	}
	todos := make([]todo, len(saved))
	for i, s := range saved {
		todos[i] = todo{title: s.Title, done: s.Done}
	}
	return todos
}

func saveTodos(todos []todo) error {
	path := dataPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	saved := make([]savedTodo, len(todos))
	for i, t := range todos {
		saved[i] = savedTodo{Title: t.title, Done: t.done}
	}
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ── Todo item ─────────────────────────────────────────────────────────────────

type todo struct {
	title string
	done  bool
}

func (t todo) Title() string {
	check := checkPending
	text := pendingStyle.Render(t.title)
	if t.done {
		check = checkDone
		text = doneStyle.Render(t.title)
	}
	return fmt.Sprintf("%s  %s", check, text)
}

func (t todo) Description() string { return "" }
func (t todo) FilterValue() string  { return t.title }

// ── Key bindings ──────────────────────────────────────────────────────────────

type keyMap struct {
	Toggle   key.Binding
	Add      key.Binding
	Delete   key.Binding
	MoveUp   key.Binding
	MoveDown key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle done"),
	),
	Add: key.NewBinding(
		key.WithKeys("a", "n"),
		key.WithHelp("a/n", "add task"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "x", "delete"),
		key.WithHelp("d/x", "delete"),
	),
	MoveUp: key.NewBinding(
		key.WithKeys("shift+up", "K"),
		key.WithHelp("shift+↑/K", "move up"),
	),
	MoveDown: key.NewBinding(
		key.WithKeys("shift+down", "J"),
		key.WithHelp("shift+↓/J", "move down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// ── View state ────────────────────────────────────────────────────────────────

type viewState int

const (
	viewList viewState = iota
	viewAdd
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	list      list.Model
	input     textinput.Model
	state     viewState
	todos     []todo
	width     int
	height    int
	statusMsg string
}

func newModel() model {
	initial := loadTodos()

	items := todosToItems(initial)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#7C3AED")).
		BorderLeftForeground(lipgloss.Color("#7C3AED"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#9F7AEA")).
		BorderLeftForeground(lipgloss.Color("#7C3AED"))
	delegate.ShowDescription = false

	l := list.New(items, delegate, 0, 0)
	l.Title = "My Tasks"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Toggle, keys.Add, keys.Delete, keys.MoveUp}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Toggle, keys.Add, keys.Delete, keys.MoveUp, keys.MoveDown, keys.Quit}
	}

	ti := textinput.New()
	ti.Placeholder = "What needs to be done?"
	ti.CharLimit = 120
	ti.Width = 40

	return model{
		list:  l,
		input: ti,
		todos: initial,
		state: viewList,
	}
}

func todosToItems(ts []todo) []list.Item {
	items := make([]list.Item, len(ts))
	for i, t := range ts {
		items[i] = t
	}
	return items
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		// If filtering is active, let the list handle everything
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch m.state {

		// ── List view keys ────────────────────────────────────────────────
		case viewList:
			switch {
			case key.Matches(msg, keys.Quit):
				saveTodos(m.todos)
				return m, tea.Quit

			case key.Matches(msg, keys.Add):
				m.state = viewAdd
				m.input.Reset()
				m.input.Focus()
				m.statusMsg = ""
				return m, textinput.Blink

			case key.Matches(msg, keys.Toggle):
				i, ok := m.list.SelectedItem().(todo)
				if !ok {
					break
				}
				idx := m.list.Index()
				m.todos[idx].done = !i.done
				m.list.SetItem(idx, m.todos[idx])
				saveTodos(m.todos)
				if m.todos[idx].done {
					m.statusMsg = fmt.Sprintf("Marked \"%s\" as done", i.title)
				} else {
					m.statusMsg = fmt.Sprintf("Marked \"%s\" as pending", i.title)
				}
				return m, nil

			case key.Matches(msg, keys.Delete):
				idx := m.list.Index()
				if len(m.todos) == 0 {
					break
				}
				title := m.todos[idx].title
				m.todos = append(m.todos[:idx], m.todos[idx+1:]...)
				cmd := m.list.SetItems(todosToItems(m.todos))
				saveTodos(m.todos)
				m.statusMsg = fmt.Sprintf("Deleted \"%s\"", title)
				return m, cmd

			case key.Matches(msg, keys.MoveUp):
				idx := m.list.Index()
				if idx == 0 {
					break
				}
				m.todos[idx], m.todos[idx-1] = m.todos[idx-1], m.todos[idx]
				cmd := m.list.SetItems(todosToItems(m.todos))
				m.list.Select(idx - 1)
				saveTodos(m.todos)
				m.statusMsg = ""
				return m, cmd

			case key.Matches(msg, keys.MoveDown):
				idx := m.list.Index()
				if idx >= len(m.todos)-1 {
					break
				}
				m.todos[idx], m.todos[idx+1] = m.todos[idx+1], m.todos[idx]
				cmd := m.list.SetItems(todosToItems(m.todos))
				m.list.Select(idx + 1)
				saveTodos(m.todos)
				m.statusMsg = ""
				return m, cmd
			}

		// ── Add view keys ─────────────────────────────────────────────────
		case viewAdd:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if val != "" {
					newTodo := todo{title: val}
					m.todos = append(m.todos, newTodo)
					cmd := m.list.SetItems(todosToItems(m.todos))
					m.list.Select(len(m.todos) - 1)
					saveTodos(m.todos)
					m.statusMsg = fmt.Sprintf("Added \"%s\"", val)
					m.input.Blur()
					m.state = viewList
					return m, cmd
				}
				m.input.Blur()
				m.state = viewList
				return m, nil

			case "esc":
				m.input.Blur()
				m.state = viewList
				return m, nil
			}
		}
	}

	// Delegate to sub-components
	var cmd tea.Cmd
	switch m.state {
	case viewList:
		m.list, cmd = m.list.Update(msg)
	case viewAdd:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.state == viewAdd {
		return m.addView()
	}
	return m.listView()
}

func (m model) listView() string {
	done := 0
	for _, t := range m.todos {
		if t.done {
			done++
		}
	}

	progress := fmt.Sprintf(" %d/%d done ", done, len(m.todos))
	bar := statusBarStyle.Render(progress)

	status := ""
	if m.statusMsg != "" {
		status = helpStyle.Render("  " + m.statusMsg)
	}

	footer := lipgloss.JoinHorizontal(lipgloss.Top, bar, status)

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.list.View(),
			"",
			footer,
		),
	)
}

func (m model) addView() string {
	header := titleStyle.Render("Add Task")
	label := lipgloss.NewStyle().Bold(true).Render("Enter your task:")
	input := inputStyle.Render(m.input.View())
	help := helpStyle.Render("enter  confirm  •  esc  cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		label,
		"",
		input,
		"",
		help,
	)
	return appStyle.Render(content)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
