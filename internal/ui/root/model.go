package root

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/config"
	"my-updater/internal/domain"
	"my-updater/internal/ui/menu"
	"my-updater/internal/ui/runner"
	"my-updater/internal/ui/settings"
	"my-updater/internal/ui/updatelist"
)

type State int

const (
	StateMenu State = iota
	StateUpdateList
	StateRunner
	StateSettings
)

type Model struct {
	state         State
	configManager *config.Manager
	taskManager   *domain.TaskManager
	config        *config.Config

	menuModel       menu.Model
	updateListModel updatelist.Model
	runnerModel     runner.Model
	settingsModel   settings.Model
}

func NewModel() (Model, error) {
	cm, err := config.NewManager()
	if err != nil {
		return Model{}, fmt.Errorf("failed to init config manager: %w", err)
	}

	tm := domain.NewTaskManager()

	return Model{
		state:         StateMenu,
		configManager: cm,
		taskManager:   tm,
		menuModel:     menu.NewModel(),
	}, nil
}

func (m Model) Init() tea.Cmd {
	return m.menuModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Global Quit handler can be tricky with nested inputs,
	// but Ctrl+C is usually safe to catch at root if not handled by child.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateMenu:
		var menuMsg tea.Model
		menuMsg, cmd = m.menuModel.Update(msg)
		m.menuModel = menuMsg.(menu.Model)

		if m.menuModel.Quitting {
			return m, tea.Quit
		}

		if m.menuModel.Selected != -1 {
			selection := m.menuModel.Selected
			m.menuModel.Selected = -1 // Reset for next time

			if selection == menu.ChoiceUpdate { // "Update List"
				cfg, _ := m.configManager.Load()
				m.config = cfg
				m.updateListModel = updatelist.NewModel(m.taskManager, m.config)
				m.state = StateUpdateList
				return m, m.updateListModel.Init()

			} else if selection == menu.ChoiceSettings { // "Settings"
				var err error
				m.settingsModel, err = settings.NewModel(m.configManager)
				if err != nil {
					return m, tea.Quit
				}
				m.state = StateSettings
				return m, m.settingsModel.Init()
			}
		}

	case StateUpdateList:
		var ulMsg tea.Model
		ulMsg, cmd = m.updateListModel.Update(msg)
		m.updateListModel = ulMsg.(updatelist.Model)

		if m.updateListModel.Exit {
			m.state = StateMenu
			return m, nil
		}

		if m.updateListModel.SelectedTask != nil {
			// Transition to Runner
			task := *m.updateListModel.SelectedTask
			m.updateListModel.SelectedTask = nil // Reset
			m.runnerModel = runner.NewModel(task)
			m.state = StateRunner
			return m, m.runnerModel.Init()
		}

	case StateRunner:
		var rMsg tea.Model
		rMsg, cmd = m.runnerModel.Update(msg)
		m.runnerModel = rMsg.(runner.Model)

		if m.runnerModel.Exit {
			// Back to Update List
			m.state = StateUpdateList
			// We might want to refresh the list or update status, but for now just go back.
			return m, nil
		}

	case StateSettings:
		var sMsg tea.Model
		sMsg, cmd = m.settingsModel.Update(msg)
		m.settingsModel = sMsg.(settings.Model)

		if m.settingsModel.Exit {
			m.state = StateMenu
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateUpdateList:
		return m.updateListModel.View()
	case StateRunner:
		return m.runnerModel.View()
	case StateSettings:
		return m.settingsModel.View()
	}
	return ""
}
