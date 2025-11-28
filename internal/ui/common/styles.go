package common

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("205")
	ColorSecondary = lipgloss.Color("63")
	ColorSubtle    = lipgloss.Color("241")
	ColorSuccess   = lipgloss.Color("42")
	ColorError     = lipgloss.Color("160")
	ColorHighlight = lipgloss.Color("212")

	// Styles
	TitleStyle = lipgloss.NewStyle().
		MarginLeft(2).
		Bold(true).
		Foreground(ColorPrimary)

	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	ItemStyle     = lipgloss.NewStyle().PaddingLeft(4)
	SelectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(ColorHighlight).SetString("> ")
)
