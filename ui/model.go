package ui

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qyinm/lazyadmin/config"
	"github.com/qyinm/lazyadmin/db"
)

type Focus int

const (
	FocusConnections Focus = iota
	FocusSidebar
	FocusTable
	FocusForm
	FocusConfirm
)

type Mode int

const (
	ModeView Mode = iota
	ModeTableBrowser
	ModeConnectionForm
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

	connSidebar list.Model
	connForm    []textinput.Model
}

func NewModel(cfg *config.Config, database *sql.DB) Model {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(DraculaPurple).
		BorderBottom(true).
		Bold(true).
		Foreground(DraculaCyan).
		Background(DraculaBackground)
	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Reverse(true)
	s.Cell = lipgloss.NewStyle()
	t.SetStyles(s)

	var connItems []list.Item
	for _, c := range cfg.Connections {
		connItems = append(connItems, ViewItem{
			title:       c.Label,
			description: fmt.Sprintf("%s (%s)", c.Driver, c.Host),
			query:       c.Name,
		})
	}
	connList := list.New(connItems, list.NewDefaultDelegate(), 0, 0)
	connList.Title = "Connections"
	connList.SetShowHelp(false)
	connList.SetShowStatusBar(false)
	connList.SetFilteringEnabled(false)
	connList.Styles.Title = TitleStyle

	tableDelegate := list.NewDefaultDelegate()
	tableDelegate.ShowDescription = false
	tableDelegate.Styles.SelectedTitle = tableDelegate.Styles.SelectedTitle.
		Foreground(DraculaPink).
		Background(DraculaBackground).
		Bold(true)
	tableDelegate.Styles.SelectedDesc = tableDelegate.Styles.SelectedDesc.
		Foreground(DraculaComment).
		Background(DraculaBackground)
	tableList := list.New([]list.Item{}, tableDelegate, 0, 0)
	tableList.Title = "Tables"
	tableList.SetShowHelp(false)
	tableList.SetShowStatusBar(false)
	tableList.SetFilteringEnabled(false)
	tableList.Styles.Title = TitleStyle

	inputs := make([]textinput.Model, 8)
	labels := []string{"Label", "Driver", "Host", "Port", "User", "Password", "Database Name", "Path (SQLite)"}
	for i := range inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(DraculaPink)
		t.Prompt = labels[i] + ": "
		t.PromptStyle = lipgloss.NewStyle().Foreground(DraculaCyan)

		if labels[i] == "Port" {
			t.Placeholder = "5432"
		}
		if labels[i] == "Password" {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		inputs[i] = t
	}

	mode := ModeView
	startFocus := FocusConnections
	if database != nil {
		startFocus = FocusSidebar
	}

	return Model{
		config:      cfg,
		db:          database,
		driver:      cfg.Database.Driver,
		sidebar:     tableList,
		table:       t,
		focus:       startFocus,
		mode:        mode,
		tableLoaded: false,
		connSidebar: connList,
		connForm:    inputs,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if m.mode == ModeConnectionForm {
		return m.updateConnectionForm(msg)
	}

	if m.showForm {
		return m.updateForm(msg)
	}

	if m.focus == FocusConfirm {
		return m.updateConfirm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "tab":
			switch m.focus {
			case FocusConnections:
				m.focus = FocusSidebar
			case FocusSidebar:
				m.focus = FocusTable
			case FocusTable:
				m.focus = FocusConnections
			}
			return m, nil

		case "enter":
			if m.focus == FocusConnections {
				return m.handleConnectionSelect()
			}
			if m.focus == FocusSidebar {
				m.focus = FocusTable
				return m.handleSidebarSelect()
			}
			return m, nil

		case "n":
			if m.focus == FocusConnections {
				m.mode = ModeConnectionForm
				m.connForm[0].Focus()
				return m, textinput.Blink
			}

		case "i":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser {
				return m.showInsertForm()
			}

		case "e":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser && m.tableLoaded {
				return m.showEditForm()
			}

		case "d":
			if m.focus == FocusTable && m.currentTable != "" && m.mode == ModeTableBrowser && m.tableLoaded {
				return m.showDeleteConfirm()
			}

		case "r":
			if m.currentTable != "" {
				return m.refreshTable()
			}

		case "t":
			return m.toggleMode()

		case "?":
			m.statusMsg = "Tab: Cycle Focus • Enter: Select • i/e/d: CRUD • n: New Conn"
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePanes()
		return m, nil
	}

	switch m.focus {
	case FocusConnections:
		m.connSidebar, cmd = m.connSidebar.Update(msg)
	case FocusSidebar:
		m.sidebar, cmd = m.sidebar.Update(msg)
	case FocusTable:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m *Model) resizePanes() {
	if m.width == 0 {
		return
	}

	connWidth := m.width * 15 / 100
	if connWidth < 20 {
		connWidth = 20
	}

	tableWidth := m.width * 20 / 100
	if tableWidth < 25 {
		tableWidth = 25
	}

	contentWidth := m.width - connWidth - tableWidth - 4

	panelHeight := m.height - 6
	contentHeight := panelHeight - 2

	m.connSidebar.SetSize(connWidth-4, contentHeight-1)

	m.sidebar.SetSize(tableWidth-4, contentHeight)

	m.table.SetWidth(contentWidth - 4)
	m.table.SetHeight(contentHeight)
}

func (m Model) handleConnectionSelect() (tea.Model, tea.Cmd) {
	index := m.connSidebar.Index()
	if index >= 0 && index < len(m.config.Connections) {
		connConfig := m.config.Connections[index]
		conn, err := db.Connect(&connConfig)
		if err != nil {
			m.err = err
			return m, nil
		}

		if m.db != nil {
			m.db.Close()
		}

		m.db = conn.DB
		m.driver = connConfig.Driver
		m.tables = nil
		m.currentTable = ""
		m.statusMsg = fmt.Sprintf("Connected to %s", connConfig.Label)

		tables, err := db.GetTables(m.db, m.driver)
		if err == nil {
			m.tables = tables
			m.mode = ModeTableBrowser
			m.refreshSidebarList()
		} else {
			m.err = err
		}

		m.focus = FocusSidebar
	}
	return m, nil
}

func (m *Model) refreshSidebarList() {
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
	m.sidebar.SetItems(items)
}

func (m Model) updateConnectionForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = ModeView
			return m, nil
		case "tab", "enter":
			focusedIndex := -1
			for i := range m.connForm {
				if m.connForm[i].Focused() {
					focusedIndex = i
					break
				}
			}

			if msg.String() == "enter" && focusedIndex == len(m.connForm)-1 {
				return m.submitConnectionForm()
			}

			if focusedIndex >= 0 {
				m.connForm[focusedIndex].Blur()
				next := (focusedIndex + 1) % len(m.connForm)
				m.connForm[next].Focus()
			}
			return m, nil
		}
	}

	cmds := make([]tea.Cmd, len(m.connForm))
	for i := range m.connForm {
		m.connForm[i], cmds[i] = m.connForm[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) submitConnectionForm() (tea.Model, tea.Cmd) {
	port, _ := strconv.Atoi(m.connForm[3].Value())

	newConn := config.DatabaseConfig{
		Label:    m.connForm[0].Value(),
		Driver:   m.connForm[1].Value(),
		Host:     m.connForm[2].Value(),
		Port:     port,
		User:     m.connForm[4].Value(),
		Password: m.connForm[5].Value(),
		Name:     m.connForm[6].Value(),
		Path:     m.connForm[7].Value(),
	}

	if newConn.Driver == "" {
		m.err = fmt.Errorf("driver is required")
		return m, nil
	}

	m.config.Connections = append(m.config.Connections, newConn)

	if err := config.Save("admin.yaml", m.config); err != nil {
		m.err = err
		return m, nil
	}

	var items []list.Item
	for _, c := range m.config.Connections {
		items = append(items, ViewItem{
			title:       c.Label,
			description: fmt.Sprintf("%s (%s)", c.Driver, c.Host),
			query:       c.Name,
		})
	}
	m.connSidebar.SetItems(items)

	m.mode = ModeView
	m.statusMsg = "Connection added"

	for i := range m.connForm {
		m.connForm[i].SetValue("")
	}

	return m, nil
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
	if m.db == nil {
		m.err = fmt.Errorf("no database connection")
		return m, nil
	}
	cols, rows, err := db.RunQuery(m.db, query)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	if len(rows) > 0 {
		m.table.SetCursor(0)
	}
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
		if err == nil {
			m.tables = tables
		}
		m.statusMsg = "Table Browser Mode"
	} else {
		m.mode = ModeView
		m.currentTable = ""
		m.statusMsg = "View Mode"
	}

	m.refreshSidebarList()
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

	if m.mode == ModeConnectionForm {
		return AppStyle.Width(m.width).Height(m.height).Render(
			lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				m.viewConnectionForm(),
			),
		)
	}

	if m.showForm {
		formStyle := lipgloss.NewStyle().
			Padding(2, 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DraculaGreen).
			Background(DraculaBackground)

		return AppStyle.Width(m.width).Height(m.height).Render(
			lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				formStyle.Render(m.viewForm()),
			),
		)
	}

	var connBox, sidebarBox, contentBox string
	connWidth := m.width * 15 / 100
	if connWidth < 20 {
		connWidth = 20
	}
	tableWidth := m.width * 20 / 100
	if tableWidth < 25 {
		tableWidth = 25
	}
	contentWidth := m.width - connWidth - tableWidth - 4
	availableHeight := m.height - 6

	connHelp := lipgloss.NewStyle().Render(
		fmt.Sprintf("%s %s",
			HelpKeyStyle.Render("n"),
			HelpDescStyle.Render("new conn"),
		),
	)
	connContent := lipgloss.JoinVertical(lipgloss.Left, m.connSidebar.View(), connHelp)

	if m.focus == FocusConnections {
		connBox = SidebarActiveStyle.Width(connWidth).MaxWidth(connWidth).Height(availableHeight).MaxHeight(availableHeight).Render(connContent)
	} else {
		connBox = SidebarStyle.Width(connWidth).MaxWidth(connWidth).Height(availableHeight).MaxHeight(availableHeight).Render(connContent)
	}

	if m.focus == FocusSidebar {
		sidebarBox = SidebarActiveStyle.Width(tableWidth).MaxWidth(tableWidth).Height(availableHeight).MaxHeight(availableHeight).Render(m.sidebar.View())
	} else {
		sidebarBox = SidebarStyle.Width(tableWidth).MaxWidth(tableWidth).Height(availableHeight).MaxHeight(availableHeight).Render(m.sidebar.View())
	}

	cStyle := ContentStyle
	if m.focus == FocusTable || m.focus == FocusForm || m.focus == FocusConfirm {
		cStyle = ContentActiveStyle
	}

	if m.focus == FocusConfirm {
		contentBox = cStyle.Width(contentWidth).MaxWidth(contentWidth).Height(availableHeight).MaxHeight(availableHeight).Render(m.renderConfirm())
	} else {
		contentBox = cStyle.Width(contentWidth).MaxWidth(contentWidth).Height(availableHeight).MaxHeight(availableHeight).Render(m.renderContent())
	}

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, connBox, sidebarBox, contentBox)

	statusStyle := lipgloss.NewStyle().
		Foreground(DraculaComment).
		Padding(0, 1)

	status := fmt.Sprintf("%s %s | %s | ?: Help", "[Tables]", m.currentTable, m.statusMsg)
	if m.err != nil {
		status = fmt.Sprintf("❌ %s", m.err.Error())
	}

	finalView := mainView + "\n" + statusStyle.Render(status)
	return AppStyle.Width(m.width).Height(m.height).Render(finalView)
}

func (m Model) viewConnectionForm() string {
	var b string
	b += TitleStyle.Render("New Connection") + "\n\n"

	for i := range m.connForm {
		b += m.connForm[i].View() + "\n"
	}

	b += "\n" + EmptyStateStyle.Render("Enter: Save • Tab: Next • Esc: Cancel")
	return b
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
		Background(DraculaBackground).
		Bold(true).
		Padding(2)

	return confirmStyle.Render(m.confirmMsg)
}

func (m Model) viewForm() string {
	return m.form.View()
}
