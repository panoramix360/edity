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
	width       int
	height      int
}

func New(clips []clip.Clip) Model {
	return Model{clips: clips}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Q: I believe we would need to map this method in a more splitted way to not clutter it
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
		}
	}
	return m, nil
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
	statusBar := statusStyle.Render("q quit   tab/shift+tab focus   ↑↓/jk navigate   " + activeLabel)
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
		dur := formatDuration(c.Meta.Duration)
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
		dimStyle.Render("  Duration   " + formatDuration(c.Meta.Duration)),
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
		total := 0.0
		for _, c := range m.clips {
			total += c.Meta.Duration
		}

		var sb strings.Builder
		remaining := innerW
		for i, c := range m.clips {
			var w int
			if i == len(m.clips)-1 {
				w = remaining
			} else {
				w = max(1, int(c.Meta.Duration/total*float64(innerW)))
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
			formatDuration(c.Meta.Duration),
			formatRes(c.Meta.Width, c.Meta.Height),
			c.Meta.FrameRate,
		)
		lines = append(lines, dimStyle.Render(info))
	}

	return m.paneStyle(PaneTimeline).Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func formatDuration(secs float64) string {
	h := int(secs) / 3600
	m := (int(secs) % 3600) / 60
	s := int(secs) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
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
