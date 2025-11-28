package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/app"
	"my-updater/internal/domain"
	"my-updater/internal/ui/common"
)

// ViewState는 설정 화면의 상태를 나타냅니다.
type ViewState int

const (
	ViewStateList ViewState = iota
	ViewStateForm
)

// item은 리스트 아이템입니다.
type item struct {
	task domain.Task
}

func (i item) Title() string {
	if i.task.Type == domain.TaskTypeBuiltin {
		return "[기본] " + i.task.Name
	}
	return i.task.Name
}
func (i item) Description() string { return i.task.Command + " " + strings.Join(i.task.Args, " ") }
func (i item) FilterValue() string { return i.task.Name }

type Model struct {
	service     *app.TaskService
	list        list.Model
	inputs      []textinput.Model
	focusIndex  int
	viewState   ViewState
	editingTask *domain.Task // 현재 수정 중인 태스크 (nil이면 신규)
	Exit        bool
	width       int
	height      int
}

func NewModel(tasks []domain.Task, service *app.TaskService) Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{task: t}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "커스텀 메뉴 관리"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = common.TitleStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "추가")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "수정")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "삭제")),
		}
	}

	// 입력 필드 초기화
	inputs := make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "태스크 이름 (예: 내 스크립트)"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "명령어 (예: bash)"
	inputs[1].CharLimit = 50
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "인자 (예: -c 'echo hello')"
	inputs[2].CharLimit = 100
	inputs[2].Width = 50

	return Model{
		service:   service,
		list:      l,
		inputs:    inputs,
		viewState: ViewStateList,
	}
}

// SetSize는 리스트의 크기를 수동으로 설정합니다.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
	m.list.SetHeight(height - 4)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	if m.viewState == ViewStateList {
		return m.updateList(msg)
	} else {
		return m.updateForm(msg)
	}
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.list.FilterState() == list.Filtering {
				break // 필터링 중일 때는 기본 동작(필터 취소)에 맡김
			}
			m.Exit = true
			return m, nil

		case "a": // Add
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.viewState = ViewStateForm
			m.editingTask = nil
			m.resetInputs()
			return m, nil

		case "e", "enter": // Edit
			if m.list.FilterState() == list.Filtering {
				break
			}
			if i, ok := m.list.SelectedItem().(item); ok {
				if i.task.Type == domain.TaskTypeBuiltin {
					return m, nil // 기본 태스크는 수정 불가
				}
				m.viewState = ViewStateForm
				m.editingTask = &i.task // 포인터 복사 주의 (여기선 단순 참조로 사용하고 저장 시 원본 교체)
				m.setInputs(i.task)
				return m, nil
			}

		case "d": // Delete
			if m.list.FilterState() == list.Filtering {
				break
			}
			if i, ok := m.list.SelectedItem().(item); ok {
				if i.task.Type == domain.TaskTypeBuiltin {
					return m, nil // 기본 태스크는 삭제 불가
				}
				// 삭제 로직
				idx := m.list.Index()
				m.list.RemoveItem(idx)
				m.saveTasks() // 저장
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// 취소하고 리스트로 복귀
			m.viewState = ViewStateList
			return m, nil

		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				// 저장
				m.submitForm()
				m.viewState = ViewStateList
				return m, nil
			}
			m.nextInput()

		case "tab", "down":
			m.nextInput()

		case "shift+tab", "up":
			m.prevInput()
		}
	}

	// Handle input updates
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) nextInput() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	m.inputs[m.focusIndex].Focus()
}

func (m *Model) prevInput() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs) - 1
	}
	m.inputs[m.focusIndex].Focus()
}

func (m *Model) resetInputs() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	m.inputs[0].Focus()
	m.focusIndex = 0
}

func (m *Model) setInputs(t domain.Task) {
	m.inputs[0].SetValue(t.Name)
	m.inputs[1].SetValue(t.Command)
	m.inputs[2].SetValue(strings.Join(t.Args, " "))
	m.inputs[0].Focus()
	m.focusIndex = 0
}

func (m *Model) submitForm() {
	newTask := domain.Task{
		Name:    m.inputs[0].Value(),
		Command: m.inputs[1].Value(),
		Args:    strings.Fields(m.inputs[2].Value()), // 단순 공백 분리
		Type:    domain.TaskTypeCustom,
	}

	if m.editingTask == nil {
		// 추가
		newTask.ID = fmt.Sprintf("custom-%d", len(m.list.Items())) // 간단 ID
		m.list.InsertItem(len(m.list.Items()), item{task: newTask})
	} else {
		// 수정: 리스트에서 찾아서 교체
		// 현재 선택된 인덱스의 아이템을 교체
		idx := m.list.Index()
		newTask.ID = m.editingTask.ID // ID 유지
		m.list.SetItem(idx, item{task: newTask})
	}

	m.saveTasks()
}

func (m *Model) saveTasks() {
	tasks := []domain.Task{}
	for _, it := range m.list.Items() {
		tasks = append(tasks, it.(item).task)
	}
	// Service를 통해 저장
	// Error handling is omitted for brevity in TUI
	_ = m.service.SaveCustomTasks(tasks)
}

func (m Model) View() string {
	if m.viewState == ViewStateList {
		return "\n" + m.list.View()
	}

	// Form View
	s := strings.Builder{}
	s.WriteString(common.TitleStyle.Render("태스크 편집"))
	s.WriteString("\n\n")

	labels := []string{"이름", "명령어", "인자"}
	for i := range m.inputs {
		s.WriteString(labels[i] + "\n")
		s.WriteString(m.inputs[i].View())
		s.WriteString("\n\n")
	}

	s.WriteString(common.HelpStyle.Render("Enter: 저장 • Esc: 취소"))

	return lipgloss.NewStyle().Margin(1, 2).Render(s.String())
}
