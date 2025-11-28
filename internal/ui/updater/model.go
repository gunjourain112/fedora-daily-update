package updater

import (
	"fmt"
	"os/exec"
	"strings"

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

// SetSize는 리스트의 크기를 수동으로 설정합니다.
func (m *Model) SetSize(width, height int) {
	m.list.SetWidth(width)
	m.list.SetHeight(height - 4)
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
				// 명령어를 sh -c로 감싸서 실행 후 대기
				fullCmd := fmt.Sprintf("%s %s", i.task.Command, strings.Join(i.task.Args, " "))

				// 쉘 명령어로 변환: 명령어 실행 -> 줄바꿈 -> 프롬프트 -> 입력 대기
				shellCmd := fmt.Sprintf("%s; echo; read -p '엔터를 누르면 돌아갑니다...' _", fullCmd)

				c := exec.Command("sh", "-c", shellCmd)
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return finishedMsg{err}
				})
			}
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case finishedMsg:
		// 실행 완료 후 처리
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
