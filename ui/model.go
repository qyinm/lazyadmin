package ui

import (
	"database/sql"
	"errors"
	"fmt"

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
	FocusForm
	FocusConfirm
)

type Mode int

const (
	ModeView Mode = iota
	ModeTableBrowser
)

type Model struct {
	config        *config.Config
	db            *sql.DB
	driver        string
	sidebar       list.Model
	table         table.Model
	focus         Focus
	mode          Mode
	width         int
	height        int
	tableLoaded   bool
	err           error
	statusMsg     string
	currentTable  string
	pkColumn      string
	columns       []db.ColumnInfo
	form          FormModel
	showForm      bool
	confirmMsg    string
	confirmAction func()
	tables        []db.TableInfo
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
		driver:      cfg.Database.Driver,
		sidebar:     list.Model{},
		table:       t,
		focus:       FocusSidebar,
		mode:        ModeView,
		tableLoaded: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.showForm {
		return m.updateForm(msg)
	}

	if m.focus == FocusConfirm {
		return m.updateConfirm(msg)
	}

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
				return m.handleSidebarSelect()
			}
			return m, nil

		case "i":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser {
				return m.showInsertForm()
			}
			return m, nil

		case "e":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser && m.tableLoaded {
				return m.showEditForm()
			}
			return m, nil

		case "d":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser && m.tableLoaded {
				return m.showDeleteConfirm()
			}
			return m, nil

		case "r":
			if m.currentTable != "" {
				return m.refreshTable()
			}
			return m, nil

		case "t":
			return m.toggleMode()

		case "?":
			m.statusMsg = "i:Insert e:Edit d:Delete r:Refresh t:Toggle Mode Tab:Switch Focus q:Quit"
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		sidebarWidth := m.width * 30 / 100
		contentWidth := m.width - sidebarWidth - 8

		if m.mode == ModeTableBrowser && m.tables == nil {
			tables, err := db.GetTables(m.db, m.driver)
			if err == nil {
				m.tables = tables
			}
		}

		m.sidebar = m.buildSidebar(sidebarWidth-4, m.height-6)
		m.sidebar.Title = m.config.ProjectName
		m.sidebar.Styles.Title = TitleStyle

		m.table.SetWidth(contentWidth - 4)
		m.table.SetHeight(m.height - 10)

		return m, nil
	}

	if m.focus == FocusSidebar {
		m.sidebar, cmd = m.sidebar.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) buildSidebar(width, height int) list.Model {
	var items []list.Item

	if m.mode == ModeTableBrowser {
		for _, t := range m.tables {
			items = append(items, ViewItem{
				title:       "📋 " + t.Name,
				description: "Browse table",
				query:       t.Name,
				isTable:     true,
			})
		}
	} else {
		for _, v := range m.config.Views {
			items = append(items, ViewItem{
				title:       v.Title,
				description: v.Description,
				query:       v.Query,
				isTable:     false,
			})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return l
}

func (m Model) handleSidebarSelect() (tea.Model, tea.Cmd) {
	item, ok := m.sidebar.SelectedItem().(ViewItem)
	if !ok {
		return m, nil
	}

	if item.isTable {
		validTable := false
		if m.tables == nil {
			tables, err := db.GetTables(m.db, m.driver)
			if err == nil {
				m.tables = tables
			}
		}

		for _, t := range m.tables {
			if t.Name == item.query {
				validTable = true
				break
			}
		}

		if !validTable {
			m.err = fmt.Errorf("invalid table name: %s", item.query)
			return m, nil
		}

		m.currentTable = item.query
		m.mode = ModeTableBrowser

		columns, err := db.GetColumns(m.db, m.driver, m.currentTable)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.columns = columns

		pkCol, err := db.GetPrimaryKey(m.db, m.driver, m.currentTable)
		if err != nil {
			if errors.Is(err, db.ErrNoPrimaryKey) {
				m.statusMsg = fmt.Sprintf("Warning: %s has no primary key", m.currentTable)
			} else {
				m.err = err
				return m, nil
			}
		}
		m.pkColumn = pkCol

		query := db.BuildSelectAllQuery(m.driver, m.currentTable, 100)
		return m.executeQuery(query)
	}

	return m.executeQuery(item.Query())
}

func (m Model) executeQuery(query string) (tea.Model, tea.Cmd) {
	cols, rows, err := db.RunQuery(m.db, query)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.tableLoaded = true
	m.err = nil
	m.statusMsg = fmt.Sprintf("Loaded %d rows", len(rows))

	return m, nil
}

func (m Model) refreshTable() (tea.Model, tea.Cmd) {
	if m.currentTable == "" {
		return m, nil
	}
	query := db.BuildSelectAllQuery(m.driver, m.currentTable, 100)
	m.statusMsg = "Refreshing..."
	return m.executeQuery(query)
}

func (m Model) toggleMode() (tea.Model, tea.Cmd) {
	if m.mode == ModeView {
		m.mode = ModeTableBrowser
		tables, err := db.GetTables(m.db, m.driver)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.tables = tables
		m.statusMsg = "Table Browser Mode"
	} else {
		m.mode = ModeView
		m.currentTable = ""
		m.statusMsg = "View Mode"
	}

	m.sidebar = m.buildSidebar(m.width*30/100-4, m.height-6)
	m.sidebar.Title = m.config.ProjectName
	m.sidebar.Styles.Title = TitleStyle
	m.tableLoaded = false

	return m, nil
}

func (m Model) showInsertForm() (tea.Model, tea.Cmd) {
	m.form = NewFormModel(m.columns, FormModeInsert, m.currentTable, m.pkColumn, nil, nil)
	m.showForm = true
	m.focus = FocusForm
	return m, m.form.Init()
}

func (m Model) showEditForm() (tea.Model, tea.Cmd) {
	if m.pkColumn == "" {
		m.err = fmt.Errorf("no primary key found for table %s", m.currentTable)
		return m, nil
	}

	selectedRow := m.table.SelectedRow()
	if selectedRow == nil {
		m.err = fmt.Errorf("no row selected")
		return m, nil
	}

	cols := m.table.Columns()
	var pkValue interface{}
	for i, col := range cols {
		if col.Title == m.pkColumn {
			pkValue = selectedRow[i]
			break
		}
	}

	if pkValue == nil {
		m.err = fmt.Errorf("could not find primary key value")
		return m, nil
	}

	record, err := db.GetRecordByPK(m.db, m.driver, m.currentTable, m.pkColumn, pkValue)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.form = NewFormModel(m.columns, FormModeEdit, m.currentTable, m.pkColumn, pkValue, record)
	m.showForm = true
	m.focus = FocusForm
	return m, m.form.Init()
}

func (m Model) showDeleteConfirm() (tea.Model, tea.Cmd) {
	if m.pkColumn == "" {
		m.err = fmt.Errorf("no primary key found for table %s", m.currentTable)
		return m, nil
	}

	selectedRow := m.table.SelectedRow()
	if selectedRow == nil {
		m.err = fmt.Errorf("no row selected")
		return m, nil
	}

	cols := m.table.Columns()
	var pkValue interface{}
	for i, col := range cols {
		if col.Title == m.pkColumn {
			pkValue = selectedRow[i]
			break
		}
	}

	m.confirmMsg = fmt.Sprintf("Delete record with %s = %v? (y/n)", m.pkColumn, pkValue)
	m.confirmAction = func() {
		err := db.DeleteRecord(m.db, m.driver, m.currentTable, m.pkColumn, pkValue)
		if err != nil {
			m.err = err
			m.statusMsg = "Delete failed: " + err.Error()
		} else {
			m.statusMsg = "Record deleted successfully"
		}
	}
	m.focus = FocusConfirm

	return m, nil
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.IsCancelled() {
		m.showForm = false
		m.focus = FocusTable
		m.statusMsg = "Cancelled"
		return m, nil
	}

	if m.form.IsSubmitted() {
		if m.form.mode == FormModeInsert {
			data, err := m.form.GetData()
			if err != nil {
				m.err = err
				m.statusMsg = "Validation failed: " + err.Error()
				m.showForm = false
				m.focus = FocusTable
				return m, nil
			}
			if len(data) > 0 {
				err := db.InsertRecord(m.db, m.driver, m.currentTable, data)
				if err != nil {
					m.err = err
					m.statusMsg = "Insert failed: " + err.Error()
				} else {
					m.statusMsg = "Record inserted successfully"
				}
			}
		} else {
			data := m.form.GetChangedData()
			if len(data) > 0 {
				err := db.UpdateRecord(m.db, m.driver, m.currentTable, m.pkColumn, m.form.pkValue, data)
				if err != nil {
					m.err = err
					m.statusMsg = "Update failed: " + err.Error()
				} else {
					m.statusMsg = "Record updated successfully"
				}
			} else {
				m.statusMsg = "No changes made"
			}
		}

		m.showForm = false
		m.focus = FocusTable

		return m.refreshTable()
	}

	return m, cmd
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.confirmAction != nil {
				m.confirmAction()
			}
			m.focus = FocusTable
			m.confirmMsg = ""
			m.confirmAction = nil
			return m.refreshTable()

		case "n", "N", "esc":
			m.focus = FocusTable
			m.confirmMsg = ""
			m.confirmAction = nil
			m.statusMsg = "Cancelled"
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showForm {
		formStyle := lipgloss.NewStyle().
			Padding(2, 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaGreen)

		return formStyle.Render(m.form.View())
	}

	sidebarWidth := m.width * 30 / 100
	contentWidth := m.width - sidebarWidth - 8

	var sidebarBox, contentBox string

	modeIndicator := "[View]"
	if m.mode == ModeTableBrowser {
		modeIndicator = "[Tables]"
	}

	if m.focus == FocusSidebar {
		sidebarBox = SidebarActiveStyle.Width(sidebarWidth).Height(m.height - 6).Render(m.sidebar.View())
		contentBox = ContentStyle.Width(contentWidth).Height(m.height - 6).Render(m.renderContent())
	} else if m.focus == FocusConfirm {
		sidebarBox = SidebarStyle.Width(sidebarWidth).Height(m.height - 6).Render(m.sidebar.View())
		contentBox = ContentActiveStyle.Width(contentWidth).Height(m.height - 6).Render(m.renderConfirm())
	} else {
		sidebarBox = SidebarStyle.Width(sidebarWidth).Height(m.height - 6).Render(m.sidebar.View())
		contentBox = ContentActiveStyle.Width(contentWidth).Height(m.height - 6).Render(m.renderContent())
	}

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, contentBox)

	statusStyle := lipgloss.NewStyle().
		Foreground(DraculaComment).
		Padding(0, 1)

	status := fmt.Sprintf("%s %s | %s | ?: Help", modeIndicator, m.currentTable, m.statusMsg)
	if m.err != nil {
		status = fmt.Sprintf("❌ %s", m.err.Error())
	}

	return mainView + "\n" + statusStyle.Render(status)
}

func (m Model) renderContent() string {
	if m.err != nil {
		return EmptyStateStyle.Render("Error: " + m.err.Error())
	}
	if !m.tableLoaded {
		if m.mode == ModeTableBrowser {
			return EmptyStateStyle.Render("Select a table to browse data\n\ni: Insert  e: Edit  d: Delete  r: Refresh")
		}
		return EmptyStateStyle.Render("Select a menu item to view data")
	}
	return m.table.View()
}

func (m Model) renderConfirm() string {
	confirmStyle := lipgloss.NewStyle().
		Foreground(DraculaPink).
		Bold(true).
		Padding(2)

	return confirmStyle.Render(m.confirmMsg)
}
