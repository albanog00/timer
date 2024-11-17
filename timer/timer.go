package timer

import (
	"albanog/timer/keymaps"
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	interval          = time.Millisecond * 100
	paddingVertical   = 1
	paddingHorizontal = 4
)

var (
	adaptiveColor = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}

	ui = lipgloss.
		NewStyle().
		Padding(paddingVertical, paddingHorizontal).
		Border(lipgloss.NormalBorder()).
		BorderForeground(adaptiveColor).
		Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA7B7E"))
	logsStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#11F696"))
)

type Model struct {
	timer     *timer.Model
	textinput *textinput.Model
	progress  *progress.Model
	keymaps   *keymaps.Model

	start   time.Time
	end     time.Time
	stopped time.Time

	total  float64
	passed float64

	width  int
	height int

	adding bool

	logs string
	err  error
}

func New(timeout time.Duration) *Model {
	timer := timer.NewWithInterval(timeout, interval)
	textinput := textinput.New()
	progress := progress.New(progress.WithDefaultGradient())
	keymaps := keymaps.New()

	return &Model{
		timer:     &timer,
		textinput: &textinput,
		progress:  &progress,
		keymaps:   &keymaps,
		start:     time.Now(),
		end:       time.Now().Add(timeout),
		total:     timeout.Seconds(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		if !m.timer.Running() {
			break
		}
		m.passed += m.timer.Interval.Seconds()
		*m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		*m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if m.adding {
			var cmd tea.Cmd
			cmd = m.input(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keymaps.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymaps.Clear):
			m.logs = ""
			m.err = nil
			return m, nil

		case key.Matches(msg, m.keymaps.Start, m.keymaps.Stop):
			if m.keymaps.Stop.Enabled() {
				m.stopped = time.Now()
			} else {
				m.end = m.end.Add(time.Now().Sub(m.stopped))
			}

			m.keymaps.Start.SetEnabled(!m.keymaps.Start.Enabled())
			m.keymaps.Stop.SetEnabled(!m.keymaps.Stop.Enabled())

			return m, m.timer.Toggle()

		case key.Matches(msg, m.keymaps.Add):
			m.adding = true
			m.keymapsToggleOnAdd()
			m.textinput.Focus()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *Model) input(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	*m.textinput, cmd = m.textinput.Update(msg)

	if key.Matches(msg, m.keymaps.Send, m.keymaps.StopAdd) {
		m.err = nil
		m.adding = false

		if key.Matches(msg, m.keymaps.Send) {
			addTime, err := time.ParseDuration(m.textinput.Value())
			if err != nil {
				m.err = errors.New("Invalid input, try again.")
				cmd = nil
			} else {
				m.timer.Timeout += addTime
				m.total += addTime.Seconds()
				m.end = m.end.Add(addTime)
				if addTime < 0 {
					m.logs = fmt.Sprintf("> %s\n",
						logsStyle.Render("Removed", addTime.Abs().String()))
				} else {
					m.logs = fmt.Sprintf("> %s\n",
						logsStyle.Render("Added", addTime.Abs().String()))
				}
			}
		}
		m.keymapsToggleOnAdd()
		m.textinput.Reset()
	}

	return cmd
}

func (m *Model) keymapsToggleOnAdd() {
	m.keymaps.Stop.SetEnabled(!m.adding && m.timer.Running())
	m.keymaps.Start.SetEnabled(!m.adding && !m.timer.Running())

	m.keymaps.Clear.SetEnabled(!m.adding)
	m.keymaps.Quit.SetEnabled(!m.adding)
	m.keymaps.Add.SetEnabled(!m.adding)

	m.keymaps.StopAdd.SetEnabled(m.adding)
	m.keymaps.Send.SetEnabled(m.adding)
}

func (m *Model) View() string {
	buffer := &bytes.Buffer{}
	buffer.WriteString(m.start.Format("Start: 15:04:05"))
	buffer.WriteString(" - ")
	buffer.WriteString(m.end.Format("End: 15:04:05"))
	buffer.WriteByte('\n')
	buffer.WriteByte('\n')
	buffer.WriteString(m.timer.View())
	buffer.WriteByte('\n')
	buffer.WriteString(m.progress.ViewAs(m.passed / m.total))
	buffer.WriteByte('\n')
	buffer.WriteString(m.logs)
	if m.adding {
		buffer.WriteString("Insert time to add.\n")
		buffer.WriteString(m.textinput.View())
		buffer.WriteByte('\n')
	}
	if m.err != nil {
		buffer.WriteString(errorStyle.Render(m.err.Error()))
		buffer.WriteByte('\n')
	}
	buffer.WriteString(m.keymaps.View())

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			time.Now().Format("15:04:05"),
			ui.Render(buffer.String()),
		))
}
