// Package ui provides a declarative UI framework for terminal applications.
package ui

import "github.com/wwsheng009/mint/runtime/style"

// VNode is the virtual node interface - the core of the declarative UI system.
// It represents a lightweight description of what should be rendered.
type VNode interface {
	// Type returns the node type for diff algorithm
	Type() VNodeType

	// Props returns the node properties
	Props() Props
	SetProps(p Props)

	// Children returns the child nodes
	Children() []VNode
	SetChildren(children []VNode)

	// Key returns the key for list diffing
	Key() string
	SetKey(key string)

	// Style returns the visual style
	Style() style.Style
	SetStyle(s style.Style)
}

// VNodeType represents the type of VNode
type VNodeType int

const (
	// VNodeElement is a standard element node (div, span, etc.)
	VNodeElement VNodeType = iota

	// VNodeText is a text node with content
	VNodeText

	// VNodeComponent is a function component
	VNodeComponent

	// VNodeFragment is a fragment that doesn't add extra DOM nodes
	VNodeFragment
)

// String returns the string representation of VNodeType
func (t VNodeType) String() string {
	switch t {
	case VNodeElement:
		return "Element"
	case VNodeText:
		return "Text"
	case VNodeComponent:
		return "Component"
	case VNodeFragment:
		return "Fragment"
	default:
		return "Unknown"
	}
}

// Props represents a map of properties for a VNode
type Props map[string]interface{}

// Get returns a property value
func (p Props) Get(key string) interface{} {
	if p == nil {
		return nil
	}
	return p[key]
}

// Set sets a property value
func (p Props) Set(key string, value interface{}) Props {
	if p == nil {
		p = make(Props)
	}
	p[key] = value
	return p
}

// GetString returns a string property
func (p Props) GetString(key string) string {
	if v, ok := p.Get(key).(string); ok {
		return v
	}
	return ""
}

// GetInt returns an int property
func (p Props) GetInt(key string) int {
	if v, ok := p.Get(key).(int); ok {
		return v
	}
	return 0
}

// GetBool returns a bool property
func (p Props) GetBool(key string) bool {
	if v, ok := p.Get(key).(bool); ok {
		return v
	}
	return false
}

// GetFunc returns a function property
func (p Props) GetFunc(key string) func() {
	if v, ok := p.Get(key).(func()); ok {
		return v
	}
	return nil
}

// Merge merges another Props into this one
func (p Props) Merge(other Props) Props {
	result := make(Props)
	for k, v := range p {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

// Clone creates a copy of the Props
func (p Props) Clone() Props {
	result := make(Props, len(p))
	for k, v := range p {
		result[k] = v
	}
	return result
}

// ComponentFunc represents a function component that returns a VNode
type ComponentFunc func() VNode

// ComponentFuncWithProps represents a component that accepts props
type ComponentFuncWithProps func(Props) VNode
