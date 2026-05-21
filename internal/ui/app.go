package ui

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"
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

const (
	epsilon    = 0.001
	defaultFPS = 30.0
)

var (
	borderColor       = lipgloss.Color("8")
	activeBorderColor = lipgloss.Color("10")

	defaultPaneStyle = lipgloss.NewStyle().
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

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	selectedBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("10")).
				Foreground(lipgloss.Color("0"))

	normalBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15"))

	playheadCursorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("9")).
				Foreground(lipgloss.Color("15")).
				Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(1)

	activePaneLabel = lipgloss.NewStyle().
			Foreground(activeBorderColor).
			Bold(true)
)

type Model struct {
	clips                []clip.Clip
	selected             int
	focusedPane          Pane
	playheadPos          float64
	width                int
	height               int
	binW, binH           int
	previewW, previewH   int
	timelineW, timelineH int
	bin                  MediaBin
	modal                ExportModal
}

func New(clips []clip.Clip) Model {
	return Model{
		clips: clips,
		bin:   newMediaBin(clips),
		modal: newExportModal(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.modal.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wm.Width
		m.height = wm.Height

		statusH := 1

		m.binW = m.width / 3
		m.binH = (m.height - statusH) * 2 / 3

		m.timelineW = m.width
		m.timelineH = m.height - m.binH - statusH

		m.previewW = m.width - m.binW
		m.previewH = m.binH

		m.bin = m.bin.SetSize(m.binW-2, m.binH-2)
		return m, nil
	}

	if m.modal.IsActive() {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}

	km, isKeyPress := msg.(tea.KeyPressMsg)

	if isKeyPress {
		switch km.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focusedPane = (m.focusedPane + 1) % 3
			return m, nil
		case "shift+tab":
			m.focusedPane = (m.focusedPane + 2) % 3
			return m, nil
		}
	}

	switch m.focusedPane {
	case PaneMediaBin:
		if isKeyPress {
			switch km.String() {
			case "q":
				if !m.bin.IsFiltering() {
					return m, tea.Quit
				}
			case "e":
				if !m.bin.IsFiltering() {
					return m.openExportModal()
				}
			}
		}
		var cmd tea.Cmd
		m.bin, cmd = m.bin.Update(msg)
		m.selected = m.bin.GlobalIndex()
		return m, cmd

	default:
		if !isKeyPress {
			return m, nil
		}
		switch km.String() {
		case "q":
			return m, tea.Quit
		case "e":
			return m.openExportModal()
		case "left", "h":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = max(0, m.playheadPos-1)
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "right", "l":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = min(m.timelineTotal(), m.playheadPos+1)
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "shift+left":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = max(0, m.playheadPos-5)
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "shift+right":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = min(m.timelineTotal(), m.playheadPos+5)
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "[":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.prevBoundary()
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "]":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.nextBoundary()
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "s":
			if m.focusedPane == PaneTimeline {
				return m.splitAtPlayhead()
			}
		case ",":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.prevFrame()
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case ".":
			if m.focusedPane == PaneTimeline {
				m.playheadPos = m.nextFrame()
				m.selected, _ = m.clipAtPlayhead()
				m.bin = m.bin.Select(m.selected)
			}
		case "d", "backspace":
			if m.focusedPane == PaneTimeline {
				return m.deleteSelected()
			}
		}
	}
	return m, nil
}

func (m Model) openExportModal() (Model, tea.Cmd) {
	if len(m.clips) == 0 {
		return m, nil
	}
	m.modal = m.modal.Open(m.clips, m.timelineTotal())
	return m, nil
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}
	var content string
	if m.modal.IsActive() {
		content = m.modal.Render(m.width, m.height)
	} else {
		content = m.renderMainView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	v.Cursor = m.modal.Cursor(m.width, m.height)
	return v
}

func (m Model) renderMainView() string {
	paneNames := []string{"Media Bin", "Preview", "Timeline"}
	activeLabel := activePaneLabel.Render("[" + paneNames[m.focusedPane] + "]")
	statusBar := statusStyle.Render("q quit   tab/shift+tab focus   ↑↓/jk navigate   ←→/hl playhead   shift+←→ fast   [/] boundaries   ,/. frame   s split   d delete   e export   " + activeLabel)
	mediaBin := m.bin.Render(m.binW, m.binH, m.focusedPane == PaneMediaBin)
	preview := m.renderPreview(m.previewW, m.previewH)
	top := lipgloss.JoinHorizontal(lipgloss.Top, mediaBin, preview)
	timeline := m.renderTimeline(m.timelineW, m.timelineH)

	return lipgloss.JoinVertical(lipgloss.Left, top, timeline, statusBar)
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

func (m Model) splitAtPlayhead() (Model, tea.Cmd) {
	if len(m.clips) == 0 {
		return m, nil
	}
	idx, offset := m.clipAtPlayhead()
	if offset <= 0 || offset >= m.clips[idx].Duration() {
		return m, nil
	}
	orig := m.clips[idx]
	left := orig
	left.OutPoint = orig.InPoint + offset
	right := orig
	right.InPoint = orig.InPoint + offset

	m.clips = slices.Replace(m.clips, idx, idx+1, left, right)
	var cmd tea.Cmd
	m.bin, cmd = m.bin.SetClips(m.clips)
	return m, cmd
}

func (m Model) deleteSelected() (Model, tea.Cmd) {
	if len(m.clips) == 0 {
		return m, nil
	}
	m.clips = slices.Delete(m.clips, m.selected, m.selected+1)
	if len(m.clips) == 0 {
		m.selected = 0
		m.playheadPos = 0
		var cmd tea.Cmd
		m.bin, cmd = m.bin.SetClips(m.clips)
		return m, cmd
	}
	if m.selected >= len(m.clips) {
		m.selected = len(m.clips) - 1
	}
	newStart := 0.0
	for i := 0; i < m.selected; i++ {
		newStart += m.clips[i].Duration()
	}
	m.playheadPos = newStart
	var cmd tea.Cmd
	m.bin, cmd = m.bin.SetClips(m.clips)
	m.bin = m.bin.Select(m.selected)
	return m, cmd
}

func (m Model) nextBoundary() float64 {
	for _, b := range m.clipBoundaries() {
		if b > m.playheadPos+epsilon {
			return b
		}
	}
	return m.timelineTotal()
}

func (m Model) prevBoundary() float64 {
	boundaries := m.clipBoundaries()
	for i := len(boundaries) - 1; i >= 0; i-- {
		if boundaries[i] < m.playheadPos-epsilon {
			return boundaries[i]
		}
	}
	return 0
}

func (m Model) currentFPS() float64 {
	if len(m.clips) == 0 {
		return defaultFPS
	}
	idx, _ := m.clipAtPlayhead()
	fps := m.clips[idx].Meta.FrameRate
	if fps <= 0 {
		return defaultFPS
	}
	return fps
}

func (m Model) prevFrame() float64 {
	fps := m.currentFPS()
	frame := math.Round(m.playheadPos * fps)
	return max(0, (frame-1)/fps)
}

func (m Model) nextFrame() float64 {
	fps := m.currentFPS()
	frame := math.Round(m.playheadPos * fps)
	return min(m.timelineTotal(), (frame+1)/fps)
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
	return defaultPaneStyle
}

func (m Model) headerFor(p Pane, label string) string {
	if m.focusedPane == p {
		return activeHeaderStyle.Render(label)
	}
	return headerStyle.Render(label)
}

func (m Model) renderPreview(width, height int) string {
	if len(m.clips) == 0 {
		return m.paneStyle(PanePreview).Width(width).Height(height).Render(m.headerFor(PanePreview, "Preview"))
	}
	c := m.clips[m.selected]
	lines := []string{
		m.headerFor(PanePreview, "Preview"),
		"",
		"  " + filepath.Base(c.Path),
		"",
		dimStyle.Render("  Duration   " + formatDuration(c.Duration())),
		dimStyle.Render("  Video      " + formatRes(c.Meta.Width, c.Meta.Height) + "  " + fmt.Sprintf("%.2ffps", c.Meta.FrameRate)),
		dimStyle.Render("  Codec      " + c.Meta.Codec),
		"",
		dimStyle.Render("  Playback coming in v0.4"),
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
		colStart := 0
		for i, c := range m.clips {
			var w int
			if i == len(m.clips)-1 {
				w = remaining
			} else {
				w = max(1, int(c.Duration()/total*float64(innerW)))
				remaining -= w
			}
			label := truncate(filepath.Base(c.Path), w)
			runes := []rune(label + strings.Repeat(" ", max(0, w-lipgloss.Width(label))))

			barStyle := normalBarStyle
			if i == m.selected {
				barStyle = selectedBarStyle
			}

			localCol := playheadCol - colStart
			if localCol >= 0 && localCol < w {
				if localCol > 0 {
					sb.WriteString(barStyle.Render(string(runes[:localCol])))
				}
				sb.WriteString(playheadCursorStyle.Render(string(runes[localCol : localCol+1])))
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

		c := m.clips[m.selected]
		info := fmt.Sprintf("  %s   %s   %s   %.2ffps",
			filepath.Base(c.Path),
			formatDuration(c.Duration()),
			formatRes(c.Meta.Width, c.Meta.Height),
			c.Meta.FrameRate,
		)
		lines = append(lines, dimStyle.Render(info))
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  Playhead  %s", formatPlayhead(m.playheadPos, c.Meta.FrameRate))))
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

func formatPlayhead(secs, fps float64) string {
	if fps <= 0 {
		fps = defaultFPS
	}
	fpsInt := int(math.Round(fps))
	totalFrames := int(math.Round(secs * fps))
	ff := totalFrames % fpsInt
	totalSecs := totalFrames / fpsInt
	s := totalSecs % 60
	m := (totalSecs / 60) % 60
	h := totalSecs / 3600
	frameWidth := 2
	if fpsInt >= 100 {
		frameWidth = 3
	}
	return fmt.Sprintf("%02d:%02d:%02d:%0*d", h, m, s, frameWidth, ff)
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
