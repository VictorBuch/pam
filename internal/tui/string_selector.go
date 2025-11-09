package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type stringItem struct {
	title string
}

func (i stringItem) FilterValue() string {
	return i.title
}

func (i stringItem) Title() string {
	return i.title
}

func (i stringItem) Description() string {
	return ""
}

type stringListModel struct {
	list     list.Model
	choice   *stringItem
	quitting bool
}

func (m stringListModel) Init() tea.Cmd {
	return nil
}

func (m stringListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(stringItem); ok {
				m.choice = &i
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

func (m stringListModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

func ShowStringListSelector(title string, options []string) (string, error) {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = stringItem{title: opt}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetFilteringEnabled(true)

	model := stringListModel{list: l, quitting: false, choice: nil}

	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("Error running the program: %w", err)
	}

	if result, ok := m.(stringListModel); ok {
		if result.choice == nil {
			return "", fmt.Errorf("Selection cancelled")
		} else {
			return result.choice.title, nil
		}
	}

	return "", fmt.Errorf("unexpected model type")
}
