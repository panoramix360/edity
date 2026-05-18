package ffmpeg

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/panoramix360/edity/internal/clip"
)

var timeRe = regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == '\r' || b == '\n' {
			skip := 1
			if b == '\r' && i+1 < len(data) && data[i+1] == '\n' {
				skip = 2
			}
			return i + skip, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func Export(clips []clip.Clip, outputPath string, totalDuration float64, onProgress func(float64)) error {
	if len(clips) == 0 {
		return fmt.Errorf("no clips to export")
	}

	first := clips[0].Meta
	targetW, targetH := first.Width, first.Height
	targetFPS := first.FrameRate
	if targetW == 0 {
		targetW, targetH = 1920, 1080
	}
	if targetFPS <= 0 {
		targetFPS = 30
	}

	withAudio := false
	for _, c := range clips {
		if c.Meta.HasAudio {
			withAudio = true
			break
		}
	}

	args := buildArgs(clips, outputPath, targetW, targetH, targetFPS, withAudio)
	cmd := exec.Command("ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	var tail []string
	scanner := bufio.NewScanner(stderr)
	scanner.Split(scanLines)
	for scanner.Scan() {
		line := scanner.Text()
		tail = append(tail, line)
		if len(tail) > 12 {
			tail = tail[1:]
		}
		m := timeRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		h, _ := strconv.Atoi(m[1])
		mn, _ := strconv.Atoi(m[2])
		s, _ := strconv.Atoi(m[3])
		cs, _ := strconv.Atoi(m[4])
		elapsed := float64(h*3600+mn*60+s) + float64(cs)/100.0
		if totalDuration > 0 && onProgress != nil {
			p := elapsed / totalDuration
			if p > 1.0 {
				p = 1.0
			}
			onProgress(p)
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w\n%s", err, strings.Join(tail, "\n"))
	}
	return nil
}

func buildArgs(clips []clip.Clip, outputPath string, w, h int, fps float64, withAudio bool) []string {
	args := []string{}

	for _, c := range clips {
		args = append(args,
			"-ss", fmt.Sprintf("%.6f", c.InPoint),
			"-t", fmt.Sprintf("%.6f", c.Duration()),
			"-i", c.Path,
		)
	}

	var filters []string
	var concatInputs strings.Builder
	n := len(clips)

	for i, c := range clips {
		vOut := fmt.Sprintf("[v%d]", i)
		filters = append(filters, fmt.Sprintf(
			"[%d:v]scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,fps=%.3f,setsar=1%s",
			i, w, h, w, h, fps, vOut,
		))
		concatInputs.WriteString(vOut)

		if withAudio {
			aOut := fmt.Sprintf("[a%d]", i)
			if c.Meta.HasAudio {
				filters = append(filters, fmt.Sprintf("[%d:a]aresample=48000%s", i, aOut))
			} else {
				filters = append(filters, fmt.Sprintf(
					"anullsrc=r=48000:cl=stereo,atrim=duration=%.6f,aresample=48000%s",
					c.Duration(), aOut,
				))
			}
			concatInputs.WriteString(aOut)
		}
	}

	if withAudio {
		filters = append(filters, fmt.Sprintf(
			"%sconcat=n=%d:v=1:a=1[outv][outa]", concatInputs.String(), n,
		))
	} else {
		filters = append(filters, fmt.Sprintf(
			"%sconcat=n=%d:v=1:a=0[outv]", concatInputs.String(), n,
		))
	}

	args = append(args, "-filter_complex", strings.Join(filters, "; "))
	args = append(args, "-map", "[outv]")
	if withAudio {
		args = append(args, "-map", "[outa]", "-c:a", "aac", "-b:a", "192k")
	}
	args = append(args,
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "fast",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)
	return args
}
