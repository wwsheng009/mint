// Package modal provides Fiber-first Modal dialog component.
package modal

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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
	title     string
	isOpen    bool
	centered  bool
	closeable bool

	// === Layout Props ===
	width  int
	height int

	// === Content ===
	content rtui.VNode
	footer  rtui.VNode

	// === Style ===
	modalStyle style.Style
	borderStyle string

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
		ElementVNode: rtui.NewElement("modal"),
		isOpen:       false,
		centered:     true,
		closeable:    true,
		width:        40,
		height:       15,
		borderStyle:  "double", // ✨ Modal 默认使用双线边框
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string           { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string           { return "modal" }
func (v *VNode) Type() rtui.VNodeType  { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	var children []rtui.VNode
	if v.content != nil {
		children = append(children, v.content)
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

func (v *VNode) GetLayer() rtui.Layer   { return rtui.LayerModal }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style    { return v.modalStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.modalStyle = s; return v }

func (v *VNode) TextContent() string   { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         v.key,
		"title":       v.title,
		"label":       v.title, // ✨ 映射 title 到 label（边框标签）
		"isOpen":      v.isOpen,
		"centered":    v.centered,
		"closeable":   v.closeable,
		"width":       v.width,
		"height":      v.height,
		"content":     v.content,
		"footer":      v.footer,
		"modalStyle":  v.modalStyle,
		"borderStyle": v.borderStyle,
		"closeIntent": v.closeIntent,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["title"].(string); ok {
		v.title = val
	}
	// ✨ 从 "label" 属性读取标题
	if val, ok := p["label"].(string); ok {
		v.title = val
	}
	if val, ok := p["isOpen"].(bool); ok {
		v.isOpen = val
	}
	if val, ok := p["centered"].(bool); ok {
		v.centered = val
	}
	if val, ok := p["closeable"].(bool); ok {
		v.closeable = val
	}
	if val, ok := p["width"].(int); ok {
		v.width = val
	}
	if val, ok := p["height"].(int); ok {
		v.height = val
	}
	if val, ok := p["content"].(rtui.VNode); ok {
		v.content = val
	}
	if val, ok := p["footer"].(rtui.VNode); ok {
		v.footer = val
	}
	if val, ok := p["modalStyle"].(style.Style); ok {
		v.modalStyle = val
	}
	if val, ok := p["borderStyle"].(string); ok {
		v.borderStyle = val
	}
	if val, ok := p["closeIntent"].(intent.Intent); ok {
		v.closeIntent = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":         v.key,
		"title":       v.title,
		"label":       v.title, // ✨ 映射到 label（边框标签）
		"isOpen":      v.isOpen,
		"centered":    v.centered,
		"closeable":   v.closeable,
		"width":       v.width,
		"height":      v.height,
		"content":     v.content,
		"footer":      v.footer,
		"modalStyle":  v.modalStyle,
		"borderStyle": v.borderStyle,
		"closeIntent": v.closeIntent,
	})
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetTitle(title string) *VNode   { v.title = title; return v }
func (v *VNode) SetContent(content rtui.VNode) *VNode { v.content = content; return v }
func (v *VNode) SetFooter(footer rtui.VNode) *VNode   { v.footer = footer; return v }
func (v *VNode) SetOpen(isOpen bool) *VNode          { v.isOpen = isOpen; return v }
func (v *VNode) SetWidth(width int) *VNode           { v.width = width; return v }
func (v *VNode) SetHeight(height int) *VNode         { v.height = height; return v }
func (v *VNode) SetCentered(centered bool) *VNode    { v.centered = centered; return v }
func (v *VNode) SetCloseable(closeable bool) *VNode  { v.closeable = closeable; return v }
func (v *VNode) SetBorderStyle(style string) *VNode  { v.borderStyle = style; return v }
func (v *VNode) SetIntent(i intent.Intent) *VNode    { v.closeIntent = i; return v }

func (v *VNode) Open() *VNode    { return v.SetOpen(true) }
func (v *VNode) Close() *VNode   { return v.SetOpen(false) }
func (v *VNode) Toggle() *VNode  { v.isOpen = !v.isOpen; return v }

func (v *VNode) Size(w, h int) *VNode { return v.SetWidth(w).SetHeight(h) }

func (v *VNode) Single() *VNode   { return v.SetBorderStyle("single") }
func (v *VNode) Double() *VNode   { return v.SetBorderStyle("double") }
func (v *VNode) Rounded() *VNode  { return v.SetBorderStyle("rounded") }
func (v *VNode) Dashed() *VNode   { return v.SetBorderStyle("dashed") }

func (v *VNode) Center() *VNode  { return v.SetCentered(true) }

func (v *VNode) OnClose(i intent.Intent) *VNode { return v.SetIntent(i) }

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Title() string      { return v.title }
func (v *VNode) Content() rtui.VNode { return v.content }
func (v *VNode) Footer() rtui.VNode  { return v.footer }
func (v *VNode) IsOpen() bool        { return v.isOpen }
func (v *VNode) Width() int          { return v.width }
func (v *VNode) Height() int         { return v.height }
func (v *VNode) Centered() bool      { return v.centered }
func (v *VNode) Closeable() bool     { return v.closeable }
func (v *VNode) BorderStyle() string { return v.borderStyle }
