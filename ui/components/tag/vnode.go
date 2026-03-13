package tag

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

const (
	propKey          = "key"
	propText         = "text"
	propColor        = "color"
	propClosable     = "closable"
	propCloseIntent  = "closeIntent"
	propIcon         = "icon"
	propStyle        = "style"
)

// =============================================================================
// Tag Color
// =============================================================================

// TagColor defines the color variant of the tag.
type TagColor int

const (
	ColorDefault    TagColor = iota // default gray/surface
	ColorPrimary                    // blue
	ColorSuccess                    // green
	ColorWarning                    // yellow
	ColorError                      // red
	ColorProcessing                 // cyan
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the immutable description of a Tag component.
type VNode struct {
	*rtui.ElementVNode

	key         string
	text        string
	color       TagColor
	closable    bool
	closeIntent interface{}
	icon        string
	tagStyle    style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Tag VNode with the given text.
func New(text string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("tag"),
		text:         text,
		color:        ColorDefault,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (v *VNode) Key() string                                   { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                  { v.key = key; return v }
func (v *VNode) Tag() string                                   { return "tag" }
func (v *VNode) Style() style.Style                            { return v.tagStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode             { v.tagStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                        { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode  { return v }
func (v *VNode) GetLayer() rtui.Layer                          { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode              { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:         v.key,
		propText:        v.text,
		propColor:       v.color,
		propClosable:    v.closable,
		propCloseIntent: v.closeIntent,
		propIcon:        v.icon,
		propStyle:       v.tagStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if s, ok := props[propKey].(string); ok {
		v.key = s
	}
	if s, ok := props[propText].(string); ok {
		v.text = s
	}
	if c, ok := props[propColor].(TagColor); ok {
		v.color = c
	}
	if b, ok := props[propClosable].(bool); ok {
		v.closable = b
	}
	if ci, ok := props[propCloseIntent]; ok {
		v.closeIntent = ci
	}
	if s, ok := props[propIcon].(string); ok {
		v.icon = s
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.tagStyle = s
	}
	return v
}

// =============================================================================
// InstanceFactory
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (v *VNode) SetText(text string) *VNode              { v.text = text; return v }
func (v *VNode) SetColor(c TagColor) *VNode              { v.color = c; return v }
func (v *VNode) SetClosable(closable bool) *VNode        { v.closable = closable; return v }
func (v *VNode) SetCloseIntent(ci interface{}) *VNode    { v.closeIntent = ci; return v }
func (v *VNode) SetIcon(icon string) *VNode              { v.icon = icon; return v }
func (v *VNode) SetTagStyle(s style.Style) *VNode        { v.tagStyle = s; return v }

func (v *VNode) Default() *VNode    { v.color = ColorDefault; return v }
func (v *VNode) Primary() *VNode    { v.color = ColorPrimary; return v }
func (v *VNode) Success() *VNode    { v.color = ColorSuccess; return v }
func (v *VNode) Warning() *VNode    { v.color = ColorWarning; return v }
func (v *VNode) Error() *VNode      { v.color = ColorError; return v }
func (v *VNode) Processing() *VNode { v.color = ColorProcessing; return v }

// =============================================================================
// Props Accessors
// =============================================================================

func (v *VNode) Text() string           { return v.text }
func (v *VNode) Color() TagColor        { return v.color }
func (v *VNode) Closable() bool         { return v.closable }
func (v *VNode) CloseIntent() interface{} { return v.closeIntent }
func (v *VNode) Icon() string           { return v.icon }
