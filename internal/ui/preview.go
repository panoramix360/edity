package ui

import (
	"fmt"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
)

type Preview struct {
	width, height int
}

func newPreview() Preview {
	return Preview{}
}

func (p Preview) ShortHelp() []key.Binding {
	return []key.Binding{Keys.Quit, Keys.NextPane, Keys.PrevPane, Keys.Help}
}

func (p Preview) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{Keys.Quit, Keys.ForceQuit, Keys.NextPane, Keys.PrevPane, Keys.Help},
	}
}

func (p Preview) SetSize(w, h int) Preview {
	p.width = w
	p.height = h
	return p
}

func (p Preview) Render(clips []clip.Clip, selected int, focused bool) string {
	paneStyle := theme.Pane.Default
	headerS := theme.Pane.Header
	if focused {
		paneStyle = theme.Pane.Active
		headerS = theme.Pane.ActiveHeader
	}

	header := headerS.Render("Preview")

	var lines []string
	lines = append(lines, header)

	if len(clips) > 0 && selected >= 0 && selected < len(clips) {
		c := clips[selected]
		lines = append(lines,
			"",
			"  "+filepath.Base(c.Path),
			"",
			theme.Dim.Render("  Duration   "+formatDuration(c.Duration())),
			theme.Dim.Render("  Video      "+formatRes(c.Meta.Width, c.Meta.Height)+"  "+fmt.Sprintf("%.2ffps", c.Meta.FrameRate)),
			theme.Dim.Render("  Codec      "+c.Meta.Codec),
			"",
			theme.Dim.Render("  Playback coming in v0.4"),
		)
	}

	return paneStyle.Width(p.width).Height(p.height).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
