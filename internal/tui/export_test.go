package tui

func (m Model) CurrentTab() Tab { return m.tab }

func (m Model) Paused() bool { return m.paused }
