package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/panoramix360/edity/internal/clip"
	"github.com/panoramix360/edity/internal/ffprobe"
	"github.com/panoramix360/edity/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: edity <video1> [video2 ...]")
		os.Exit(1)
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Fprintln(os.Stderr, "error: ffmpeg not found in PATH")
		os.Exit(1)
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		fmt.Fprintln(os.Stderr, "error: ffprobe not found in PATH")
		os.Exit(1)
	}

	paths := os.Args[1:]
	clips := make([]clip.Clip, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(os.Stderr, "error: file not found: %s\n", path)
			os.Exit(1)
		}
		meta, err := ffprobe.Probe(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		clips = append(clips, clip.Clip{Path: path, Meta: meta})
	}

	p := tea.NewProgram(ui.New(clips))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
