package cursor

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides the fluent API for standalone cursors.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new cursor builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) Config(cfg Config) *Builder {
	b.node.SetConfig(cfg)
	return b
}

func (b *Builder) Shape(shape Shape) *Builder {
	b.node.SetShape(shape)
	return b
}

func (b *Builder) Theme(theme ThemeRole) *Builder {
	b.node.SetTheme(theme)
	return b
}

func (b *Builder) Glyph(glyph string) *Builder {
	b.node.SetGlyph(glyph)
	return b
}

func (b *Builder) Visible(visible bool) *Builder {
	b.node.SetVisible(visible)
	return b
}

func (b *Builder) Blink(enabled bool) *Builder {
	b.node.SetBlink(enabled)
	return b
}

func (b *Builder) BlinkInterval(interval time.Duration) *Builder {
	b.node.SetBlinkInterval(interval)
	return b
}

func (b *Builder) FastBlink() *Builder {
	return b.BlinkInterval(FastBlinkInterval)
}

func (b *Builder) NormalBlink() *Builder {
	return b.BlinkInterval(NormalBlinkInterval)
}

func (b *Builder) SlowBlink() *Builder {
	return b.BlinkInterval(SlowBlinkInterval)
}

func (b *Builder) Steady() *Builder {
	return b.Blink(false)
}

func (b *Builder) Block() *Builder {
	return b.Shape(ShapeBlock)
}

func (b *Builder) Underline() *Builder {
	return b.Shape(ShapeUnderline)
}

func (b *Builder) Bar() *Builder {
	return b.Shape(ShapeBar)
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
