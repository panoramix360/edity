package ui

import "charm.land/lipgloss/v2"

var (
	colorMuted  = lipgloss.Color("8")  // gray  — borders, dim text
	colorAccent = lipgloss.Color("10") // green — active, selected, success
	colorError  = lipgloss.Color("9")  // red   — errors, playhead
	colorInfo   = lipgloss.Color("12") // blue  — modal, key chips
	colorDark   = lipgloss.Color("0")  // black — text on colored backgrounds
	colorLight  = lipgloss.Color("15") // white — text on dark backgrounds
)

type PaneStyles struct {
	Default      lipgloss.Style
	Active       lipgloss.Style
	Header       lipgloss.Style
	ActiveHeader lipgloss.Style
}

type ModalStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Hint      lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	KeyChip   lipgloss.Style
}

type BinStyles struct {
	NormalTitle   lipgloss.Style
	NormalDesc    lipgloss.Style
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style
	DimmedTitle   lipgloss.Style
	Spinner       lipgloss.Style
	FilterPrompt  lipgloss.Style
}

type TimelineStyles struct {
	Bar            lipgloss.Style
	SelectedBar    lipgloss.Style
	PlayheadCursor lipgloss.Style
	PlayheadArrow  lipgloss.Style
}

type Theme struct {
	Pane     PaneStyles
	Modal    ModalStyles
	Bin      BinStyles
	Timeline TimelineStyles
	Dim      lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Pane: PaneStyles{
			Default:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorMuted),
			Active:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent),
			Header:       lipgloss.NewStyle().Foreground(colorMuted).Bold(true),
			ActiveHeader: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		},
		Modal: ModalStyles{
			Container: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorInfo).Padding(1, 3),
			Title:     lipgloss.NewStyle().Foreground(colorInfo).Bold(true),
			Hint:      lipgloss.NewStyle().Foreground(colorMuted),
			Success:   lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
			Error:     lipgloss.NewStyle().Foreground(colorError).Bold(true),
			KeyChip:   lipgloss.NewStyle().Foreground(colorDark).Background(colorInfo).Padding(0, 1),
		},
		Bin: BinStyles{
			NormalTitle: lipgloss.NewStyle().Padding(0, 0, 0, 2),
			NormalDesc:  lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 0, 0, 2),
			SelectedTitle: lipgloss.NewStyle().
				Foreground(colorAccent).Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorAccent).
				Padding(0, 0, 0, 1),
			SelectedDesc: lipgloss.NewStyle().
				Foreground(colorMuted).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorAccent).
				Padding(0, 0, 0, 1),
			DimmedTitle:  lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 0, 0, 2),
			Spinner:      lipgloss.NewStyle().Foreground(colorInfo),
			FilterPrompt: lipgloss.NewStyle().Foreground(colorInfo),
		},
		Timeline: TimelineStyles{
			Bar:            lipgloss.NewStyle().Background(colorMuted).Foreground(colorLight),
			SelectedBar:    lipgloss.NewStyle().Background(colorAccent).Foreground(colorDark),
			PlayheadCursor: lipgloss.NewStyle().Background(colorError).Foreground(colorLight).Bold(true),
			PlayheadArrow:  lipgloss.NewStyle().Foreground(colorError),
		},
		Dim: lipgloss.NewStyle().Foreground(colorMuted),
	}
}

var theme = DefaultTheme()
