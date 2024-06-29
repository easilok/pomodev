package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

// https://pomodoro-tracker.com/

const restRatio float64 = 5.0 / 20.0

type APP_MODE int

const (
	APP_MODE_INIT APP_MODE = iota
	APP_MODE_WORK
	APP_MODE_REST
)

type model struct {
	mode      APP_MODE
	stopwatch stopwatch.Model
	timer     timer.Model
	keymap    keymap
	help      help.Model
	quitting  bool
}

type keymap struct {
	start      key.Binding
	finishWork key.Binding
	finishRest key.Binding
	reset      key.Binding
	quit       key.Binding
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	s := ""
	if !m.quitting {
		switch m.mode {
		case APP_MODE_INIT:
			s = "You can start your work session any time" + "\n"
		case APP_MODE_WORK:
			s = m.stopwatch.View() + "\n"
			s = "Work session: " + s
		case APP_MODE_REST:
			s = m.timer.View() + "\n"
			s = "Rest session: " + s
		}
		s += m.helpView()
	}
	return s
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.finishWork,
		m.keymap.finishRest,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TimeoutMsg:
		m.mode = APP_MODE_WORK
		m.timer.Stop()
		m.keymap.finishWork.SetEnabled(true)
		m.keymap.finishRest.SetEnabled(false)
		m.stopwatch = stopwatch.NewWithInterval(time.Second)
		return m, m.stopwatch.Init()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keymap.reset):
			if m.mode == APP_MODE_WORK {
				m.mode = APP_MODE_INIT
				return m, m.stopwatch.Reset()
			}
			m.mode = APP_MODE_INIT
			return m, m.timer.Stop()

		case key.Matches(msg, m.keymap.finishWork):
			m.mode = APP_MODE_REST
			workTime := m.stopwatch.Elapsed()
			restTime := workTime.Seconds() * restRatio
			m.stopwatch.Stop()
			m.stopwatch.Reset()
			m.timer = timer.NewWithInterval(time.Second*time.Duration(restTime), time.Second)
			m.keymap.finishWork.SetEnabled(false)
			m.keymap.finishRest.SetEnabled(true)
			return m, m.timer.Init()

		case key.Matches(msg, m.keymap.finishRest):
			m.mode = APP_MODE_WORK
			m.timer.Stop()
			m.keymap.finishWork.SetEnabled(true)
			m.keymap.finishRest.SetEnabled(false)
			m.stopwatch = stopwatch.NewWithInterval(time.Second)
			return m, m.stopwatch.Init()

		case key.Matches(msg, m.keymap.start):
			m.keymap.start.SetEnabled(false)
			m.keymap.finishWork.SetEnabled(true)
			m.mode = APP_MODE_WORK
			m.stopwatch = stopwatch.NewWithInterval(time.Second)
			return m, m.stopwatch.Init()
		}
	}

	var cmd tea.Cmd = nil
	switch m.mode {
	case APP_MODE_WORK:
		m.stopwatch, cmd = m.stopwatch.Update(msg)
	case APP_MODE_REST:
		m.timer, cmd = m.timer.Update(msg)
	}

	return m, cmd
}

func main() {
	m := model{
		mode:      APP_MODE_INIT,
		stopwatch: stopwatch.NewWithInterval(time.Second),
		timer:     timer.NewWithInterval(1, time.Second),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			finishWork: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "finish work"),
			),
			finishRest: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "finish rest"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.New(),
	}

	m.keymap.finishWork.SetEnabled(false)
	m.keymap.finishRest.SetEnabled(false)
	m.keymap.reset.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Oh no, it didn't work:", err)
		os.Exit(1)
	}
}
