package scroll

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

type VerticalScrollbarConfig struct {
	Rail                 string
	Thumb                string
	HideWhenUnscrollable bool
}

func DefaultVerticalScrollbarConfig() VerticalScrollbarConfig {
	return VerticalScrollbarConfig{
		Rail:                 "│",
		Thumb:                "█",
		HideWhenUnscrollable: true,
	}
}

func DrawVerticalScrollbar(
	x, y, height int,
	viewport VerticalViewport,
	scrollbarStyle style.Style,
	config VerticalScrollbarConfig,
) []paint.DrawCmd {
	if height <= 0 {
		return nil
	}

	if config.Rail == "" {
		config.Rail = "│"
	}
	if config.Thumb == "" {
		config.Thumb = "█"
	}

	viewport.Normalize()
	if !viewport.IsScrollable() && config.HideWhenUnscrollable {
		return nil
	}

	thumbStart := 0
	thumbLen := height

	if viewport.IsScrollable() {
		thumbLen = (height * viewport.ViewSize) / viewport.ContentSize
		if thumbLen < 1 {
			thumbLen = 1
		}
		if thumbLen > height {
			thumbLen = height
		}

		trackSpan := height - thumbLen
		maxOffset := viewport.MaxOffset()
		if trackSpan > 0 && maxOffset > 0 {
			thumbStart = (viewport.Offset * trackSpan) / maxOffset
		}
	}

	cmds := make([]paint.DrawCmd, 0, height)
	for row := 0; row < height; row++ {
		text := config.Rail
		if row >= thumbStart && row < thumbStart+thumbLen {
			text = config.Thumb
		}
		cmds = append(cmds, paint.NewTextCmd(x, y+row, text, scrollbarStyle))
	}
	return cmds
}
