package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	SwitchTab key.Binding
	Pause     key.Binding
	Quit      key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		SwitchTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("⇥", "switch tab"),
		),
		Pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("␣", "pause"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k keyMap) footerHints() []key.Binding {
	return []key.Binding{k.SwitchTab, k.Pause, k.Quit}
}
