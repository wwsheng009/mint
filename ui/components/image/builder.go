package image

import (
	stdimage "image"

	"github.com/wwsheng009/mint/runtime/style"
)

// Builder provides a fluent API for creating Image VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Image builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) ID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Alt(alt string) *Builder {
	b.node.SetAlt(alt)
	return b
}

// SourceDataURI sets a PNG, JPEG or GIF data URI as the image source.
// SVG data URIs are accepted as metadata but render as the fallback text until
// callers provide a raster payload.
func (b *Builder) SourceDataURI(dataURI string) *Builder {
	b.node.SetDataURI(dataURI)
	return b
}

func (b *Builder) SourceImage(img stdimage.Image) *Builder {
	b.node.SetSourceImage(img)
	return b
}

func (b *Builder) SourceRGBA(rgba []byte, width, height int) *Builder {
	b.node.SetRGBA(rgba, width, height)
	return b
}

func (b *Builder) Size(widthCells, heightCells int) *Builder {
	b.node.SetSize(widthCells, heightCells)
	return b
}

func (b *Builder) Width(widthCells int) *Builder {
	b.node.SetSize(widthCells, b.node.heightCells)
	return b
}

func (b *Builder) Height(heightCells int) *Builder {
	b.node.SetSize(b.node.widthCells, heightCells)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) Build() *VNode {
	return b.node
}
