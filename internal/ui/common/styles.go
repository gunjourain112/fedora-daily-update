package common

import "github.com/charmbracelet/lipgloss"

var (
	// 기본 색상 팔레트
	ColorPrimary   = lipgloss.Color("205") // Pink
	ColorSecondary = lipgloss.Color("63")  // Purple
	ColorText      = lipgloss.Color("252") // White-ish
	ColorSubtle    = lipgloss.Color("241") // Grey

	// 스타일 정의
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(1, 0, 1, 0)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(ColorPrimary).
				Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			MarginTop(1)
)
