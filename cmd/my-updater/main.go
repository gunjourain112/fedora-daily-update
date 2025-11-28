package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"my-updater/internal/ui/root"
)

func main() {
	model, err := root.NewModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
