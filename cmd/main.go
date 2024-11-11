package main

import (
	"albanog/timer/keymaps"
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <duration>.\n", os.Args[0])
		os.Exit(1)
	}

	if os.Args[1] == "help" {
		fmt.Printf("usage: %s <duration>.\n", os.Args[0])
		os.Exit(0)
	}
	timeout, err := time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Printf("Invalid duration provided.\nExample: 10m (10 minutes)")
		os.Exit(1)
	}

	m := New(timeout)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("There's been an error: %v\n", err)
		os.Exit(1)
	}
}

const (
	interval = time.Millisecond * 100
	padding  = 2
	maxWidth = 80
)

var (
	boldStyle = lipgloss.NewStyle().Bold(true)
)

type model struct {
	timer     timer.Model
	textinput textinput.Model
	progress  progress.Model
	keymaps   keymaps.Model

	start time.Time
	end   time.Time

	initial time.Duration
	total   time.Duration
	passed  time.Duration

	adding   bool
	quitting bool

	logs []string
	err  error
}

func New(timeout time.Duration) model {
	return model{
		timer:     timer.NewWithInterval(timeout, interval),
		textinput: textinput.New(),
		progress:  progress.New(progress.WithDefaultGradient()),
		keymaps:   keymaps.New(),
		start:     time.Now(),
		end:       time.Now().Add(timeout),
		initial:   timeout,
		total:     timeout,
		passed:    0,
		adding:    false,
		quitting:  false,
		logs:      make([]string, 0),
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		if !m.timer.Running() {
			return m, nil
		}
		var cmds []tea.Cmd
		var cmd tea.Cmd

		m.passed += m.timer.Interval
		pct := m.passed.Milliseconds() * 100 / m.total.Milliseconds()
		cmds = append(cmds, m.progress.SetPercent(float64(pct)/100))

		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, tea.Batch(cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
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
		case key.Matches(msg, m.keymaps.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymaps.Reset):
			m.timer.Timeout = m.initial
			m.total = m.initial
			m.passed = time.Duration(0)
			return m, m.progress.SetPercent(0)
		case key.Matches(msg, m.keymaps.Start, m.keymaps.Stop):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymaps.Add):
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

	if key.Matches(msg, m.keymaps.Send, m.keymaps.StopAdd) {
		m.err = nil
		m.adding = false

		if key.Matches(msg, m.keymaps.Send) {
			addTime, err := time.ParseDuration(m.textinput.Value())
			if err != nil {
				m.err = errors.New("Invalid input, try again.")
				cmd = nil
			} else if addTime.Abs() > time.Duration(0) {
				m.timer.Timeout += addTime
				m.total += addTime
				m.end = m.end.Add(addTime)
				if addTime < time.Duration(0) {
					m.logs = append(m.logs, fmt.Sprintf(" > Removed %s", addTime.Abs().String()))
				} else {
					m.logs = append(m.logs, fmt.Sprintf(" > Added %s", addTime.String()))
				}
			}
		}
		m.keymaps = m.keymapsToggleOnAdd()
		m.textinput.Reset()
	}

	return m, cmd
}

func (m model) keymapsToggleOnAdd() keymaps.Model {
	m.keymaps.Stop.SetEnabled(!m.adding && m.timer.Running())
	m.keymaps.Start.SetEnabled(!m.adding && !m.timer.Running())

	m.keymaps.Reset.SetEnabled(!m.adding)
	m.keymaps.Quit.SetEnabled(!m.adding)
	m.keymaps.Add.SetEnabled(!m.adding)

	m.keymaps.StopAdd.SetEnabled(m.adding)
	m.keymaps.Send.SetEnabled(m.adding)

	return m.keymaps
}

func (m model) View() string {
	builder := &bytes.Buffer{}
	builder.WriteString(fmt.Sprintf(" %s - %s\n",
		m.start.Format("Start: 15:04:05"),
		m.end.Format("End: 15:04:05"),
	))
	builder.WriteString(fmt.Sprintf("\n - %s\n %s", m.timer.View(), m.progress.View()))

	if !m.quitting {
		for _, line := range m.logs {
			builder.WriteString(fmt.Sprintf("\n%s", line))
		}
		if m.adding {
			builder.WriteString(fmt.Sprintf(
				"\n Please insert time to add.\n\n %s\n\n %s\n",
				m.textinput.View(),
				"(esc to go back)",
			))
		}
		if m.err != nil {
			builder.WriteString(fmt.Sprintf("\n%s\n", m.err.Error()))
		}
		builder.WriteString(fmt.Sprintf("\n%s", m.keymaps.View()))
	}
	if m.timer.Timedout() {
		builder.WriteString(" Time is up!")
	}

	builder.WriteByte('\n')
	return boldStyle.Render(builder.String())
}
