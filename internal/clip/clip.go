package clip

import "github.com/panoramix360/edity/internal/ffprobe"

type Clip struct {
	Path string
	Meta ffprobe.Meta
}
