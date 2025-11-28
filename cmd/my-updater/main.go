package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/app"
	"my-updater/internal/config"
	"my-updater/internal/ui"
)

func main() {
	// 설정 매니저 초기화
	configManager, err := config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// 서비스 초기화
	taskService := app.NewTaskService(configManager)

	// UI 모델 초기화
	model := ui.NewAppModel(taskService)

	// TUI 실행
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
