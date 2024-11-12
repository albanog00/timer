package timer

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
	"golang.org/x/term"
)

const (
	interval          = time.Millisecond * 100
	paddingVertical   = 1
	paddingHorizontal = 4
)

var (
	ui = lipgloss.NewStyle().
		Padding(paddingVertical, paddingHorizontal).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: "#000000",
			Dark:  "#FFFFFF",
		}).
		Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA7B7E"))
	logsStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#11F696"))
)

type Model struct {
	timer     timer.Model
	textinput textinput.Model
	progress  progress.Model
	keymaps   keymaps.Model

	start time.Time
	end   time.Time

	initial time.Duration
	total   time.Duration
	passed  time.Duration
	stopped time.Time

	width  int
	height int

	adding   bool
	quitting bool

	logs string
	err  error
}

func New(timeout time.Duration) *Model {
	return &Model{
		timer:     timer.NewWithInterval(timeout, interval),
		textinput: textinput.New(),
		progress:  progress.New(progress.WithDefaultGradient()),
		keymaps:   keymaps.New(),
		start:     time.Now(),
		end:       time.Now().Add(timeout),
		initial:   timeout,
		total:     timeout,
	}
}

func (m Model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		if !m.timer.Running() {
			return m, nil
		}
		var cmds []tea.Cmd
		var cmd tea.Cmd

		m.passed += m.timer.Interval
		cmds = append(cmds,
			m.progress.SetPercent(
				float64(m.passed.Microseconds())/float64(m.total.Microseconds()),
			))

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

	case tea.WindowSizeMsg:
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			m.width = width
			m.height = height
		}
		return m, nil

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
	}

	return m, nil
}

func (m *Model) input(msg tea.KeyMsg) tea.Cmd {
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
			} else {
				m.timer.Timeout += addTime
				m.total += addTime
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
	buffer.WriteString(fmt.Sprintf("%s - %s\n",
		m.start.Format("Start: 15:04:05"),
		m.end.Format("End: 15:04:05"),
	))
	buffer.WriteString(fmt.Sprintf("\n%s\n%s\n", m.timer.View(), m.progress.View()))

	if !m.quitting {
		buffer.WriteString(m.logs)
		if m.adding {
			buffer.WriteString(fmt.Sprintf(
				"Insert time to add.\n%s\n",
				m.textinput.View(),
			))
		}
		if m.err != nil {
			buffer.WriteString(fmt.Sprintf("%s\n", errorStyle.Render(m.err.Error())))
		}
		buffer.WriteString(fmt.Sprintf("%s", m.keymaps.View()))
	}
	if m.timer.Timedout() {
		buffer.WriteString("Time is up!")
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, ui.Render(buffer.String()))
}
