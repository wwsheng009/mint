package modal

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Modal VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Modal builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.vnode.SetKey(key)
	return b
}

// Title sets the modal title.
func (b *Builder) Title(title string) *Builder {
	b.vnode.SetTitle(title)
	return b
}

// Content sets the modal content.
func (b *Builder) Content(content rtui.VNode) *Builder {
	b.vnode.SetContent(content)
	return b
}

// Footer sets the modal footer.
func (b *Builder) Footer(footer rtui.VNode) *Builder {
	b.vnode.SetFooter(footer)
	return b
}

// Open sets the modal as open.
func (b *Builder) Open(open bool) *Builder {
	b.vnode.SetOpen(open)
	return b
}

// Width sets the modal width.
func (b *Builder) Width(width int) *Builder {
	b.vnode.SetWidth(width)
	return b
}

// Height sets the modal height.
func (b *Builder) Height(height int) *Builder {
	b.vnode.SetHeight(height)
	return b
}

// Size sets both width and height.
func (b *Builder) Size(w, h int) *Builder {
	b.vnode.Size(w, h)
	return b
}

// Centered sets whether the modal is centered.
func (b *Builder) Centered(centered bool) *Builder {
	b.vnode.SetCentered(centered)
	return b
}

// Center sets the modal as centered (convenience method).
func (b *Builder) Center() *Builder {
	b.vnode.Center()
	return b
}

// Closeable sets whether the modal can be closed.
func (b *Builder) Closeable(closeable bool) *Builder {
	b.vnode.SetCloseable(closeable)
	return b
}

// BorderStyle sets the border style.
func (b *Builder) BorderStyle(style string) *Builder {
	b.vnode.SetBorderStyle(style)
	return b
}

// OnClose sets the close intent.
func (b *Builder) OnClose(intent intent.Intent) *Builder {
	b.vnode.OnClose(intent)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c style.Color) *Builder {
	s := b.vnode.Style()
	s.FG = c
	b.vnode.SetStyle(s)
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c style.Color) *Builder {
	s := b.vnode.Style()
	s.BG = c
	b.vnode.SetStyle(s)
	return b
}

// Border style convenience methods
func (b *Builder) Single() *Builder   { b.vnode.Single(); return b }
func (b *Builder) Double() *Builder   { b.vnode.Double(); return b }
func (b *Builder) Rounded() *Builder  { b.vnode.Rounded(); return b }
func (b *Builder) Dashed() *Builder   { b.vnode.Dashed(); return b }

// Open state convenience methods
func (b *Builder) Opened() *Builder   { return b.Open(true) }
func (b *Builder) Closed() *Builder   { return b.Open(false) }

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.vnode
}

// BuildVNode returns the configured *VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.vnode
}

// BuildInstance directly creates an Instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.vnode.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a Modal with the given content.
func Of(content rtui.VNode) rtui.VNode {
	return NewBuilder().Content(content).Build()
}

// OfSize creates a Modal with explicit dimensions.
func OfSize(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Content(content).Width(width).Height(height).Build()
}

// Titled creates a Modal with a title and content.
func Titled(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Rounded().Build()
}

// WithHeader creates a full-featured Modal with header (title), content, and footer.
func WithHeader(title, content, footer rtui.VNode) rtui.VNode {
	return NewBuilder().Title("").Content(content).Footer(footer).Rounded().Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Modal creates a new Modal builder.
func Modal() *Builder {
	return NewBuilder()
}

// WithTitle creates a titled Modal builder.
func WithTitle(title string) *Builder {
	return NewBuilder().Title(title).Rounded()
}

// Alert creates an alert-style modal.
func Alert(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		Content(newtext.New(message)).
		Width(40).
		Height(10).
		Rounded().
		Build()
}

// Confirm creates a confirm-style modal.
func Confirm(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		Content(newtext.New(message)).
		Width(40).
		Height(10).
		Rounded().
		Build()
}
