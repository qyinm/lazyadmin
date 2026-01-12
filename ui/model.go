package ui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qyinm/lazyadmin/config"
	"github.com/qyinm/lazyadmin/db"
)

type Focus int

const (
	FocusSidebar Focus = iota
	FocusTable
)

type Model struct {
	config      *config.Config
	db          *sql.DB
	sidebar     list.Model
	table       table.Model
	focus       Focus
	width       int
	height      int
	tableLoaded bool
	err         error
}

func NewModel(cfg *config.Config, database *sql.DB) Model {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(DraculaPurple).
		BorderBottom(true).
		Bold(true).
		Foreground(DraculaCyan)
	s.Selected = s.Selected.
		Foreground(DraculaBackground).
		Background(DraculaPurple).
		Bold(false)
	t.SetStyles(s)

	return Model{
		config:      cfg,
		db:          database,
		sidebar:     list.Model{},
		table:       t,
		focus:       FocusSidebar,
		tableLoaded: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focus == FocusSidebar {
				m.focus = FocusTable
				m.table.Focus()
			} else {
				m.focus = FocusSidebar
				m.table.Blur()
			}
			return m, nil
		case "enter":
			if m.focus == FocusSidebar {
				if item, ok := m.sidebar.SelectedItem().(ViewItem); ok {
					cols, rows, err := db.RunQuery(m.db, item.Query())
					if err != nil {
						m.err = err
						return m, nil
					}
					m.table.SetRows([]table.Row{})
					m.table.SetColumns(cols)
					m.table.SetRows(rows)
					m.tableLoaded = true
					m.err = nil
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		sidebarWidth := m.width * 30 / 100
		contentWidth := m.width - sidebarWidth - 8

		m.sidebar = NewSidebar(m.config.Views, sidebarWidth-4, m.height-6)
		m.sidebar.Title = m.config.ProjectName
		m.sidebar.Styles.Title = TitleStyle

		m.table.SetWidth(contentWidth - 4)
		m.table.SetHeight(m.height - 8)

		return m, nil
	}

	if m.focus == FocusSidebar {
		m.sidebar, cmd = m.sidebar.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	sidebarWidth := m.width * 30 / 100
	contentWidth := m.width - sidebarWidth - 8

	var sidebarBox, contentBox string

	if m.focus == FocusSidebar {
		sidebarBox = SidebarActiveStyle.Width(sidebarWidth).Height(m.height - 4).Render(m.sidebar.View())
		contentBox = ContentStyle.Width(contentWidth).Height(m.height - 4).Render(m.renderContent())
	} else {
		sidebarBox = SidebarStyle.Width(sidebarWidth).Height(m.height - 4).Render(m.sidebar.View())
		contentBox = ContentActiveStyle.Width(contentWidth).Height(m.height - 4).Render(m.renderContent())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, contentBox)
}

func (m Model) renderContent() string {
	if m.err != nil {
		return EmptyStateStyle.Render("Error: " + m.err.Error())
	}
	if !m.tableLoaded {
		return EmptyStateStyle.Render("Select a menu item to view data")
	}
	return m.table.View()
}
