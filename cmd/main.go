package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	tick    = time.Millisecond * 100
	timeout = time.Second * 20
)

type model struct {
	timer     timer.Model
	help      help.Model
	textinput textinput.Model
	err       error
	keymaps   keymaps
	adding    bool
	quitting  bool
}

type keymaps struct {
	stop    key.Binding
	start   key.Binding
	add     key.Binding
	stopAdd key.Binding
	send    key.Binding
	reset   key.Binding
	quit    key.Binding
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymaps.stop.SetEnabled(m.timer.Running())
		m.keymaps.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		if m.adding {
			var cmd tea.Cmd
			m, cmd = m.input(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keymaps.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymaps.reset):
			m.timer.Timeout = timeout
		case key.Matches(msg, m.keymaps.start, m.keymaps.stop):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymaps.add):
			m.adding = true

			m.keymaps = m.keymapsToggleOnAdd()
			m.textinput.Focus()
		}
	}

	return m, nil
}

func (m model) input(msg tea.KeyMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)

	if key.Matches(msg, m.keymaps.send, m.keymaps.stopAdd) {
		if key.Matches(msg, m.keymaps.send) {
			addTime, err := time.ParseDuration(m.textinput.Value())
			if err != nil {
				m.err = errors.New("Invalid input, please try again.")
				return m, nil
			}
			m.timer.Timeout += addTime
		}
		m.adding = false
		m.err = nil

		m.keymaps = m.keymapsToggleOnAdd()
		m.textinput.Reset()
	}

	return m, cmd
}

func (m model) keymapsToggleOnAdd() keymaps {
	m.keymaps.stop.SetEnabled(!m.adding && m.timer.Running())
	m.keymaps.start.SetEnabled(!m.adding && !m.timer.Running())

	m.keymaps.reset.SetEnabled(!m.adding)
	m.keymaps.quit.SetEnabled(!m.adding)
	m.keymaps.add.SetEnabled(!m.adding)

	m.keymaps.stopAdd.SetEnabled(m.adding)
	m.keymaps.send.SetEnabled(m.adding)

	return m.keymaps
}

func (m model) helpView() string {
	return m.help.ShortHelpView([]key.Binding{
		m.keymaps.start,
		m.keymaps.stop,
		m.keymaps.add,
		m.keymaps.stopAdd,
		m.keymaps.send,
		m.keymaps.reset,
		m.keymaps.quit,
	})
}

func (m model) View() string {
	s := m.timer.View()

	if m.timer.Timedout() {
		s = "All done!"
	}
	s += "\n"

	if !m.quitting {
		if m.adding {
			s += fmt.Sprintf(
				"\nPlease insert time to add.\n\n%s\n\n%s",
				m.textinput.View(),
				"(esc to go back)",
			) + "\n"
		}
		s = fmt.Sprintf(" - %s", s)

		if m.err != nil && !m.adding {
			s += fmt.Sprintf("\n%s\n", m.err.Error())
		}
		s += fmt.Sprintf("\n%s", m.helpView())
	}
	return s
}

func main() {
	m := model{
		timer:     timer.NewWithInterval(timeout, tick),
		help:      help.New(),
		textinput: textinput.New(),
		keymaps: keymaps{
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			add: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add"),
			),
			stopAdd: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "go back"),
			),
			send: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "send"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
	}
	m.keymaps.start.SetEnabled(false)
	m.keymaps.stopAdd.SetEnabled(false)
	m.keymaps.send.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
