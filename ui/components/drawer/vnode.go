// Package drawer provides a side-panel Drawer component that slides in from a screen edge.
package drawer

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Placement Type
// =============================================================================

// Placement controls which edge the drawer slides in from.
type Placement int

const (
	PlacementRight  Placement = iota // Default — slides in from the right
	PlacementLeft                    // Slides in from the left
	PlacementTop                     // Slides in from the top
	PlacementBottom                  // Slides in from the bottom
)

// =============================================================================
// Prop Key Constants
// =============================================================================

const (
	propBorderStyle     = "borderStyle"
	propCloseable       = "closeable"
	propCloseIntent     = "closeIntent"
	propCloseOnBackdrop = "closeOnBackdrop"
	propCloseOnEsc      = "closeOnEsc"
	propContent         = "content"
	propDashed          = "dashed"
	propDouble          = "double"
	propDrawerStyle     = "drawerStyle"
	propFooter          = "footer"
	propHeight          = "height"
	propIsOpen          = "isOpen"
	propKey             = "key"
	propPadding         = "padding"
	propPlacement       = "placement"
	propRounded         = "rounded"
	propShadowStyle     = "shadowStyle"
	propShowShadow      = "showShadow"
	propSingle          = "single"
	propTitle           = "title"
	propWidth           = "width"
)

// =============================================================================
// VNode — Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the drawer description.
// Contains ONLY declarative information — no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Drawer Props ===
	placement       Placement
	title           string
	isOpen          bool
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool

	// === Layout Props ===
	width   int // used when placement is Left/Right
	height  int // used when placement is Top/Bottom
	padding int

	// === Content ===
	content rtui.VNode
	footer  rtui.VNode

	// === Style ===
	drawerStyle style.Style
	shadowStyle style.Style
	borderStyle string
	showShadow  bool

	// === Intent (No Closures!) ===
	closeIntent intent.Intent
}

// Compile-time interface checks.
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Drawer VNode with default values.
func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("drawer"),
		placement:       PlacementRight,
		isOpen:          false,
		closeable:       true,
		closeOnEsc:      true,
		closeOnBackdrop: true,
		width:           30,
		height:          15,
		padding:         0,
		borderStyle:     "single",
		showShadow:      true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string                  { return "drawer" }
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

func (v *VNode) Style() style.Style                { return v.drawerStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.drawerStyle = s; return v }

func (v *VNode) TextContent() string { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:             v.key,
		propPlacement:       int(v.placement),
		propTitle:           v.title,
		propIsOpen:          v.isOpen,
		propCloseable:       v.closeable,
		propCloseOnEsc:      v.closeOnEsc,
		propCloseOnBackdrop: v.closeOnBackdrop,
		propWidth:           v.width,
		propHeight:          v.height,
		propPadding:         v.padding,
		propContent:         v.content,
		propFooter:          v.footer,
		propDrawerStyle:     v.drawerStyle,
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
	if val, ok := p[propPlacement].(int); ok {
		v.placement = Placement(val)
	}
	if val, ok := p[propTitle].(string); ok {
		v.title = val
	}
	if val, ok := p[propIsOpen].(bool); ok {
		v.isOpen = val
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
	if val, ok := p[propDrawerStyle].(style.Style); ok {
		v.drawerStyle = val
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
	return NewInstance(v.Props())
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetPlacement(p Placement) *VNode       { v.placement = p; return v }
func (v *VNode) SetTitle(title string) *VNode           { v.title = title; return v }
func (v *VNode) SetContent(content rtui.VNode) *VNode  { v.content = content; return v }
func (v *VNode) SetFooter(footer rtui.VNode) *VNode    { v.footer = footer; return v }
func (v *VNode) SetOpen(isOpen bool) *VNode            { v.isOpen = isOpen; return v }
func (v *VNode) SetWidth(width int) *VNode             { v.width = width; return v }
func (v *VNode) SetHeight(height int) *VNode           { v.height = height; return v }
func (v *VNode) SetPadding(padding int) *VNode         { v.padding = padding; return v }
func (v *VNode) SetCloseable(closeable bool) *VNode    { v.closeable = closeable; return v }
func (v *VNode) SetCloseOnEsc(v2 bool) *VNode          { v.closeOnEsc = v2; return v }
func (v *VNode) SetCloseOnBackdrop(v2 bool) *VNode     { v.closeOnBackdrop = v2; return v }
func (v *VNode) SetBorderStyle(s string) *VNode        { v.borderStyle = s; return v }
func (v *VNode) SetIntent(i intent.Intent) *VNode      { v.closeIntent = i; return v }
func (v *VNode) SetShadow(show bool) *VNode            { v.showShadow = show; return v }
func (v *VNode) SetShadowStyle(s style.Style) *VNode   { v.shadowStyle = s; return v }

func (v *VNode) Open() *VNode   { return v.SetOpen(true) }
func (v *VNode) Close() *VNode  { return v.SetOpen(false) }
func (v *VNode) Toggle() *VNode { v.isOpen = !v.isOpen; return v }

func (v *VNode) Right() *VNode  { return v.SetPlacement(PlacementRight) }
func (v *VNode) Left() *VNode   { return v.SetPlacement(PlacementLeft) }
func (v *VNode) Top() *VNode    { return v.SetPlacement(PlacementTop) }
func (v *VNode) Bottom() *VNode { return v.SetPlacement(PlacementBottom) }

func (v *VNode) Single() *VNode  { return v.SetBorderStyle("single") }
func (v *VNode) Double() *VNode  { return v.SetBorderStyle("double") }
func (v *VNode) Rounded() *VNode { return v.SetBorderStyle("rounded") }
func (v *VNode) Dashed() *VNode  { return v.SetBorderStyle("dashed") }

func (v *VNode) OnClose(i intent.Intent) *VNode { return v.SetIntent(i) }

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Placement() Placement   { return v.placement }
func (v *VNode) Title() string          { return v.title }
func (v *VNode) Content() rtui.VNode    { return v.content }
func (v *VNode) Footer() rtui.VNode     { return v.footer }
func (v *VNode) IsOpen() bool           { return v.isOpen }
func (v *VNode) Width() int             { return v.width }
func (v *VNode) Height() int            { return v.height }
func (v *VNode) Padding() int           { return v.padding }
func (v *VNode) Closeable() bool        { return v.closeable }
func (v *VNode) CloseOnEsc() bool       { return v.closeOnEsc }
func (v *VNode) CloseOnBackdrop() bool  { return v.closeOnBackdrop }
func (v *VNode) BorderStyle() string    { return v.borderStyle }
func (v *VNode) Shadow() bool           { return v.showShadow }

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Drawer VNode.
func (v *VNode) GetBoxModel() layout.BoxModel {
	boxModel := layout.BoxModel{}

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
