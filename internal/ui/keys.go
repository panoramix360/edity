package ui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	ForceQuit key.Binding
	Quit      key.Binding
	NextPane  key.Binding
	PrevPane  key.Binding
	Export    key.Binding
	Help      key.Binding

	Up   key.Binding
	Down key.Binding

	PlayheadLeft  key.Binding
	PlayheadRight key.Binding
	FastLeft      key.Binding
	FastRight     key.Binding
	PrevBoundary  key.Binding
	NextBoundary  key.Binding
	PrevFrame     key.Binding
	NextFrame     key.Binding

	Split  key.Binding
	Delete key.Binding
}

var Keys = KeyMap{
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	NextPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	PrevPane: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev pane"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PlayheadLeft: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "playhead ←"),
	),
	PlayheadRight: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "playhead →"),
	),
	FastLeft: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+←", "fast ←"),
	),
	FastRight: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+→", "fast →"),
	),
	PrevBoundary: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "prev boundary"),
	),
	NextBoundary: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next boundary"),
	),
	PrevFrame: key.NewBinding(
		key.WithKeys(","),
		key.WithHelp(",", "prev frame"),
	),
	NextFrame: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "next frame"),
	),
	Split: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "split"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "backspace"),
		key.WithHelp("d", "delete"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit, k.NextPane, k.Up, k.Down,
		k.PlayheadLeft, k.PlayheadRight,
		k.Split, k.Delete, k.Export, k.Help,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.ForceQuit, k.NextPane, k.PrevPane, k.Export},
		{k.Up, k.Down},
		{k.PlayheadLeft, k.PlayheadRight, k.FastLeft, k.FastRight},
		{k.PrevBoundary, k.NextBoundary, k.PrevFrame, k.NextFrame},
		{k.Split, k.Delete},
	}
}
