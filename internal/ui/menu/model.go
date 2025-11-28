package menu

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/ui/common"
)

// MenuChoice는 메인 메뉴 선택지를 나타냅니다.
type MenuChoice int

const (
	ChoiceUpdate MenuChoice = iota
	ChoiceSettings
	ChoiceExit
)

type Model struct {
	choices  []string
	cursor   int
	Selected MenuChoice // 선택된 항목 (상위 모델에서 확인용)
	Chosen   bool       // 선택이 완료되었는지 여부
}

func NewModel() Model {
	return Model{
		choices:  []string{"시스템 업데이트 관리", "환경 설정", "종료"},
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
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.Selected = MenuChoice(m.cursor)
			m.Chosen = true
			if m.Selected == ChoiceExit {
				return m, tea.Quit
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := common.TitleStyle.Render("My Updater - 시스템 업데이트 관리자") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		style := common.MenuItemStyle
		if m.cursor == i {
			cursor = ">"
			style = common.SelectedItemStyle
		}
		s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
	}

	s += common.HelpStyle.Render("\n방향키: 이동 • Enter: 선택")
	return s
}
