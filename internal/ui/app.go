package ui

import (
	"fmt"
	"math"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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

)

type Model struct {
	clips                []clip.Clip
	selected             int
	focusedPane          Pane
	width                int
	height               int
	binW, binH           int
	previewW, previewH   int
	timelineW, timelineH int
	bin                  MediaBin
	timeline             Timeline
	preview              Preview
	modal                ExportModal
	help                 help.Model
}

func New(clips []clip.Clip) Model {
	return Model{
		clips:    clips,
		bin:      newMediaBin(clips),
		timeline: newTimeline(clips),
		preview:  newPreview(),
		modal:    newExportModal(),
		help:     help.New(),
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
		m.timeline = m.timeline.SetSize(m.timelineW, m.timelineH)
		m.preview = m.preview.SetSize(m.previewW, m.previewH)
		m.help.SetWidth(m.width)
		return m, nil
	}

	if m.modal.IsActive() {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}

	km, isKeyPress := msg.(tea.KeyPressMsg)

	if isKeyPress {
		switch {
		case key.Matches(km, Keys.ForceQuit):
			return m, tea.Quit
		case key.Matches(km, Keys.NextPane):
			m.focusedPane = (m.focusedPane + 1) % 3
			return m, nil
		case key.Matches(km, Keys.PrevPane):
			m.focusedPane = (m.focusedPane + 2) % 3
			return m, nil
		case key.Matches(km, Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}

	switch m.focusedPane {
	case PaneMediaBin:
		if isKeyPress {
			switch {
			case key.Matches(km, Keys.Quit) && !m.bin.IsFiltering():
				return m, tea.Quit
			case key.Matches(km, Keys.Export) && !m.bin.IsFiltering():
				return m.openExportModal()
			}
		}
		var cmd tea.Cmd
		m.bin, cmd = m.bin.Update(msg)
		m.selected = m.bin.GlobalIndex()
		m.timeline = m.timeline.Select(m.selected)
		return m, cmd

	case PaneTimeline:
		if isKeyPress {
			switch {
			case key.Matches(km, Keys.Quit):
				return m, tea.Quit
			case key.Matches(km, Keys.Export):
				return m.openExportModal()
			}
		}

		var cmd tea.Cmd
		m.timeline, cmd = m.timeline.Update(msg)
		m.clips = m.timeline.Clips()
		m.selected = m.timeline.Selected()

		var binCmd tea.Cmd
		m.bin, binCmd = m.bin.SetClips(m.clips)
		m.bin = m.bin.Select(m.selected)

		return m, tea.Batch(cmd, binCmd)
	}

	return m, nil
}

func (m Model) openExportModal() (Model, tea.Cmd) {
	if len(m.clips) == 0 {
		return m, nil
	}
	m.modal = m.modal.Open(m.clips, m.timeline.timelineTotal())
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
	mediaBin := m.bin.Render(m.binW, m.binH, m.focusedPane == PaneMediaBin)
	preview := m.preview.Render(m.clips, m.selected, m.focusedPane == PanePreview)
	top := lipgloss.JoinHorizontal(lipgloss.Top, mediaBin, preview)
	timeline := m.timeline.Render(m.focusedPane == PaneTimeline)
	helpBar := lipgloss.NewStyle().PaddingLeft(1).Render(m.help.View(Keys))

	return lipgloss.JoinVertical(lipgloss.Left, top, timeline, helpBar)
}

func decomposeSecs(secs float64) (h, m, s int) {
	h = int(secs) / 3600
	m = (int(secs) % 3600) / 60
	s = int(secs) % 60
	return
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
