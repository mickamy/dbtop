package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) CurrentTab() Tab { return m.tab }

func (m Model) Paused() bool { return m.paused }

func (m Model) AwaitingResult() bool { return m.awaitingResult }

func TickMsg() tea.Msg { return tickMsg{} }
