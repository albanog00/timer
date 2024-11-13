package keymaps

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	help *help.Model

	Start   key.Binding
	Stop    key.Binding
	Add     key.Binding
	StopAdd key.Binding
	Send    key.Binding
	Clear   key.Binding
	Quit    key.Binding
}

func New() Model {
	help := help.New()

	m := Model{
		help:    &help,
		Stop:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "stop")),
		Start:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
		Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		StopAdd: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "go back")),
		Send:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
		Clear:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
	m.Start.SetEnabled(false)
	m.StopAdd.SetEnabled(false)
	m.Send.SetEnabled(false)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Cmd) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return " " + m.help.ShortHelpView([]key.Binding{
		m.Start,
		m.Stop,
		m.Add,
		m.StopAdd,
		m.Send,
		m.Clear,
		m.Quit,
	})
}
