package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type multiStringDelegate struct{}

func (d multiStringDelegate) Height() int {
	return 1
}

func (d multiStringDelegate) Spacing() int {
	return 0
}

func (d multiStringDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d multiStringDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if item, ok := item.(multiStringItem); ok {
		checkbox := "[ ]"
		if item.selected {
			checkbox = "[✓]"
		}
		cursor := " "
		if index == m.Index() {
			cursor = ">"
		}
		fmt.Fprintf(w, "%s %s %s", cursor, checkbox, item.title)

	}
}

type multiStringItem struct {
	title    string
	selected bool
}

func (i multiStringItem) FilterValue() string {
	return i.title
}

func (i multiStringItem) Title() string {
	return i.title
}

func (i multiStringItem) Description() string {
	return ""
}

type multiStringListModel struct {
	list     list.Model
	choice   []string
	quitting bool
}

func (m multiStringListModel) Init() tea.Cmd {
	return nil
}

func (m multiStringListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case " ":
			index := m.list.Index()
			if item, ok := m.list.Items()[index].(multiStringItem); ok {
				item.selected = !item.selected
				m.list.SetItem(index, item)
			}
			return m, nil

		case "enter":
			for _, item := range m.list.Items() {
				if multiItem, ok := item.(multiStringItem); ok {
					if multiItem.selected {
						m.choice = append(m.choice, multiItem.title)
					}
				}
			}
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m multiStringListModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

func ShowMultiStringListSelector(title string, options []string) ([]string, error) {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = multiStringItem{title: opt}
	}

	l := list.New(items, multiStringDelegate{}, 0, 0)
	l.Title = title
	l.SetFilteringEnabled(true)

	model := multiStringListModel{list: l, quitting: false, choice: []string{}}

	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running the program: %w", err)
	}

	if result, ok := m.(multiStringListModel); ok {
		if len(result.choice) == 0 {
			return []string{}, fmt.Errorf("Selection cancelled")
		}
		return result.choice, nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
