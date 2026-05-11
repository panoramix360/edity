package clip

import "github.com/panoramix360/edity/internal/ffprobe"

type Clip struct {
	Path     string
	Meta     ffprobe.Meta
	InPoint  float64
	OutPoint float64
}

func (c Clip) Duration() float64 {
	return c.OutPoint - c.InPoint
}
