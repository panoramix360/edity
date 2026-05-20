package ui

import (
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/panoramix360/edity/internal/clip"
	ffmpegpkg "github.com/panoramix360/edity/internal/ffmpeg"
)

type exportStage int

const (
	exportIdle exportStage = iota
	exportInput
	exportRunning
	exportDone
	exportFailed
)

type exportProgressMsg float64
type exportDoneMsg struct {
	outputPath string
	err        error
}

const (
	modalW       = 58
	modalContent = modalW - 8
)

var (
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 3)

	modalTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	modalHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	modalSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	modalErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	keyChipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("12")).
			Padding(0, 1)
)

type ExportModal struct {
	stage    exportStage
	input    textinput.Model
	progress progress.Model
	spinner  spinner.Model
	output   string
	errMsg   string
	ch       chan float64
	clips    []clip.Clip
	total    float64
}

func newExportModal() ExportModal {
	ti := textinput.New()
	ti.Placeholder = "output.mp4"
	ti.SetValue("output.mp4")
	ti.SetVirtualCursor(false)
	ti.SetWidth(modalContent)

	return ExportModal{input: ti}
}

func (m ExportModal) Init() tea.Cmd {
	return textinput.Blink
}

func (m ExportModal) Open(clips []clip.Clip, total float64) ExportModal {
	m.clips = clips
	m.total = total
	m.stage = exportInput
	m.input.Focus()
	return m
}

func (m ExportModal) IsActive() bool {
	return m.stage != exportIdle
}

func (m ExportModal) Update(msg tea.Msg) (ExportModal, tea.Cmd) {
	switch msg := msg.(type) {
	case exportProgressMsg:
		cmd := m.progress.SetPercent(float64(msg))
		return m, tea.Batch(listenProgress(m.ch), cmd)
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case exportDoneMsg:
		if msg.err != nil {
			m.stage = exportFailed
			m.errMsg = msg.err.Error()
		} else {
			m.stage = exportDone
			m.output = msg.outputPath
		}
		return m, nil
	}

	var tiCmd tea.Cmd
	if m.stage == exportInput {
		m.input, tiCmd = m.input.Update(msg)
	}

	if km, ok := msg.(tea.KeyPressMsg); ok {
		switch m.stage {
		case exportInput:
			switch km.String() {
			case "esc":
				m.stage = exportIdle
			case "enter":
				if m.input.Value() == "" {
					return m, tiCmd
				}
				outputPath := m.resolveOutputPath()
				m.stage = exportRunning
				m.progress = progress.New(progress.WithDefaultBlend())
				m.progress.SetWidth(modalContent)
				m.ch = make(chan float64, 50)
				m.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
				return m, tea.Batch(tiCmd, m.runExportCmd(outputPath), listenProgress(m.ch), m.spinner.Tick)
			}
		case exportDone, exportFailed:
			switch km.String() {
			case "esc", "enter":
				m.stage = exportIdle
			}
		}
	}

	return m, tiCmd
}

func (m ExportModal) Render(termW, termH int) string {
	var body string
	switch m.stage {
	case exportInput:
		body = m.inputBody()
	case exportRunning:
		body = m.runningBody()
	case exportDone:
		body = m.doneBody()
	case exportFailed:
		body = m.failedBody()
	}
	modal := modalStyle.Width(modalW).Render(body)
	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, modal)
}

func (m ExportModal) Cursor(termW, termH int) *tea.Cursor {
	if m.stage != exportInput {
		return nil
	}
	modal := modalStyle.Width(modalW).Render(m.inputBody())
	fgH := lipgloss.Height(modal)
	fgW := lipgloss.Width(modal)
	startY := max((termH-fgH)/2, 0)
	startX := max((termW-fgW)/2, 0)
	c := m.input.Cursor()
	c.Y += startY + 5
	c.X += startX + 4
	return c
}

func (m ExportModal) resolveOutputPath() string {
	name := m.input.Value()
	cwd, err := os.Getwd()
	if err != nil {
		return name
	}
	return filepath.Join(cwd, name)
}

func (m ExportModal) runExportCmd(outputPath string) tea.Cmd {
	clips := m.clips
	total := m.total
	ch := m.ch
	return func() tea.Msg {
		err := ffmpegpkg.Export(clips, outputPath, total, func(p float64) {
			ch <- p
		})
		close(ch)
		return exportDoneMsg{outputPath: outputPath, err: err}
	}
}

func listenProgress(ch chan float64) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return exportProgressMsg(p)
	}
}

func (m ExportModal) inputBody() string {
	cwd, _ := os.Getwd()
	hints := lipgloss.JoinHorizontal(lipgloss.Left,
		keyChipStyle.Render("enter"), modalHintStyle.Render(" Confirm   "),
		keyChipStyle.Render("esc"), modalHintStyle.Render(" Cancel"),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		modalTitleStyle.Render("Export"),
		"",
		"Output filename:",
		lipgloss.NewStyle().Width(modalContent).Render(m.input.View()),
		"",
		modalHintStyle.Render("Directory: "+truncate(cwd, modalContent)),
		"",
		hints,
	)
}

func (m ExportModal) runningBody() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		modalTitleStyle.Render("Exporting..."),
		"",
		m.progress.View(),
		"",
		modalHintStyle.Render(m.spinner.View()+" Exporting"),
	)
}

func (m ExportModal) doneBody() string {
	hints := lipgloss.JoinHorizontal(lipgloss.Left,
		keyChipStyle.Render("enter"), modalHintStyle.Render(" Close"),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		modalSuccessStyle.Render("Export complete"),
		"",
		"Output:",
		dimStyle.Render(truncate(m.output, modalContent)),
		"",
		hints,
	)
}

func (m ExportModal) failedBody() string {
	errLines := strings.Split(m.errMsg, "\n")
	var shown []string
	for _, l := range errLines {
		shown = append(shown, truncate(l, modalContent))
		if len(shown) >= 6 {
			break
		}
	}
	hints := lipgloss.JoinHorizontal(lipgloss.Left,
		keyChipStyle.Render("enter"), modalHintStyle.Render(" Close"),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		modalErrorStyle.Render("Export failed"),
		"",
		dimStyle.Render(strings.Join(shown, "\n")),
		"",
		hints,
	)
}
