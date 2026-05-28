package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	activeTabStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	dimStyle       = lipgloss.NewStyle().Faint(true)
	errStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// fallbackHeight is used before the first WindowSizeMsg arrives.
const fallbackHeight = 24

// View assembles the frame as exact lines — tab bar, body, footer — so the
// total never exceeds the terminal height and the tab bar and footer stay put.
func (m Model) View() string {
	header := []string{m.tabBar(), ""}
	if m.err != nil {
		header = append(header, errStyle.Render("error: "+m.err.Error()), "")
	}

	footer := []string{m.footer()}

	height := m.height
	if height <= 0 {
		height = fallbackHeight
	}

	bodyMax := max(height-len(header)-len(footer), 1)

	lines := make([]string, 0, len(header)+bodyMax+len(footer))
	lines = append(lines, header...)
	lines = append(lines, m.bodyLines(bodyMax)...)
	lines = append(lines, footer...)

	return strings.Join(lines, "\n")
}

func (m Model) tabBar() string {
	tabs := make([]string, 0, 3)

	for _, t := range []Tab{TabActivity, TabMetrics, TabStatements} {
		if t == m.tab {
			tabs = append(tabs, activeTabStyle.Render(t.String()))
		} else {
			tabs = append(tabs, dimStyle.Render(t.String()))
		}
	}

	title := "dbtop"
	if m.paused {
		title += " " + dimStyle.Render("[paused]")
	}

	return title + "  " + strings.Join(tabs, "  ")
}

func (m Model) bodyLines(maxLines int) []string {
	switch m.tab {
	case TabActivity:
		return m.activityLines(maxLines)
	case TabMetrics:
		return m.metricsLines()
	case TabStatements:
		return m.statementsLines(maxLines)
	}

	return nil
}

func (m Model) activityLines(maxLines int) []string {
	if len(m.backends) == 0 {
		return []string{dimStyle.Render("no backends")}
	}

	lines := []string{fmt.Sprintf("%d backends", len(m.backends))}
	show, more := capRows(len(m.backends), maxLines-1)

	for _, be := range m.backends[:show] {
		dur := "—"
		if d, ok := be.Duration(); ok {
			dur = d.Truncate(time.Millisecond).String()
		}

		wait := be.Wait()
		if wait == "" {
			wait = "—"
		}

		row := fmt.Sprintf("  %-7d %s@%s %-14s %-10s %s",
			be.PID, be.User, be.DB, be.State, dur, wait)
		lines = append(lines, m.clip(row))
	}

	return appendMore(lines, more)
}

func (m Model) metricsLines() []string {
	c := m.metrics.Conns

	return []string{
		m.clip(fmt.Sprintf("connections %d/%d  active %d  idle %d  idle-tx %d",
			c.Total, m.metrics.MaxConnections, c.Active, c.Idle, c.IdleInTx)),
		m.clip(fmt.Sprintf("waiting locks %d  replicas %d",
			m.metrics.WaitingLocks, len(m.metrics.Replicas))),
	}
}

func (m Model) statementsLines(maxLines int) []string {
	if !m.caps.Statements {
		return []string{
			dimStyle.Render("pg_stat_statements is not installed."),
			dimStyle.Render("run:  CREATE EXTENSION pg_stat_statements;"),
		}
	}

	if len(m.statements) == 0 {
		return []string{dimStyle.Render("no statements yet")}
	}

	lines := []string{fmt.Sprintf("%d statements", len(m.statements))}
	show, more := capRows(len(m.statements), maxLines-1)

	for _, s := range m.statements[:show] {
		row := fmt.Sprintf("  calls %-8d total %-12s mean %-10s %s",
			s.Calls,
			s.Total.Truncate(time.Microsecond),
			s.Mean.Truncate(time.Microsecond),
			truncate(s.Query, 60))
		lines = append(lines, m.clip(row))
	}

	return appendMore(lines, more)
}

func (m Model) footer() string {
	hints := m.keys.footerHints()
	parts := make([]string, 0, len(hints))

	for _, b := range hints {
		h := b.Help()
		parts = append(parts, h.Key+" "+h.Desc)
	}

	return dimStyle.Render(strings.Join(parts, " · "))
}

func appendMore(lines []string, more int) []string {
	if more > 0 {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  … %d more", more)))
	}

	return lines
}

// capRows splits total items into how many to show within budget and how many
// remain, reserving one line for the "… N more" note when truncating.
func capRows(total, budget int) (show, more int) {
	if budget <= 0 {
		return 0, total
	}

	if total <= budget {
		return total, 0
	}

	return budget - 1, total - (budget - 1)
}

// clip shortens a plain (unstyled) line to the terminal width so it never wraps.
func (m Model) clip(s string) string {
	if m.width <= 0 {
		return s
	}

	r := []rune(s)
	if len(r) <= m.width {
		return s
	}

	if m.width == 1 {
		return "…"
	}

	return string(r[:m.width-1]) + "…"
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}

	return s[:n-1] + "…"
}
