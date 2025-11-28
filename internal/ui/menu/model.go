package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Choice represents a menu item.
type Choice int

const (
	ChoiceUpdate Choice = iota
	ChoiceSettings
)

var (
	titleStyle    = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("205"))
	itemStyle     = lipgloss.NewStyle().PaddingLeft(4)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).SetString("> ")
)

type Model struct {
	Choices  []string
	Cursor   int
	Selected Choice
	Quitting bool
}

func NewModel() Model {
	return Model{
		Choices:  []string{"업데이트 시작 (Start Update)", "설정 (Settings)"},
		Selected: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter", " ":
			m.Selected = Choice(m.Cursor)
			// The parent model will handle the transition based on m.Selected
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := "\n" + titleStyle.Render("시스템 메뉴를 선택하세요") + "\n\n"

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = selectedStyle.String()
		} else {
			cursor = itemStyle.String() // indent for non-selected
		}

		// If selected, maybe highlight?
		line := choice
		if m.Cursor == i {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(line)
		}

		s += cursor + line + "\n"
	}

	s += "\n(이동: ↑/↓, 선택: Enter, 종료: q)\n"
	return s
}
