package ffprobe

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type Meta struct {
	Duration  float64
	Width     int
	Height    int
	Codec     string
	FrameRate float64
}

type probeOutput struct {
	Streams []struct {
		CodecName  string `json:"codec_name"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func Probe(path string) (Meta, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return Meta{}, fmt.Errorf("ffprobe: %w", err)
	}

	var probe probeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return Meta{}, fmt.Errorf("ffprobe parse: %w", err)
	}

	meta := Meta{}

	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		meta.Duration = d
	}

	for _, s := range probe.Streams {
		if s.Width > 0 {
			meta.Width = s.Width
			meta.Height = s.Height
			meta.Codec = s.CodecName
			meta.FrameRate = parseFrameRate(s.RFrameRate)
			break
		}
	}

	return meta, nil
}

func parseFrameRate(s string) float64 {
	var num, den int
	if _, err := fmt.Sscanf(s, "%d/%d", &num, &den); err != nil || den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}
