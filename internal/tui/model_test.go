package tui_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mickamy/dbtop/internal/driver"
	"github.com/mickamy/dbtop/internal/tui"
)

type fakeDriver struct {
	backends   []driver.Backend
	sample     driver.MetricSample
	statements []driver.Statement
}

func (f fakeDriver) Activity(context.Context) ([]driver.Backend, error) {
	return f.backends, nil
}

func (f fakeDriver) Metrics(context.Context) (driver.MetricSample, error) {
	return f.sample, nil
}

func (f fakeDriver) Statements(context.Context) ([]driver.Statement, error) {
	return f.statements, nil
}

func (f fakeDriver) ResetStatements(context.Context) error  { return nil }
func (f fakeDriver) Cancel(context.Context, int64) error    { return nil }
func (f fakeDriver) Terminate(context.Context, int64) error { return nil }
func (f fakeDriver) Close() error                           { return nil }

func (f fakeDriver) Capabilities(context.Context) (driver.Capabilities, error) {
	return driver.Capabilities{}, nil
}

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestActivityPollRenders(t *testing.T) {
	t.Parallel()

	d := fakeDriver{backends: []driver.Backend{
		{PID: 42, User: "app", DB: "myapp", State: driver.StateActive},
	}}

	m := tui.New(d, driver.Capabilities{}, time.Second)
	view := mustUpdate(t, m, m.Init()()).View()

	if !strings.Contains(view, "42") || !strings.Contains(view, "app") {
		t.Errorf("view missing backend data:\n%s", view)
	}
}

func TestStatementsGuidanceWithoutExtension(t *testing.T) {
	t.Parallel()

	// caps.Statements is false: the tab must guide, not poll and error.
	m := tui.New(fakeDriver{}, driver.Capabilities{}, time.Second)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})    // Metrics
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // Statements

	view := mustUpdate(t, model, asModel(t, model).Init()()).View()

	if !strings.Contains(view, "CREATE EXTENSION pg_stat_statements") {
		t.Errorf("statements tab did not show guidance:\n%s", view)
	}
}

func TestBodyFitsHeight(t *testing.T) {
	t.Parallel()

	backends := make([]driver.Backend, 50)
	for i := range backends {
		backends[i] = driver.Backend{PID: int64(i + 1), User: "app", DB: "db", State: driver.StateActive}
	}

	const (
		width  = 100
		height = 12
	)

	var model tea.Model = tui.New(fakeDriver{backends: backends}, driver.Capabilities{}, time.Second)
	model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	model = mustUpdate(t, model, asModel(t, model).Init()())

	view := model.View()

	if lines := strings.Count(view, "\n") + 1; lines > height {
		t.Errorf("view has %d lines, exceeds height %d:\n%s", lines, height, view)
	}

	if !strings.Contains(view, "more") {
		t.Errorf("expected a truncation indicator:\n%s", view)
	}

	if !strings.Contains(view, "Activity") {
		t.Errorf("tab bar dropped off the top:\n%s", view)
	}

	// The tab bar must be the very first line, never scrolled off.
	if first := strings.SplitN(view, "\n", 2)[0]; !strings.Contains(first, "Activity") {
		t.Errorf("first line is not the tab bar: %q", first)
	}
}

// A wide row must be clipped to the width, otherwise it wraps and the wrapped
// lines push the tab bar off the top.
func TestRowsClippedToWidth(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("SELECT a, b, c FROM very_long_table_name x JOIN y ", 5)
	statements := make([]driver.Statement, 20)
	for i := range statements {
		statements[i] = driver.Statement{Calls: 100, Query: long}
	}

	const (
		width  = 60
		height = 14
	)

	var model tea.Model = tui.New(fakeDriver{statements: statements}, driver.Capabilities{Statements: true}, time.Second)
	model, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // Metrics
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // Statements
	model = mustUpdate(t, model, asModel(t, model).Init()())

	for line := range strings.SplitSeq(model.View(), "\n") {
		if w := lipgloss.Width(line); w > width {
			t.Errorf("line width %d exceeds %d: %q", w, width, line)
		}
	}
}

func TestQuitKey(t *testing.T) {
	t.Parallel()

	m := tui.New(fakeDriver{}, driver.Capabilities{}, time.Second)

	_, cmd := m.Update(keyRunes("q"))
	if cmd == nil {
		t.Fatal("expected a command for q")
	}

	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("q produced %T, want tea.QuitMsg", cmd())
	}
}

func TestTabKeyCycles(t *testing.T) {
	t.Parallel()

	var model tea.Model = tui.New(fakeDriver{}, driver.Capabilities{}, time.Second)

	for _, want := range []tui.Tab{tui.TabMetrics, tui.TabStatements, tui.TabActivity} {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		if got := asModel(t, model).CurrentTab(); got != want {
			t.Errorf("tab = %v, want %v", got, want)
		}
	}
}

func TestSpaceTogglesPause(t *testing.T) {
	t.Parallel()

	var model tea.Model = tui.New(fakeDriver{}, driver.Capabilities{}, time.Second)

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !asModel(t, model).Paused() {
		t.Error("space did not pause")
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	if asModel(t, model).Paused() {
		t.Error("space did not unpause")
	}
}

func mustUpdate(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()

	updated, _ := m.Update(msg)

	return updated
}

func asModel(t *testing.T, m tea.Model) tui.Model {
	t.Helper()

	got, ok := m.(tui.Model)
	if !ok {
		t.Fatalf("model is %T, want tui.Model", m)
	}

	return got
}
