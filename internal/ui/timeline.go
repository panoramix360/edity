package ui

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
)

type Timeline struct {
	width, height int
	clips         []clip.Clip
	playheadPos   float64
	selected      int
}

func newTimeline(clips []clip.Clip) Timeline {
	return Timeline{
		clips: clips,
	}
}

func (t Timeline) SetSize(w, h int) Timeline {
	t.width = w
	t.height = h
	return t
}


func (t Timeline) Select(i int) Timeline {
	t.selected = i
	return t
}

func (t Timeline) Clips() []clip.Clip {
	return t.clips
}

func (t Timeline) Selected() int {
	return t.selected
}

func (t Timeline) Update(msg tea.Msg) (Timeline, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return t, nil
	}

	switch {
	case key.Matches(km, Keys.PlayheadLeft):
		t.playheadPos = max(0, t.playheadPos-1)
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.PlayheadRight):
		t.playheadPos = min(t.timelineTotal(), t.playheadPos+1)
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.FastLeft):
		t.playheadPos = max(0, t.playheadPos-5)
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.FastRight):
		t.playheadPos = min(t.timelineTotal(), t.playheadPos+5)
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.PrevBoundary):
		t.playheadPos = t.prevBoundary()
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.NextBoundary):
		t.playheadPos = t.nextBoundary()
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.Split):
		return t.splitAtPlayhead()
	case key.Matches(km, Keys.PrevFrame):
		t.playheadPos = t.prevFrame()
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.NextFrame):
		t.playheadPos = t.nextFrame()
		t.selected, _ = t.clipAtPlayhead()
	case key.Matches(km, Keys.Delete):
		return t.deleteSelected()
	}

	return t, nil
}

func (t Timeline) Render(focused bool) string {
	paneStyle := theme.Pane.Default
	headerS := theme.Pane.Header
	if focused {
		paneStyle = theme.Pane.Active
		headerS = theme.Pane.ActiveHeader
	}

	innerW := t.width - 2
	var lines []string
	lines = append(lines, headerS.Render("Timeline"))
	lines = append(lines, "")

	if len(t.clips) == 0 {
		lines = append(lines, theme.Dim.Render("  No clips loaded"))
	} else {
		total := t.timelineTotal()

		lines = append(lines, t.renderRuler(innerW, total))

		playheadCol := 0
		if total > 0 {
			playheadCol = int(t.playheadPos / total * float64(innerW))
			if playheadCol >= innerW {
				playheadCol = innerW - 1
			}
		}
		playheadRow := strings.Repeat(" ", playheadCol) + theme.Timeline.PlayheadArrow.Render("▼")
		lines = append(lines, playheadRow)

		var sb strings.Builder
		remaining := innerW
		colStart := 0
		for i, c := range t.clips {
			var w int
			if i == len(t.clips)-1 {
				w = remaining
			} else {
				w = max(1, int(c.Duration()/total*float64(innerW)))
				remaining -= w
			}
			label := truncate(filepath.Base(c.Path), w)
			runes := []rune(label + strings.Repeat(" ", max(0, w-lipgloss.Width(label))))

			barStyle := theme.Timeline.Bar
			if i == t.selected {
				barStyle = theme.Timeline.SelectedBar
			}

			localCol := playheadCol - colStart
			if localCol >= 0 && localCol < w {
				if localCol > 0 {
					sb.WriteString(barStyle.Render(string(runes[:localCol])))
				}
				sb.WriteString(theme.Timeline.PlayheadCursor.Render(string(runes[localCol : localCol+1])))
				if localCol+1 < w {
					sb.WriteString(barStyle.Render(string(runes[localCol+1:])))
				}
			} else {
				sb.WriteString(barStyle.Render(string(runes)))
			}

			colStart += w
		}
		lines = append(lines, sb.String())
		lines = append(lines, "")

		c := t.clips[t.selected]
		info := fmt.Sprintf("  %s   %s   %s   %.2ffps",
			filepath.Base(c.Path),
			formatDuration(c.Duration()),
			formatRes(c.Meta.Width, c.Meta.Height),
			c.Meta.FrameRate,
		)
		lines = append(lines, theme.Dim.Render(info))
		lines = append(lines, theme.Dim.Render(fmt.Sprintf("  Playhead  %s", formatPlayhead(t.playheadPos, c.Meta.FrameRate))))
	}

	return paneStyle.Width(t.width).Height(t.height).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (t Timeline) timelineTotal() float64 {
	total := 0.0
	for _, c := range t.clips {
		total += c.Duration()
	}
	return total
}

func (t Timeline) clipBoundaries() []float64 {
	b := make([]float64, 0, len(t.clips)+1)
	curr := 0.0
	b = append(b, curr)
	for _, c := range t.clips {
		curr += c.Duration()
		b = append(b, curr)
	}
	return b
}

func (t Timeline) clipAtPlayhead() (idx int, offsetInClip float64) {
	curr := 0.0
	for i, c := range t.clips {
		end := curr + c.Duration()
		if t.playheadPos < end || i == len(t.clips)-1 {
			return i, t.playheadPos - curr
		}
		curr = end
	}
	return 0, 0
}

func (t Timeline) splitAtPlayhead() (Timeline, tea.Cmd) {
	if len(t.clips) == 0 {
		return t, nil
	}
	idx, offset := t.clipAtPlayhead()
	if offset <= 0 || offset >= t.clips[idx].Duration() {
		return t, nil
	}
	orig := t.clips[idx]
	left := orig
	left.OutPoint = orig.InPoint + offset
	right := orig
	right.InPoint = orig.InPoint + offset

	t.clips = slices.Replace(t.clips, idx, idx+1, left, right)
	return t, nil
}

func (t Timeline) deleteSelected() (Timeline, tea.Cmd) {
	if len(t.clips) == 0 {
		return t, nil
	}
	t.clips = slices.Delete(t.clips, t.selected, t.selected+1)
	if len(t.clips) == 0 {
		t.selected = 0
		t.playheadPos = 0
		return t, nil
	}
	if t.selected >= len(t.clips) {
		t.selected = len(t.clips) - 1
	}
	newStart := 0.0
	for i := 0; i < t.selected; i++ {
		newStart += t.clips[i].Duration()
	}
	t.playheadPos = newStart
	return t, nil
}

func (t Timeline) nextBoundary() float64 {
	for _, b := range t.clipBoundaries() {
		if b > t.playheadPos+epsilon {
			return b
		}
	}
	return t.timelineTotal()
}

func (t Timeline) prevBoundary() float64 {
	boundaries := t.clipBoundaries()
	for i := len(boundaries) - 1; i >= 0; i-- {
		if boundaries[i] < t.playheadPos-epsilon {
			return boundaries[i]
		}
	}
	return 0
}

func (t Timeline) currentFPS() float64 {
	if len(t.clips) == 0 {
		return defaultFPS
	}
	idx, _ := t.clipAtPlayhead()
	fps := t.clips[idx].Meta.FrameRate
	if fps <= 0 {
		return defaultFPS
	}
	return fps
}

func (t Timeline) prevFrame() float64 {
	fps := t.currentFPS()
	frame := math.Round(t.playheadPos * fps)
	return max(0, (frame-1)/fps)
}

func (t Timeline) nextFrame() float64 {
	fps := t.currentFPS()
	frame := math.Round(t.playheadPos * fps)
	return min(t.timelineTotal(), (frame+1)/fps)
}

func (t Timeline) ShortHelp() []key.Binding {
	return []key.Binding{Keys.PlayheadLeft, Keys.PlayheadRight, Keys.Split, Keys.Delete, Keys.Export, Keys.Help}
}

func (t Timeline) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{Keys.Quit, Keys.ForceQuit, Keys.NextPane, Keys.PrevPane, Keys.Export},
		{Keys.PlayheadLeft, Keys.PlayheadRight, Keys.FastLeft, Keys.FastRight},
		{Keys.PrevBoundary, Keys.NextBoundary, Keys.PrevFrame, Keys.NextFrame},
		{Keys.Split, Keys.Delete, Keys.Help},
	}
}

func (t Timeline) renderRuler(innerW int, total float64) string {
	buf := []rune(strings.Repeat(" ", innerW))
	interval := rulerInterval(total, innerW)
	for curr := 0.0; curr <= total; curr += interval {
		col := int(curr / total * float64(innerW))
		if col >= innerW {
			col = innerW - 1
		}
		for i, r := range []rune(formatRulerMark(curr)) {
			if col+i < innerW {
				buf[col+i] = r
			}
		}
	}
	return theme.Dim.Render(string(buf))
}

func rulerInterval(total float64, width int) float64 {
	labelWidth := float64(len(formatRulerMark(total))) + 2
	minInterval := labelWidth * total / float64(width)
	for _, s := range []float64{1, 2, 5, 10, 15, 20, 30, 60, 120, 300, 600, 1800, 3600} {
		if s >= minInterval {
			return s
		}
	}
	return 3600
}

func formatRulerMark(secs float64) string {
	h, m, s := decomposeSecs(secs)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
