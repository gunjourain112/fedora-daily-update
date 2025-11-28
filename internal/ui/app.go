package ui

import (
	"my-updater/internal/app"
	"my-updater/internal/ui/menu"
	"my-updater/internal/ui/settings"
	"my-updater/internal/ui/updater"

	tea "github.com/charmbracelet/bubbletea"
)

// AppState는 애플리케이션의 현재 화면 상태를 나타냅니다.
type AppState int

const (
	StateMenu AppState = iota
	StateUpdater
	StateSettings
)

type AppModel struct {
	state       AppState
	taskService *app.TaskService

	menuModel     menu.Model
	updaterModel  updater.Model
	settingsModel settings.Model
}

func NewAppModel(ts *app.TaskService) AppModel {
	return AppModel{
		state:       StateMenu,
		taskService: ts,
		menuModel:   menu.NewModel(),
		// updaterModel과 settingsModel은 필요할 때 초기화하거나 여기서 기본값으로 생성
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.menuModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// 글로벌 키 처리 (예: Ctrl+C)
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateMenu:
		newMenu, menuCmd := m.menuModel.Update(msg)
		m.menuModel = newMenu.(menu.Model)
		cmd = menuCmd

		if m.menuModel.Chosen {
			m.menuModel.Chosen = false // 리셋
			switch m.menuModel.Selected {
			case menu.ChoiceUpdate:
				tasks, _ := m.taskService.GetAllTasks() // 에러 처리 필요
				m.updaterModel = updater.NewModel(tasks)
				m.state = StateUpdater
				return m, m.updaterModel.Init()
			case menu.ChoiceSettings:
				tasks, _ := m.taskService.GetAllTasks()
				m.settingsModel = settings.NewModel(tasks, m.taskService)
				m.state = StateSettings
				return m, m.settingsModel.Init()
			}
		}

	case StateUpdater:
		newUpdater, updaterCmd := m.updaterModel.Update(msg)
		m.updaterModel = newUpdater.(updater.Model)
		cmd = updaterCmd

		if m.updaterModel.Exit {
			m.state = StateMenu
		}

	case StateSettings:
		newSettings, settingsCmd := m.settingsModel.Update(msg)
		m.settingsModel = newSettings.(settings.Model)
		cmd = settingsCmd

		if m.settingsModel.Exit {
			m.state = StateMenu
		}
	}

	return m, cmd
}

func (m AppModel) View() string {
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
