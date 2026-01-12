package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/qyinm/lazyadmin/config"
)

type ViewItem struct {
	title       string
	description string
	query       string
	isTable     bool
}

func (v ViewItem) Title() string       { return v.title }
func (v ViewItem) Description() string { return v.description }
func (v ViewItem) FilterValue() string { return v.title }
func (v ViewItem) Query() string       { return v.query }
func (v ViewItem) IsTable() bool       { return v.isTable }

func NewSidebar(views []config.View, width, height int) list.Model {
	items := make([]list.Item, len(views))
	for i, v := range views {
		items[i] = ViewItem{
			title:       v.Title,
			description: v.Description,
			query:       v.Query,
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return l
}
