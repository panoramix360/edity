package ui

import (
	"fmt"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
)

type ClipItem struct{ c clip.Clip }

func (i ClipItem) FilterValue() string { return filepath.Base(i.c.Path) }
func (i ClipItem) Title() string       { return filepath.Base(i.c.Path) }
func (i ClipItem) Description() string {
	return formatDuration(i.c.Duration()) + "  " + formatRes(i.c.Meta.Width, i.c.Meta.Height) + "  " + fmt.Sprintf("%.2ffps", i.c.Meta.FrameRate)
}

func clipsToItems(clips []clip.Clip) []list.Item {
	items := make([]list.Item, len(clips))
	for i, c := range clips {
		items[i] = ClipItem{c}
	}
	return items
}

type MediaBin struct {
	list list.Model
}

func newMediaBin(clips []clip.Clip) MediaBin {
	return MediaBin{list: newBinList(clips)}
}

func (b MediaBin) Update(msg tea.Msg) (MediaBin, tea.Cmd) {
	var cmd tea.Cmd
	b.list, cmd = b.list.Update(msg)
	return b, cmd
}

func (b MediaBin) Render(width, height int, focused bool) string {
	list := b.list
	paneStyle := defaultPaneStyle
	if focused {
		list.Styles.Title = activeHeaderStyle
		paneStyle = activePaneStyle
	} else {
		list.Styles.Title = headerStyle
	}
	return paneStyle.Width(width).Height(height).Render(list.View())
}

func (b MediaBin) SetSize(w, h int) MediaBin {
	b.list.SetSize(w, h)
	return b
}

func (b MediaBin) SetClips(clips []clip.Clip) (MediaBin, tea.Cmd) {
	cmd := b.list.SetItems(clipsToItems(clips))
	return b, cmd
}

func (b MediaBin) Select(i int) MediaBin {
	b.list.Select(i)
	return b
}

func (b MediaBin) GlobalIndex() int {
	return b.list.GlobalIndex()
}

func (b MediaBin) IsFiltering() bool {
	return b.list.FilterState() == list.Filtering
}

func (b MediaBin) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Quit, Keys.NextPane, Keys.Up, Keys.Down, Keys.Export, Keys.Help}
}

func (b MediaBin) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{Keys.Quit, Keys.ForceQuit, Keys.NextPane, Keys.PrevPane, Keys.Export},
		{Keys.Up, Keys.Down, Keys.Help},
	}
}

func newBinList(clips []clip.Clip) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).Padding(0, 0, 0, 2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("10")).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("10")).
		Padding(0, 0, 0, 1)

	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).Padding(0, 0, 0, 2)
	d.Styles.DimmedDesc = d.Styles.DimmedTitle

	d.SetSpacing(0)

	list := list.New(clipsToItems(clips), d, 0, 0)
	list.Title = "Media Bin"

	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.DisableQuitKeybindings()

	list.Styles.Title = headerStyle
	list.Styles.TitleBar = lipgloss.NewStyle().PaddingBottom(1)
	list.Styles.NoItems = dimStyle.PaddingLeft(2)
	list.Styles.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	list.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	list.Styles.Filter.Blurred.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	return list
}
