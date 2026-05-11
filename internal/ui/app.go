package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
)

type Pane int

const (
	PaneMediaBin Pane = iota
	PanePreview
	PaneTimeline
)

var (
	borderColor       = lipgloss.Color("8")
	activeBorderColor = lipgloss.Color("10")

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(activeBorderColor)

	headerStyle = lipgloss.NewStyle().
			Foreground(borderColor).
			Bold(true)

	activeHeaderStyle = lipgloss.NewStyle().
				Foreground(activeBorderColor).
				Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	selectedBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("10")).
				Foreground(lipgloss.Color("0"))

	normalBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(1)

	activePaneLabel = lipgloss.NewStyle().
			Foreground(activeBorderColor).
			Bold(true)
)

type Model struct {
	clips       []clip.Clip
	selected    int
	focusedPane Pane
	playheadPos float64
	width       int
	height      int
}

func New(clips []clip.Clip) Model {
	return Model{clips: clips}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focusedPane = (m.focusedPane + 1) % 3
		case "shift+tab":
			m.focusedPane = (m.focusedPane + 2) % 3
		case "up", "k":
			if m.focusedPane == PaneMediaBin && m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.focusedPane == PaneMediaBin && m.selected < len(m.clips)-1 {
				m.selected++
			}
		case "left", "h":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = max(0, m.playheadPos-1)
			}
		case "right", "l":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = min(m.timelineTotal(), m.playheadPos+1)
			}
		case "shift+left":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = max(0, m.playheadPos-5)
			}
		case "shift+right":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = min(m.timelineTotal(), m.playheadPos+5)
			}
		case "[":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.prevBoundary()
			}
		case "]":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.nextBoundary()
			}
		case "s":
			if m.focusedPane == PaneTimeline {
				m = m.splitAtPlayhead()
			}
		}
	}
	return m, nil
}

func (m Model) timelineTotal() float64 {
	total := 0.0
	for _, c := range m.clips {
		total += c.Duration()
	}
	return total
}

func (m Model) clipBoundaries() []float64 {
	b := make([]float64, 0, len(m.clips)+1)
	t := 0.0
	b = append(b, t)
	for _, c := range m.clips {
		t += c.Duration()
		b = append(b, t)
	}
	return b
}

func (m Model) clipAtPlayhead() (idx int, offsetInClip float64) {
	t := 0.0
	for i, c := range m.clips {
		end := t + c.Duration()
		if m.playheadPos < end || i == len(m.clips)-1 {
			return i, m.playheadPos - t
		}
		t = end
	}
	return 0, 0
}

func (m Model) splitAtPlayhead() Model {
	if len(m.clips) == 0 {
		return m
	}
	idx, offset := m.clipAtPlayhead()
	if offset <= 0 || offset >= m.clips[idx].Duration() {
		return m
	}
	orig := m.clips[idx]
	left := orig
	left.OutPoint = orig.InPoint + offset
	right := orig
	right.InPoint = orig.InPoint + offset

	newClips := make([]clip.Clip, 0, len(m.clips)+1)
	newClips = append(newClips, m.clips[:idx]...)
	newClips = append(newClips, left, right)
	newClips = append(newClips, m.clips[idx+1:]...)
	m.clips = newClips
	return m
}

func (m Model) nextBoundary() float64 {
	for _, b := range m.clipBoundaries() {
		if b > m.playheadPos+0.01 {
			return b
		}
	}
	return m.timelineTotal()
}

func (m Model) prevBoundary() float64 {
	boundaries := m.clipBoundaries()
	for i := len(boundaries) - 1; i >= 0; i-- {
		if boundaries[i] < m.playheadPos-0.01 {
			return boundaries[i]
		}
	}
	return 0
}

func (m Model) renderRuler(innerW int, total float64) string {
	buf := []rune(strings.Repeat(" ", innerW))
	interval := rulerInterval(total, innerW)
	for t := 0.0; t <= total; t += interval {
		col := int(t / total * float64(innerW))
		if col >= innerW {
			col = innerW - 1
		}
		for i, r := range []rune(formatRulerMark(t)) {
			if col+i < innerW {
				buf[col+i] = r
			}
		}
	}
	return dimStyle.Render(string(buf))
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

func (m Model) paneStyle(p Pane) lipgloss.Style {
	if m.focusedPane == p {
		return activePaneStyle
	}
	return paneStyle
}

func (m Model) headerFor(p Pane, label string) string {
	if m.focusedPane == p {
		return activeHeaderStyle.Render(label)
	}
	return headerStyle.Render(label)
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}

	paneNames := []string{"Media Bin", "Preview", "Timeline"}
	activeLabel := activePaneLabel.Render("[" + paneNames[m.focusedPane] + "]")
	statusBar := statusStyle.Render("q quit   tab/shift+tab focus   ↑↓/jk navigate   ←→/hl playhead   shift+←→ fast   [/] boundaries   " + activeLabel)
	statusHeight := 1

	topHeight := (m.height - statusHeight) * 2 / 3
	bottomHeight := m.height - topHeight - statusHeight

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	mediaBin := m.renderMediaBin(leftWidth, topHeight)
	preview := m.renderPreview(rightWidth, topHeight, m.clips[m.selected])
	top := lipgloss.JoinHorizontal(lipgloss.Top, mediaBin, preview)

	timeline := m.renderTimeline(m.width, bottomHeight)

	content := lipgloss.JoinVertical(lipgloss.Left, top, timeline, statusBar)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) renderMediaBin(width, height int) string {
	inner := width - 2
	var lines []string
	lines = append(lines, m.headerFor(PaneMediaBin, "Media Bin"))
	lines = append(lines, "")

	for i, c := range m.clips {
		name := filepath.Base(c.Path)
		dur := formatDuration(c.Duration())
		res := formatRes(c.Meta.Width, c.Meta.Height)
		fps := fmt.Sprintf("%.2ffps", c.Meta.FrameRate)

		nameLine := truncate(name, inner-7) + "  " + dimStyle.Render(dur)
		metaLine := dimStyle.Render("  " + res + " " + fps)

		if i == m.selected {
			lines = append(lines, selectedStyle.Render("> "+nameLine))
		} else {
			lines = append(lines, "  "+nameLine)
		}
		lines = append(lines, metaLine)
	}

	return m.paneStyle(PaneMediaBin).Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderPreview(width, height int, c clip.Clip) string {
	lines := []string{
		m.headerFor(PanePreview, "Preview"),
		"",
		"  " + filepath.Base(c.Path),
		"",
		dimStyle.Render("  Duration   " + formatDuration(c.Duration())),
		dimStyle.Render("  Video      " + formatRes(c.Meta.Width, c.Meta.Height) + "  " + fmt.Sprintf("%.2ffps", c.Meta.FrameRate)),
		dimStyle.Render("  Codec      " + c.Meta.Codec),
		"",
		dimStyle.Render("  Playback coming in v0.3"),
	}
	return m.paneStyle(PanePreview).Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderTimeline(width, height int) string {
	innerW := width - 2
	var lines []string
	lines = append(lines, m.headerFor(PaneTimeline, "Timeline"))
	lines = append(lines, "")

	if len(m.clips) == 0 {
		lines = append(lines, dimStyle.Render("  No clips loaded"))
	} else {
		total := m.timelineTotal()

		lines = append(lines, m.renderRuler(innerW, total))

		playheadCol := 0
		if total > 0 {
			playheadCol = int(m.playheadPos / total * float64(innerW))
			if playheadCol >= innerW {
				playheadCol = innerW - 1
			}
		}
		playheadRow := strings.Repeat(" ", playheadCol) + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("▼")
		lines = append(lines, playheadRow)

		var sb strings.Builder
		remaining := innerW
		for i, c := range m.clips {
			var w int
			if i == len(m.clips)-1 {
				w = remaining
			} else {
				w = max(1, int(c.Duration()/total*float64(innerW)))
				remaining -= w
			}
			label := truncate(filepath.Base(c.Path), w)
			padded := label + strings.Repeat(" ", w-len([]rune(label)))
			if i == m.selected {
				sb.WriteString(selectedBarStyle.Render(padded))
			} else {
				sb.WriteString(normalBarStyle.Render(padded))
			}
		}
		lines = append(lines, sb.String())
		lines = append(lines, "")

		c := m.clips[m.selected]
		info := fmt.Sprintf("  %s   %s   %s   %.2ffps",
			filepath.Base(c.Path),
			formatDuration(c.Duration()),
			formatRes(c.Meta.Width, c.Meta.Height),
			c.Meta.FrameRate,
		)
		lines = append(lines, dimStyle.Render(info))
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  Playhead  %s", formatPlayhead(m.playheadPos))))
	}

	return m.paneStyle(PaneTimeline).Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func decomposeSecs(secs float64) (h, m, s int) {
	h = int(secs) / 3600
	m = (int(secs) % 3600) / 60
	s = int(secs) % 60
	return
}

func formatRulerMark(secs float64) string {
	h, m, s := decomposeSecs(secs)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatDuration(secs float64) string {
	h, m, s := decomposeSecs(secs)
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func formatPlayhead(secs float64) string {
	h, m, s := decomposeSecs(secs)
	cs := int(secs*100) % 100
	return fmt.Sprintf("%02d:%02d:%02d.%02d", h, m, s, cs)
}

func formatRes(w, h int) string {
	if w == 0 {
		return "unknown"
	}
	return fmt.Sprintf("%dx%d", w, h)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if max <= 0 {
		return ""
	}
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
