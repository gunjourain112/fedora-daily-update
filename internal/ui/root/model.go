package root

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/config"
	"my-updater/internal/domain"
	"my-updater/internal/ui/menu"
	"my-updater/internal/ui/settings"
	"my-updater/internal/ui/updater"
)

type State int

const (
	StateMenu State = iota
	StateUpdater
	StateSettings
)

type Model struct {
	state         State
	configManager *config.Manager
	taskManager   *domain.TaskManager

	menuModel     menu.Model
	updaterModel  updater.Model
	settingsModel settings.Model
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

	// Global Quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "ctrl+c" {
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
			if m.menuModel.Selected == menu.ChoiceUpdate {
				// Initialize Updater
				cfg, _ := m.configManager.Load() // Reload config in case it changed
				tasks := []domain.Task{}

				// Convert config tasks to domain tasks
				customTasks := []domain.Task{}
				for _, ct := range cfg.CustomTasks {
					customTasks = append(customTasks, domain.Task{
						Name:    ct.Name,
						Command: ct.Command,
						Args:    ct.Args,
					})
				}
				tasks = m.taskManager.GetTasks(customTasks)

				m.updaterModel = updater.NewModel(tasks)
				m.state = StateUpdater
				return m, m.updaterModel.Init()

			} else if m.menuModel.Selected == menu.ChoiceSettings {
				// Initialize Settings
				var err error
				m.settingsModel, err = settings.NewModel(m.configManager)
				if err != nil {
					// Handle error gracefully? For now exit
					fmt.Println("Error loading settings:", err)
					return m, tea.Quit
				}
				m.state = StateSettings
				// Reset selection in menu so when we come back it's clean
				m.menuModel.Selected = -1
				return m, m.settingsModel.Init()
			}
		}

	case StateUpdater:
		var updaterMsg tea.Model
		updaterMsg, cmd = m.updaterModel.Update(msg)
		m.updaterModel = updaterMsg.(updater.Model)

		if m.updaterModel.Exit {
			m.state = StateMenu
			m.menuModel.Selected = -1 // Reset selection
			return m, nil
		}

	case StateSettings:
		var settingsMsg tea.Model
		settingsMsg, cmd = m.settingsModel.Update(msg)
		m.settingsModel = settingsMsg.(settings.Model)

		if m.settingsModel.Exit {
			m.state = StateMenu
			m.menuModel.Selected = -1 // Reset selection
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateUpdater:
		return m.updaterModel.View()
	case StateSettings:
		return m.settingsModel.View()
	}
	return ""
}
