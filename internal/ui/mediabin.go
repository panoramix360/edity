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
	paneStyle := theme.Pane.Default
	if focused {
		list.Styles.Title = theme.Pane.ActiveHeader
		paneStyle = theme.Pane.Active
	} else {
		list.Styles.Title = theme.Pane.Header
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
	d.Styles.NormalTitle = theme.Bin.NormalTitle
	d.Styles.NormalDesc = theme.Bin.NormalDesc
	d.Styles.SelectedTitle = theme.Bin.SelectedTitle
	d.Styles.SelectedDesc = theme.Bin.SelectedDesc
	d.Styles.DimmedTitle = theme.Bin.DimmedTitle
	d.Styles.DimmedDesc = theme.Bin.DimmedTitle

	d.SetSpacing(0)

	list := list.New(clipsToItems(clips), d, 0, 0)
	list.Title = "Media Bin"

	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.DisableQuitKeybindings()

	list.Styles.Title = theme.Pane.Header
	list.Styles.TitleBar = lipgloss.NewStyle().PaddingBottom(1)
	list.Styles.NoItems = theme.Dim.PaddingLeft(2)
	list.Styles.Spinner = theme.Bin.Spinner
	list.Styles.Filter.Focused.Prompt = theme.Bin.FilterPrompt
	list.Styles.Filter.Blurred.Prompt = theme.Bin.FilterPrompt
	return list
}
