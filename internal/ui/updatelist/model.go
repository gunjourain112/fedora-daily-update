package updatelist

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/config"
	"my-updater/internal/domain"
	"my-updater/internal/ui/common"
)

type Item struct {
	task domain.Task
}

func (i Item) Title() string       { return i.task.Name }
func (i Item) Description() string { return "실행하려면 Enter를 누르세요" }
func (i Item) FilterValue() string { return i.task.Name }

type Model struct {
	List         list.Model
	SelectedTask *domain.Task
	Exit         bool
}

func NewModel(tm *domain.TaskManager, cfg *config.Config) Model {
	// Convert config tasks to domain tasks
	customTasks := []domain.Task{}
	for _, ct := range cfg.CustomTasks {
		customTasks = append(customTasks, domain.Task{
			Name:    ct.Name,
			Command: ct.Command,
			Args:    ct.Args,
		})
	}

	allTasks := tm.GetTasks(customTasks)
	items := []list.Item{}
	for _, t := range allTasks {
		items = append(items, Item{task: t})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "업데이트 목록"
	l.SetShowHelp(false)

	return Model{
		List: l,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := common.DocStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		if m.List.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "esc", "q":
			m.Exit = true
			return m, nil
		case "enter":
			if i, ok := m.List.SelectedItem().(Item); ok {
				t := i.task
				m.SelectedTask = &t
				return m, nil // Parent will handle transition to Runner
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return common.DocStyle.Render(m.List.View())
}
