package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"my-updater/internal/config"
	"my-updater/internal/ui/common"
)

// ViewState tracks which sub-view we are in.
type ViewState int

const (
	StateList       ViewState = iota // Main list of tasks + "Add New"
	StateTaskAction                  // Menu for a selected task (Edit/Delete)
	StateInputName
	StateInputCommand
	StateInputArgs
)

// Item wraps the config task for the list bubble.
type Item struct {
	task   *config.CustomTask // Pointer so we can check if it's nil (Add New button)
	title  string
	desc   string
	isAdd  bool
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	ConfigManager *config.Manager
	Config        *config.Config

	// List View
	List list.Model

	// Task Action View (Simple selection)
	ActionIndex int // 0: Edit, 1: Delete, 2: Back

	// Input Form
	Inputs     []textinput.Model
	FocusIndex int

	// State
	State ViewState

	// Temp storage
	SelectedTaskIndex int // Index in Config.CustomTasks
	EditMode          bool

	Exit bool // Signal to parent
}

func NewModel(cm *config.Manager) (Model, error) {
	cfg, err := cm.Load()
	if err != nil {
		return Model{}, err
	}

	// Inputs setup
	inputs := make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "메뉴 이름 (예: 내 스크립트)"
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

	m := Model{
		ConfigManager: cm,
		Config:        cfg,
		Inputs:        inputs,
		State:         StateList,
	}

	m.refreshList()
	return m, nil
}

func (m *Model) refreshList() {
	items := []list.Item{}
	for i := range m.Config.CustomTasks {
		// Use local variable to avoid pointer issues in loop
		t := m.Config.CustomTasks[i]
		items = append(items, Item{
			task:  &t,
			title: t.Name,
			desc:  fmt.Sprintf("%s %s", t.Command, strings.Join(t.Args, " ")),
			isAdd: false,
		})
	}
	// Add "Add New" item
	items = append(items, Item{
		task:  nil,
		title: "➕ 새 항목 추가",
		desc:  "새로운 커스텀 업데이트 명령을 추가합니다.",
		isAdd: true,
	})

	m.List = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.List.Title = "설정 - 커스텀 메뉴 관리"
	m.List.SetShowHelp(false) // Clean view
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := common.DocStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch m.State {
		case StateList:
			switch msg.String() {
			case "esc", "q":
				m.Exit = true
				return m, nil
			case "enter":
				selectedItem := m.List.SelectedItem().(Item)
				if selectedItem.isAdd {
					// Go to Add Flow
					m.State = StateInputName
					m.EditMode = false
					m.resetInputs()
					return m, textinput.Blink
				} else {
					// Go to Action Menu for selected task
					m.State = StateTaskAction
					m.SelectedTaskIndex = m.List.Index()
					m.ActionIndex = 0
				}
			}

		case StateTaskAction:
			switch msg.String() {
			case "esc", "q":
				m.State = StateList
			case "up", "k":
				if m.ActionIndex > 0 {
					m.ActionIndex--
				}
			case "down", "j":
				if m.ActionIndex < 2 {
					m.ActionIndex++
				}
			case "enter":
				switch m.ActionIndex {
				case 0: // Edit
					m.State = StateInputName
					m.EditMode = true
					m.loadInputs(m.Config.CustomTasks[m.SelectedTaskIndex])
					return m, textinput.Blink
				case 1: // Delete
					// Delete logic
					m.Config.CustomTasks = append(m.Config.CustomTasks[:m.SelectedTaskIndex], m.Config.CustomTasks[m.SelectedTaskIndex+1:]...)
					m.ConfigManager.Save(m.Config)
					m.refreshList()
					m.State = StateList
				case 2: // Back
					m.State = StateList
				}
			}

		case StateInputName, StateInputCommand, StateInputArgs:
			switch msg.String() {
			case "esc":
				m.State = StateList
				return m, nil
			case "enter":
				if m.FocusIndex < len(m.Inputs)-1 {
					m.Inputs[m.FocusIndex].Blur()
					m.FocusIndex++
					m.Inputs[m.FocusIndex].Focus()
					return m, textinput.Blink
				} else {
					m.Inputs[m.FocusIndex].Blur()
					m.saveTask()
					m.State = StateList
					m.refreshList()
					return m, nil
				}
			}
		}
	}

	// Sub-model updates
	if m.State == StateList {
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.State >= StateInputName {
		for i := range m.Inputs {
			m.Inputs[i], cmd = m.Inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.State == StateList {
		return common.DocStyle.Render(m.List.View())
	}

	if m.State == StateTaskAction {
		title := m.Config.CustomTasks[m.SelectedTaskIndex].Name
		opts := []string{"✏️ 수정 (Edit)", "🗑️ 삭제 (Delete)", "↩️ 뒤로가기 (Back)"}

		s := fmt.Sprintf("\n%s\n\n", common.TitleStyle.Render("'" + title + "' 관리"))
		for i, opt := range opts {
			cursor := " "
			if i == m.ActionIndex {
				cursor = common.SelectedStyle.String()
				opt = lipgloss.NewStyle().Bold(true).Foreground(common.ColorHighlight).Render(opt)
			} else {
				cursor = common.ItemStyle.String()
			}
			s += fmt.Sprintf("%s%s\n", cursor, opt)
		}
		return common.DocStyle.Render(s)
	}

	// Input Form
	var b strings.Builder
	title := "새 메뉴 추가"
	if m.EditMode {
		title = "메뉴 수정"
	}
	b.WriteString(fmt.Sprintf("\n%s\n\n", common.TitleStyle.Render(title)))

	labels := []string{"이름", "명령어", "인자"}
	for i := range m.Inputs {
		b.WriteString(fmt.Sprintf("  %s:\n", labels[i]))
		b.WriteString(common.DocStyle.Render(m.Inputs[i].View()))
		b.WriteString("\n")
	}
	b.WriteString("\n  (Enter: 다음/저장, Esc: 취소)\n")
	return b.String()
}

func (m *Model) resetInputs() {
	for i := range m.Inputs {
		m.Inputs[i].Reset()
	}
	m.FocusIndex = 0
	m.Inputs[0].Focus()
}

func (m *Model) loadInputs(t config.CustomTask) {
	m.Inputs[0].SetValue(t.Name)
	m.Inputs[1].SetValue(t.Command)
	m.Inputs[2].SetValue(strings.Join(t.Args, " "))
	m.FocusIndex = 0
	m.Inputs[0].Focus()
}

func (m *Model) saveTask() {
	newTask := config.CustomTask{
		Name:    m.Inputs[0].Value(),
		Command: m.Inputs[1].Value(),
		Args:    strings.Fields(m.Inputs[2].Value()),
	}

	if m.EditMode {
		m.Config.CustomTasks[m.SelectedTaskIndex] = newTask
	} else {
		m.Config.CustomTasks = append(m.Config.CustomTasks, newTask)
	}
	m.ConfigManager.Save(m.Config)
}
