package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mickamy/dbtop/internal/driver"
)

// pollTimeout bounds a single poll, so a stuck query can't wedge the loop.
const pollTimeout = 5 * time.Second

type (
	tickMsg       struct{}
	activityMsg   []driver.Backend
	metricsMsg    driver.MetricSample
	statementsMsg []driver.Statement
	errMsg        struct{ err error }
)

type Model struct {
	driver   driver.Driver
	caps     driver.Capabilities
	interval time.Duration
	keys     keyMap

	tab    Tab
	paused bool

	// awaitingResult is true while the poll/tick chain is alive; it stops a
	// second chain from starting when unpausing before a pending tick fires.
	awaitingResult bool

	backends   []driver.Backend
	metrics    driver.MetricSample
	statements []driver.Statement
	err        error

	width  int
	height int
}

func New(d driver.Driver, caps driver.Capabilities, interval time.Duration) Model {
	return Model{driver: d, caps: caps, interval: interval, keys: defaultKeys(), tab: TabActivity, awaitingResult: true}
}

// Init seeds the poll loop with one immediate fetch; each result schedules the
// next tick, so only a single tick chain is ever in flight.
func (m Model) Init() tea.Cmd {
	return m.poll()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		if m.paused {
			m.awaitingResult = false

			return m, nil
		}

		return m, m.poll()

	case activityMsg:
		m.backends = msg
		m.err = nil

		return m, m.scheduleTick()

	case metricsMsg:
		m.metrics = driver.MetricSample(msg)
		m.err = nil

		return m, m.scheduleTick()

	case statementsMsg:
		m.statements = msg
		m.err = nil

		return m, m.scheduleTick()

	case errMsg:
		m.err = msg.err

		return m, m.scheduleTick()
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.SwitchTab):
		m.tab = m.tab.next()

		return m, nil

	case key.Matches(msg, m.keys.Pause):
		m.paused = !m.paused
		if m.paused {
			return m, nil
		}

		// A live chain resumes on its own; only reseed once it has stopped.
		if m.awaitingResult {
			return m, nil
		}

		m.awaitingResult = true

		return m, m.poll()
	}

	return m, nil
}

func (m Model) scheduleTick() tea.Cmd {
	return tea.Tick(m.interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) poll() tea.Cmd {
	tab := m.tab

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
		defer cancel()

		switch tab {
		case TabActivity:
			backends, err := m.driver.Activity(ctx)
			if err != nil {
				return errMsg{err}
			}

			return activityMsg(backends)

		case TabMetrics:
			sample, err := m.driver.Metrics(ctx)
			if err != nil {
				return errMsg{err}
			}

			return metricsMsg(sample)

		case TabStatements:
			if !m.caps.Statements {
				return statementsMsg(nil)
			}

			statements, err := m.driver.Statements(ctx)
			if err != nil {
				return errMsg{err}
			}

			return statementsMsg(statements)
		}

		return nil
	}
}
