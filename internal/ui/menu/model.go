package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/ui/common"
)

// Choice represents a menu item.
type Choice int

const (
	ChoiceUpdate Choice = iota
	ChoiceSettings
)

type Model struct {
	Choices  []string
	Cursor   int
	Selected Choice
	Quitting bool
}

func NewModel() Model {
	return Model{
		Choices:  []string{"업데이트 목록 (Update List)", "설정 (Settings)"},
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
		case "q", "esc": // Allow esc to quit from main menu
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
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := "\n" + common.TitleStyle.Render("시스템 메뉴를 선택하세요") + "\n\n"

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = common.SelectedStyle.String()
		} else {
			cursor = common.ItemStyle.String()
		}

		line := choice
		if m.Cursor == i {
			line = lipgloss.NewStyle().Bold(true).Foreground(common.ColorHighlight).Render(line)
		}

		s += cursor + line + "\n"
	}

	s += "\n(이동: ↑/↓, 선택: Enter, 종료: q)\n"
	return common.DocStyle.Render(s)
}
