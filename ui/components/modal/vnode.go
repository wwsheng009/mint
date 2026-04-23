// Package modal provides Fiber-first Modal dialog component.
package modal

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propBorderStyle = "borderStyle"
	propCentered = "centered"
	propCloseIntent = "closeIntent"
	propCloseOnBackdrop = "closeOnBackdrop"
	propCloseOnEsc = "closeOnEsc"
	propCloseable = "closeable"
	propContent = "content"
	propDashed = "dashed"
	propDouble = "double"
	propFooter = "footer"
	propHeight = "height"
	propIsOpen = "isOpen"
	propKey = "key"
	propLabel = "label"
	propModalStyle = "modalStyle"
	propPadding = "padding"
	propRounded = "rounded"
	propShadowStyle = "shadowStyle"
	propShowShadow = "showShadow"
	propSingle = "single"
	propTitle = "title"
	propWidth = "width"
)

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the modal dialog description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Modal Props ===
	title           string
	isOpen          bool
	centered        bool
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool

	// === Layout Props ===
	width   int
	height  int
	padding int

	// === Content ===
	content rtui.VNode
	footer  rtui.VNode

	// === Style ===
	modalStyle  style.Style
	shadowStyle style.Style
	borderStyle string
	showShadow  bool

	// === Intent (No Closures!) ===
	closeIntent intent.Intent
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new modal VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("modal"),
		isOpen:          false,
		centered:        true,
		closeable:       true,
		closeOnEsc:      true,
		closeOnBackdrop: true,
		width:           40,
		height:          15,
		padding:         0,
		borderStyle:     "double", // ✨ Modal 默认使用双线边框
		showShadow:      true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string                  { return "modal" }
func (v *VNode) Type() rtui.VNodeType         { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	var children []rtui.VNode
	if v.content != nil {
		if v.footer != nil {
			children = append(children, rtui.VStackBuilder(v.content).SetGap(0).SetFlex(1).Build())
		} else {
			children = append(children, v.content)
		}
	}
	if v.footer != nil {
		children = append(children, v.footer)
	}
	return children
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.content = children[0]
	}
	if len(children) > 1 {
		v.footer = children[1]
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer             { return rtui.LayerModal }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style                { return v.modalStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.modalStyle = s; return v }

func (v *VNode) TextContent() string { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:             v.key,
		propTitle:           v.title,
		propLabel:           v.title, // ✨ 映射 title 到 label（边框标签）
		propIsOpen:          v.isOpen,
		propCentered:        v.centered,
		propCloseable:       v.closeable,
		propCloseOnEsc:      v.closeOnEsc,
		propCloseOnBackdrop: v.closeOnBackdrop,
		propWidth:           v.width,
		propHeight:          v.height,
		propPadding:         v.padding,
		propContent:         v.content,
		propFooter:          v.footer,
		propModalStyle:      v.modalStyle,
		propShadowStyle:     v.shadowStyle,
		propBorderStyle:     v.borderStyle,
		propShowShadow:      v.showShadow,
		propCloseIntent:     v.closeIntent,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p[propKey].(string); ok {
		v.key = val
	}
	if val, ok := p[propTitle].(string); ok {
		v.title = val
	}
	// ✨ 从 "label" 属性读取标题
	if val, ok := p[propLabel].(string); ok {
		v.title = val
	}
	if val, ok := p[propIsOpen].(bool); ok {
		v.isOpen = val
	}
	if val, ok := p[propCentered].(bool); ok {
		v.centered = val
	}
	if val, ok := p[propCloseable].(bool); ok {
		v.closeable = val
	}
	if val, ok := p[propCloseOnEsc].(bool); ok {
		v.closeOnEsc = val
	}
	if val, ok := p[propCloseOnBackdrop].(bool); ok {
		v.closeOnBackdrop = val
	}
	if val, ok := p[propWidth].(int); ok {
		v.width = val
	}
	if val, ok := p[propHeight].(int); ok {
		v.height = val
	}
	if val, ok := p[propPadding].(int); ok {
		v.padding = val
	}
	if val, ok := p[propContent].(rtui.VNode); ok {
		v.content = val
	}
	if val, ok := p[propFooter].(rtui.VNode); ok {
		v.footer = val
	}
	if val, ok := p[propModalStyle].(style.Style); ok {
		v.modalStyle = val
	}
	if val, ok := p[propShadowStyle].(style.Style); ok {
		v.shadowStyle = val
	}
	if val, ok := p[propBorderStyle].(string); ok {
		v.borderStyle = val
	}
	if val, ok := p[propShowShadow].(bool); ok {
		v.showShadow = val
	}
	if val, ok := p[propCloseIntent].(intent.Intent); ok {
		v.closeIntent = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		propKey:             v.key,
		propTitle:           v.title,
		propLabel:           v.title, // ✨ 映射到 label（边框标签）
		propIsOpen:          v.isOpen,
		propCentered:        v.centered,
		propCloseable:       v.closeable,
		propCloseOnEsc:      v.closeOnEsc,
		propCloseOnBackdrop: v.closeOnBackdrop,
		propWidth:           v.width,
		propHeight:          v.height,
		propPadding:         v.padding,
		propContent:         v.content,
		propFooter:          v.footer,
		propModalStyle:      v.modalStyle,
		propShadowStyle:     v.shadowStyle,
		propBorderStyle:     v.borderStyle,
		propShowShadow:      v.showShadow,
		propCloseIntent:     v.closeIntent,
	})
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetTitle(title string) *VNode         { v.title = title; return v }
func (v *VNode) SetContent(content rtui.VNode) *VNode { v.content = content; return v }
func (v *VNode) SetFooter(footer rtui.VNode) *VNode   { v.footer = footer; return v }
func (v *VNode) SetOpen(isOpen bool) *VNode           { v.isOpen = isOpen; return v }
func (v *VNode) SetWidth(width int) *VNode            { v.width = width; return v }
func (v *VNode) SetHeight(height int) *VNode          { v.height = height; return v }
func (v *VNode) SetPadding(padding int) *VNode        { v.padding = padding; return v }
func (v *VNode) SetCentered(centered bool) *VNode     { v.centered = centered; return v }
func (v *VNode) SetCloseable(closeable bool) *VNode   { v.closeable = closeable; return v }
func (v *VNode) SetCloseOnEsc(closeOnEsc bool) *VNode { v.closeOnEsc = closeOnEsc; return v }
func (v *VNode) SetCloseOnBackdrop(closeOnBackdrop bool) *VNode {
	v.closeOnBackdrop = closeOnBackdrop
	return v
}
func (v *VNode) SetBorderStyle(style string) *VNode  { v.borderStyle = style; return v }
func (v *VNode) SetIntent(i intent.Intent) *VNode    { v.closeIntent = i; return v }
func (v *VNode) SetShadow(show bool) *VNode          { v.showShadow = show; return v }
func (v *VNode) SetShadowStyle(s style.Style) *VNode { v.shadowStyle = s; return v }

func (v *VNode) Open() *VNode   { return v.SetOpen(true) }
func (v *VNode) Close() *VNode  { return v.SetOpen(false) }
func (v *VNode) Toggle() *VNode { v.isOpen = !v.isOpen; return v }

func (v *VNode) Size(w, h int) *VNode { return v.SetWidth(w).SetHeight(h) }
func (v *VNode) InnerSize(w, h int) *VNode {
	extraHeight := 2
	if v.title != "" {
		extraHeight += 2
	}
	return v.SetWidth(w + 2 + 2*v.padding).SetHeight(h + extraHeight + 2*v.padding)
}

func (v *VNode) Single() *VNode  { return v.SetBorderStyle("single") }
func (v *VNode) Double() *VNode  { return v.SetBorderStyle("double") }
func (v *VNode) Rounded() *VNode { return v.SetBorderStyle("rounded") }
func (v *VNode) Dashed() *VNode  { return v.SetBorderStyle("dashed") }

func (v *VNode) Center() *VNode { return v.SetCentered(true) }

func (v *VNode) OnClose(i intent.Intent) *VNode { return v.SetIntent(i) }

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Title() string         { return v.title }
func (v *VNode) Content() rtui.VNode   { return v.content }
func (v *VNode) Footer() rtui.VNode    { return v.footer }
func (v *VNode) IsOpen() bool          { return v.isOpen }
func (v *VNode) Width() int            { return v.width }
func (v *VNode) Height() int           { return v.height }
func (v *VNode) Padding() int          { return v.padding }
func (v *VNode) Centered() bool        { return v.centered }
func (v *VNode) Closeable() bool       { return v.closeable }
func (v *VNode) CloseOnEsc() bool      { return v.closeOnEsc }
func (v *VNode) CloseOnBackdrop() bool { return v.closeOnBackdrop }
func (v *VNode) BorderStyle() string   { return v.borderStyle }
func (v *VNode) Shadow() bool          { return v.showShadow }

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Modal VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Modal uses borderStyle property, title is displayed in border label.
func (v *VNode) GetBoxModel() layout.BoxModel {
	boxModel := layout.BoxModel{}

	// Border from modal properties
	var borderStyle layout.BorderStyle
	switch v.borderStyle {
	case propDouble:
		borderStyle = layout.BorderDouble
	case propRounded:
		borderStyle = layout.BorderRounded
	case propDashed:
		borderStyle = layout.BorderDashed
	case propSingle:
		borderStyle = layout.BorderSingle
	default:
		borderStyle = layout.BorderNone
	}

	boxModel.Border = layout.NewBorder(borderStyle)
	boxModel.Padding = layout.Padding{
		Top:    v.padding,
		Right:  v.padding,
		Bottom: v.padding,
		Left:   v.padding,
	}
	if v.title != "" {
		boxModel.Padding.Top += 2
	}

	return boxModel
}
