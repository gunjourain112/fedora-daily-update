package updater

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/domain"
	"my-updater/internal/ui/common"
)

// item은 list.Item 인터페이스를 구현합니다.
type item struct {
	task domain.Task
}

func (i item) Title() string       { return i.task.Name }
func (i item) Description() string { return fmt.Sprintf("명령어: %s %v", i.task.Command, i.task.Args) }
func (i item) FilterValue() string { return i.task.Name }

type Model struct {
	list list.Model
	Exit bool
}

func NewModel(tasks []domain.Task) Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{task: t}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "업데이트 목록"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = common.TitleStyle

	return Model{
		list: l,
		Exit: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.Exit = true
			return m, nil
		}

		if msg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(item); ok {
				// tea.ExecProcess를 사용하여 대화형 쉘 실행
				c := exec.Command(i.task.Command, i.task.Args...)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return finishedMsg{err}
				})
			}
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // 여백 확보

	case finishedMsg:
		// 실행 완료 후 처리 (예: 성공 메시지 표시 등)
		// 지금은 단순히 다시 목록으로 돌아옴
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return "\n" + m.list.View()
}

type finishedMsg struct {
	err error
}
