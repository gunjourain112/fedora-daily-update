package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/config"
)

// ViewState tracks which sub-view we are in.
type ViewState int

const (
	StateList ViewState = iota
	StateInputName
	StateInputCommand
	StateInputArgs
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

// Item wraps the config task for the list bubble.
type Item struct {
	task config.CustomTask
}

func (i Item) Title() string       { return i.task.Name }
func (i Item) Description() string { return fmt.Sprintf("%s %s", i.task.Command, strings.Join(i.task.Args, " ")) }
func (i Item) FilterValue() string { return i.task.Name }

type Model struct {
	ConfigManager *config.Manager
	Config        *config.Config

	List          list.Model
	Inputs        []textinput.Model
	FocusIndex    int
	State         ViewState

	// Temp storage for new/edited item
	EditMode      bool // true if editing, false if adding
	EditIndex     int
	TempTask      config.CustomTask

	Exit          bool // signal to parent to go back
}

func NewModel(cm *config.Manager) (Model, error) {
	cfg, err := cm.Load()
	if err != nil {
		return Model{}, err
	}

	// Prepare list items
	items := []list.Item{}
	for _, t := range cfg.CustomTasks {
		items = append(items, Item{task: t})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "커스텀 메뉴 관리"
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "추가")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "삭제")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "뒤로가기")),
		}
	}
	// We need to customize keybindings slightly to allow 'a' and 'd'

	// Inputs
	inputs := make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "메뉴 이름 (예: 내 스크립트)"
	inputs[0].Focus()
	inputs[0].CharLimit = 30
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "명령어 (예: /path/to/script.sh)"
	inputs[1].CharLimit = 100
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "인자 (예: arg1 arg2)"
	inputs[2].CharLimit = 100
	inputs[2].Width = 50

	return Model{
		ConfigManager: cm,
		Config:        cfg,
		List:          l,
		Inputs:        inputs,
		State:         StateList,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Handle keys based on state
		switch m.State {
		case StateList:
			switch msg.String() {
			case "esc":
				m.Exit = true
				return m, nil // Parent checks Exit
			case "a": // Add
				m.State = StateInputName
				m.EditMode = false
				m.resetInputs()
				m.FocusIndex = 0
				return m, textinput.Blink
			case "enter": // Edit (optional, but let's support it)
				if len(m.Config.CustomTasks) > 0 {
					idx := m.List.Index()
					m.State = StateInputName
					m.EditMode = true
					m.EditIndex = idx
					m.loadInputs(m.Config.CustomTasks[idx])
					m.FocusIndex = 0
					return m, textinput.Blink
				}
			case "d": // Delete
				if len(m.Config.CustomTasks) > 0 {
					idx := m.List.Index()
					m.Config.CustomTasks = append(m.Config.CustomTasks[:idx], m.Config.CustomTasks[idx+1:]...)
					m.ConfigManager.Save(m.Config)
					m.List.RemoveItem(idx)
					m.List.ResetSelected()
				}
			}

		case StateInputName, StateInputCommand, StateInputArgs:
			switch msg.String() {
			case "esc":
				// Cancel edit/add
				m.State = StateList
				return m, nil
			case "enter":
				if m.FocusIndex < len(m.Inputs)-1 {
					m.Inputs[m.FocusIndex].Blur()
					m.FocusIndex++
					m.Inputs[m.FocusIndex].Focus()
					return m, textinput.Blink
				} else {
					// Finish
					m.Inputs[m.FocusIndex].Blur()
					m.saveTask()
					m.State = StateList
					return m, nil
				}
			}
		}
	}

	// Update children
	if m.State == StateList {
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// Update inputs
		for i := range m.Inputs {
			m.Inputs[i], cmd = m.Inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.State == StateList {
		return docStyle.Render(m.List.View())
	}

	// Form View
	var b strings.Builder
	title := "새 메뉴 추가"
	if m.EditMode {
		title = "메뉴 수정"
	}
	b.WriteString(fmt.Sprintf("\n  %s\n\n", lipgloss.NewStyle().Bold(true).Render(title)))

	labels := []string{"이름", "명령어", "인자"}

	for i := range m.Inputs {
		b.WriteString(fmt.Sprintf("  %s:\n", labels[i]))
		b.WriteString(docStyle.Render(m.Inputs[i].View()))
		b.WriteString("\n")
	}

	b.WriteString("\n  (Enter: 다음/저장, Esc: 취소)\n")
	return b.String()
}

// Helpers

func (m *Model) resetInputs() {
	for i := range m.Inputs {
		m.Inputs[i].Reset()
	}
	m.Inputs[0].Focus()
}

func (m *Model) loadInputs(t config.CustomTask) {
	m.Inputs[0].SetValue(t.Name)
	m.Inputs[1].SetValue(t.Command)
	m.Inputs[2].SetValue(strings.Join(t.Args, " "))
}

func (m *Model) saveTask() {
	newTask := config.CustomTask{
		Name:    m.Inputs[0].Value(),
		Command: m.Inputs[1].Value(),
		Args:    strings.Fields(m.Inputs[2].Value()),
	}

	if m.EditMode {
		m.Config.CustomTasks[m.EditIndex] = newTask
		// Update List Item
		m.List.SetItem(m.EditIndex, Item{task: newTask})
	} else {
		m.Config.CustomTasks = append(m.Config.CustomTasks, newTask)
		// Add to List
		m.List.InsertItem(len(m.Config.CustomTasks), Item{task: newTask})
	}

	m.ConfigManager.Save(m.Config)
}
