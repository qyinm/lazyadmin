package ui

import "github.com/charmbracelet/lipgloss"

var (
	DraculaBackground = lipgloss.Color("0")
	DraculaForeground = lipgloss.Color("15")
	DraculaPurple     = lipgloss.Color("5")
	DraculaPink       = lipgloss.Color("13")
	DraculaCyan       = lipgloss.Color("14")
	DraculaGreen      = lipgloss.Color("10")
	DraculaComment    = lipgloss.Color("7")

	SidebarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaPurple).
			Background(DraculaBackground)

	SidebarActiveStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.ThickBorder()).
				BorderForeground(DraculaPink).
				Background(DraculaBackground)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaCyan).
			Background(DraculaBackground)

	ContentActiveStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.ThickBorder()).
				BorderForeground(DraculaPink).
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

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(DraculaPink).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true)
)
