package drawer

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Builder — Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Drawer VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Drawer builder.
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

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.vnode.SetID(id)
	return b
}

// Placement sets the drawer placement.
func (b *Builder) Placement(p Placement) *Builder {
	b.vnode.SetPlacement(p)
	return b
}

// Title sets the drawer title.
func (b *Builder) Title(title string) *Builder {
	b.vnode.SetTitle(title)
	return b
}

// Content sets the drawer content.
func (b *Builder) Content(content rtui.VNode) *Builder {
	b.vnode.SetContent(content)
	return b
}

// Footer sets the drawer footer.
func (b *Builder) Footer(footer rtui.VNode) *Builder {
	b.vnode.SetFooter(footer)
	return b
}

// Open sets whether the drawer is open.
func (b *Builder) Open(open bool) *Builder {
	b.vnode.SetOpen(open)
	return b
}

// Width sets the drawer width (for Left/Right placement).
func (b *Builder) Width(width int) *Builder {
	b.vnode.SetWidth(width)
	return b
}

// Height sets the drawer height (for Top/Bottom placement).
func (b *Builder) Height(height int) *Builder {
	b.vnode.SetHeight(height)
	return b
}

// Padding sets the inner padding.
func (b *Builder) Padding(padding int) *Builder {
	b.vnode.SetPadding(padding)
	return b
}

// Closeable sets whether the drawer can be closed.
func (b *Builder) Closeable(closeable bool) *Builder {
	b.vnode.SetCloseable(closeable)
	return b
}

// CloseOnEsc controls whether ESC closes the drawer.
func (b *Builder) CloseOnEsc(closeOnEsc bool) *Builder {
	b.vnode.SetCloseOnEsc(closeOnEsc)
	return b
}

// CloseOnBackdrop controls whether clicking outside closes the drawer.
func (b *Builder) CloseOnBackdrop(closeOnBackdrop bool) *Builder {
	b.vnode.SetCloseOnBackdrop(closeOnBackdrop)
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

// Shadow controls whether the drawer renders a shadow.
func (b *Builder) Shadow(show bool) *Builder {
	b.vnode.SetShadow(show)
	return b
}

// ShadowStyle sets the shadow style.
func (b *Builder) ShadowStyle(s style.Style) *Builder {
	b.vnode.SetShadowStyle(s)
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

// Border style convenience methods.
func (b *Builder) Single() *Builder  { b.vnode.Single(); return b }
func (b *Builder) Double() *Builder  { b.vnode.Double(); return b }
func (b *Builder) Rounded() *Builder { b.vnode.Rounded(); return b }
func (b *Builder) Dashed() *Builder  { b.vnode.Dashed(); return b }

// Placement convenience methods.
func (b *Builder) Right() *Builder  { b.vnode.Right(); return b }
func (b *Builder) Left() *Builder   { b.vnode.Left(); return b }
func (b *Builder) Top() *Builder    { b.vnode.Top(); return b }
func (b *Builder) Bottom() *Builder { b.vnode.Bottom(); return b }

// Open state convenience methods.
func (b *Builder) Opened() *Builder   { return b.Open(true) }
func (b *Builder) Closed() *Builder   { return b.Open(false) }
func (b *Builder) NoShadow() *Builder { return b.Shadow(false) }

// Build returns the configured VNode as rtui.VNode.
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

// Of creates a Drawer with the given content (right placement by default).
func Of(content rtui.VNode) rtui.VNode {
	return NewBuilder().Content(content).Build()
}

// Titled creates a Drawer with a title and content.
func Titled(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Build()
}

// FromRight creates a right-side Drawer with title and content.
func FromRight(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Right().Build()
}

// FromLeft creates a left-side Drawer with title and content.
func FromLeft(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Left().Build()
}

// FromTop creates a top Drawer with title and content.
func FromTop(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Top().Build()
}

// FromBottom creates a bottom Drawer with title and content.
func FromBottom(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Bottom().Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Drawer creates a new Drawer builder.
func Drawer() *Builder {
	return NewBuilder()
}

// WithTitle creates a titled Drawer builder.
func WithTitle(title string) *Builder {
	return NewBuilder().Title(title)
}

// Simple creates a simple text drawer.
func Simple(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		Content(newtext.New(message)).
		Width(30).
		Build()
}
