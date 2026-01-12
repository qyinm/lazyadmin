package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qyinm/lazyadmin/db"
)

type FormMode int

const (
	FormModeInsert FormMode = iota
	FormModeEdit
)

type FormField struct {
	Input    textinput.Model
	Column   db.ColumnInfo
	Original string
}

type FormModel struct {
	fields     []FormField
	focusIndex int
	mode       FormMode
	tableName  string
	pkColumn   string
	pkValue    interface{}
	width      int
	height     int
	submitted  bool
	cancelled  bool
}

func NewFormModel(columns []db.ColumnInfo, mode FormMode, tableName, pkColumn string, pkValue interface{}, existingData map[string]interface{}) FormModel {
	fields := make([]FormField, 0, len(columns))

	for _, col := range columns {
		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("%s (%s)", col.Name, col.Type)
		ti.CharLimit = 500
		ti.Width = 50

		var original string
		if existingData != nil {
			if val, ok := existingData[col.Name]; ok && val != nil {
				original = toString(val)
				ti.SetValue(original)
			}
		}

		if col.PrimaryKey && mode == FormModeEdit {
			ti.Placeholder += " [PK]"
		}

		fields = append(fields, FormField{
			Input:    ti,
			Column:   col,
			Original: original,
		})
	}

	if len(fields) > 0 {
		fields[0].Input.Focus()
	}

	return FormModel{
		fields:     fields,
		focusIndex: 0,
		mode:       mode,
		tableName:  tableName,
		pkColumn:   pkColumn,
		pkValue:    pkValue,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= len(m.fields) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.fields) - 1
			}
			return m, m.updateFocus()

		case "enter":
			if m.focusIndex == len(m.fields)-1 {
				m.submitted = true
				return m, nil
			}
			m.focusIndex++
			if m.focusIndex >= len(m.fields) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "ctrl+s":
			m.submitted = true
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *FormModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		if i == m.focusIndex {
			cmds[i] = m.fields[i].Input.Focus()
		} else {
			m.fields[i].Input.Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		m.fields[i].Input, cmds[i] = m.fields[i].Input.Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var b strings.Builder

	title := "➕ Insert New Record"
	if m.mode == FormModeEdit {
		title = "✏️ Edit Record"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(DraculaPink).
		MarginBottom(1)

	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Width(20).
		Foreground(DraculaCyan)

	pkStyle := lipgloss.NewStyle().
		Foreground(DraculaComment).
		Italic(true)

	for i, field := range m.fields {
		label := labelStyle.Render(field.Column.Name + ":")

		cursor := "  "
		if i == m.focusIndex {
			cursor = "▸ "
		}

		extra := ""
		if field.Column.PrimaryKey {
			extra = pkStyle.Render(" [PK]")
		}
		if field.Column.Nullable {
			extra += pkStyle.Render(" [NULL OK]")
		}

		b.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, label, field.Input.View(), extra))
	}

	b.WriteString("\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(DraculaComment)

	b.WriteString(helpStyle.Render("Tab/↓↑: Navigate • Ctrl+S/Enter: Save • Esc: Cancel"))

	return b.String()
}

func (m FormModel) IsSubmitted() bool {
	return m.submitted
}

func (m FormModel) IsCancelled() bool {
	return m.cancelled
}

func (m FormModel) GetData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var missingRequired []string

	for _, field := range m.fields {
		value := strings.TrimSpace(field.Input.Value())

		if value == "" {
			if !field.Column.Nullable && !field.Column.Default.Valid {
				missingRequired = append(missingRequired, field.Column.Name)
			}
			continue
		}

		data[field.Column.Name] = value
	}

	if len(missingRequired) > 0 {
		return nil, fmt.Errorf("required fields missing: %s", strings.Join(missingRequired, ", "))
	}

	return data, nil
}

func (m FormModel) GetChangedData() map[string]interface{} {
	data := make(map[string]interface{})
	for _, field := range m.fields {
		if field.Column.PrimaryKey {
			continue
		}

		value := strings.TrimSpace(field.Input.Value())
		if value != field.Original {
			if value == "" && field.Column.Nullable {
				data[field.Column.Name] = nil
			} else {
				data[field.Column.Name] = value
			}
		}
	}
	return data
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
