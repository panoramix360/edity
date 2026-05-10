package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
)

var (
	borderColor = lipgloss.Color("2") // terminal green

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	headerStyle = lipgloss.NewStyle().
			Foreground(borderColor).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(1)
)

type Model struct {
	clips    []clip.Clip
	selected int
	width    int
	height   int
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
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.clips)-1 {
				m.selected++
			}
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}

	statusBar := statusStyle.Render("q quit   ↑↓/jk navigate")
	statusHeight := 1

	topHeight := (m.height - statusHeight) * 2 / 3
	bottomHeight := m.height - topHeight - statusHeight

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	mediaBin := m.renderMediaBin(leftWidth, topHeight)
	preview := m.renderPreview(rightWidth, topHeight)
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
	lines = append(lines, headerStyle.Render("Media Bin"))
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

	return paneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderPreview(width, height int) string {
	content := headerStyle.Render("Preview") + "\n\n" +
		dimStyle.Render("  No preview in v0.1\n\n") +
		dimStyle.Render("  Press P to open in ffplay")

	return paneStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderTimeline(width, height int) string {
	innerW := width - 2
	var lines []string
	lines = append(lines, headerStyle.Render("Timeline"))
	lines = append(lines, "")

	if len(m.clips) == 0 {
		lines = append(lines, dimStyle.Render("  No clips loaded"))
	} else {
		total := 0.0
		for _, c := range m.clips {
			total += c.Meta.Duration
		}

		fills := []string{"█", "▓", "░"}
		var sb strings.Builder
		for i, c := range m.clips {
			ratio := c.Meta.Duration / total
			w := max(1, int(ratio*float64(innerW)))
			label := truncate(filepath.Base(c.Path), w)
			block := label + strings.Repeat(fills[i%len(fills)], w-len([]rune(label)))
			if i == m.selected {
				sb.WriteString(selectedStyle.Render(block))
			} else {
				sb.WriteString(block)
			}
		}
		lines = append(lines, " "+sb.String())
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

	return paneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
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
