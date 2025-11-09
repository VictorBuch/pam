package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"nap/internal/types"
)

type packageItem struct {
	pkg types.Package
}

func (i packageItem) FilterValue() string {
	return i.pkg.PName
}

func (i packageItem) Title() string {
	return fmt.Sprintf("%s (%s)", i.pkg.PName, i.pkg.Version)
}

func (i packageItem) Description() string {
	return i.pkg.Description
}

type packageListModel struct {
	list     list.Model
	choice   *types.Package
	quitting bool
}

func (m packageListModel) Init() tea.Cmd {
	return nil
}

func (m packageListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(packageItem); ok {
				m.choice = &i.pkg
			}
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m packageListModel) View() string {
	if m.quitting {
		return ""
	} else {
		return m.list.View()
	}
}

func ShowPackageSelector(packages []types.Package) (*types.Package, error) {
	items := make([]list.Item, len(packages))
	for i, pkg := range packages {
		items[i] = packageItem{pkg: pkg}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a package to install"
	l.SetFilteringEnabled(true)

	model := packageListModel{list: l, quitting: false, choice: nil}

	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		error := fmt.Errorf("Error running the program: %w", err)
		return nil, error
	}

	if result, ok := m.(packageListModel); ok {
		if result.choice == nil {
			return nil, fmt.Errorf("Selection cancelled")
		} else {
			return result.choice, nil
		}
	}

	return nil, fmt.Errorf("unexpected model type")
}
