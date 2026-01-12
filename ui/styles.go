package ui

import "github.com/charmbracelet/lipgloss"

var (
	DraculaBackground = lipgloss.Color("#282a36")
	DraculaForeground = lipgloss.Color("#f8f8f2")
	DraculaPurple     = lipgloss.Color("#bd93f9")
	DraculaPink       = lipgloss.Color("#ff79c6")
	DraculaCyan       = lipgloss.Color("#8be9fd")
	DraculaGreen      = lipgloss.Color("#50fa7b")
	DraculaComment    = lipgloss.Color("#6272a4")

	SidebarStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaPurple).
			Background(DraculaBackground)

	SidebarActiveStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(DraculaGreen).
				Background(DraculaBackground)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaCyan).
			Background(DraculaBackground)

	ContentActiveStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(DraculaGreen).
				Background(DraculaBackground)

	TitleStyle = lipgloss.NewStyle().
			Foreground(DraculaPink).
			Bold(true).
			Padding(0, 1)

	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(DraculaComment).
			Italic(true).
			Padding(2, 4)

	AppStyle = lipgloss.NewStyle().
			Background(DraculaBackground).
			Foreground(DraculaForeground)
)
