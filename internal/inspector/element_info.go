package inspector

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime"
)

// ElementInfo represents extracted information from a VNode
type ElementInfo struct {
	// VNode reference
	VNode rtui.VNode

	// Identification
	Type   string // VNode type name
	Tag    string // Tag if available
	Key    string // Key if available
	Label  string // Label for buttons/text
	Path   string // Path from root (e.g., "root.0.1")

	// Position and Size
	Position Position
	Size     Size

	// Layout Information
	Layout    LayoutInfo
	Bounds    [4]int // [x, y, width, height] from SetBounds
	Constraints runtime.BoxConstraints

	// Additional Properties
	Properties map[string]interface{}
}

// Position represents X, Y coordinates
type Position struct {
	X int
	Y int
}

// Size represents width and height
type Size struct {
	Width  int
	Height int
}

// LayoutInfo contains layout-specific information
type LayoutInfo struct {
	NaturalWidth  int   // Content width without constraints
	LayoutWidth   int   // Width after flex calculation
	Padding       int   // Total padding
	Flex          int   // Flex coefficient
	IsFlexChild   bool  // Whether this is a flex child
	Align         string // Text alignment
}

// ExtractElementInfo extracts information from a VNode
func ExtractElementInfo(vnode rtui.VNode) ElementInfo {
	info := ElementInfo{
		VNode:      vnode,
		Properties: make(map[string]interface{}),
	}

	// Early return for nil
	if vnode == nil {
		info.Type = "nil"
		return info
	}

	// Extract basic identification
	info.Type = getTypeName(vnode)
	info.Tag = getTag(vnode)
	info.Key = getKey(vnode)
	info.Label = getLabel(vnode)
	info.Path = getPath(vnode)

	// Extract bounds if available
	if boundsAware, ok := vnode.(interface{ GetBounds() [4]int }); ok {
		info.Bounds = boundsAware.GetBounds()
		info.Position.X = info.Bounds[0]
		info.Position.Y = info.Bounds[1]
		info.Size.Width = info.Bounds[2]
		info.Size.Height = info.Bounds[3]
	}

	// Extract layout information
	extractLayoutInfo(vnode, &info)

	// Extract component-specific properties
	extractProperties(vnode, &info)

	return info
}

// getTypeName gets the type name of a VNode
func getTypeName(vnode rtui.VNode) string {
	if vnode == nil {
		return "nil"
	}
	t := reflect.TypeOf(vnode)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// getTag gets the tag from a VNode
func getTag(vnode rtui.VNode) string {
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return ""
}

// getKey gets the key from a VNode
func getKey(vnode rtui.VNode) string {
	if keyer, ok := vnode.(interface{ Key() string }); ok {
		return keyer.Key()
	}
	return ""
}

// getPath extracts the hierarchical path from a VNode's Fiber Key
// Fiber reconciliation sets VNode keys to path-based keys like /root/base[0]/vstack[0]/panel[0]
func getPath(vnode rtui.VNode) string {
	if keyer, ok := vnode.(interface{ Key() string }); ok {
		vnodeKey := keyer.Key()
		// Check if this is a path-based key (set by Fiber reconciliation)
		if vnodeKey != "" && strings.HasPrefix(vnodeKey, "/root/") {
			// Use the Fiber-generated path as our display path
			// Remove the "/root/" prefix for cleaner display
			// /root/base[0]/vstack[0]/panel[0] → base[0]/vstack[0]/panel[0]
			if len(vnodeKey) > 6 { // "/root/" is 6 characters
				return vnodeKey[6:] // Skip "/root/" prefix
			}
			return vnodeKey
		}
	}
	return ""
}

// getLabel gets the label from a VNode (for buttons, text, etc.)
func getLabel(vnode rtui.VNode) string {
	// Try Label() method
	if labeler, ok := vnode.(interface{ Label() string }); ok {
		return labeler.Label()
	}

	// Try extracting from Text content
	if text := rtui.GetTextContent(vnode); text != "" {
		// Truncate long text to 20 characters maximum
		// Use rune-based truncation to avoid cutting UTF-8 multibyte characters
		if utf8.RuneCountInString(text) > 20 {
			runes := []rune(text)
			return string(runes[:17]) + "..."
		}
		return text
	}

	return ""
}

// extractLayoutInfo extracts layout-specific information
func extractLayoutInfo(vnode rtui.VNode, info *ElementInfo) {
	// Check if it's a flex child
	if props := vnode.Props(); props != nil {
		if flex, ok := props["flex"].(int); ok {
			info.Layout.Flex = flex
			info.Layout.IsFlexChild = flex > 0
		}
		if align, ok := props["textAlign"].(runtime.TextAlign); ok {
			info.Layout.Align = textAlignToString(align)
		}
	}

	// Extract natural width for components that support measurement
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		// Measure with unbounded constraints to get natural size
		naturalSize := measurable.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		})
		info.Layout.NaturalWidth = naturalSize.Width

		// Layout width comes from bounds
		if info.Bounds[2] > 0 {
			info.Layout.LayoutWidth = info.Bounds[2]
		} else {
			info.Layout.LayoutWidth = naturalSize.Width
		}
	}

	// For Button, extract label width
	if labeler, ok := vnode.(interface{ Label() string }); ok {
		label := labeler.Label()
		if label != "" {
			// Calculate natural width: label + brackets
			info.Layout.NaturalWidth = utf8.RuneCountInString(label) + 2

			// Add padding if BoxModel is available
			if boxModel, ok := vnode.(interface {
				GetPadding() [4]int
			}); ok {
				padding := boxModel.GetPadding()
				if padding[1] > 0 || padding[3] > 0 { // right or left padding
					info.Layout.Padding = padding[1] + padding[3]
				}
			}

			// Add focus indicator space
			info.Layout.NaturalWidth += 2 // for focus indicator
		}
	}

	// For Text, extract content width
	if text := rtui.GetTextContent(vnode); text != "" {
		info.Layout.NaturalWidth = utf8.RuneCountInString(text)

		// Layout width from bounds or content
		if info.Bounds[2] > 0 {
			info.Layout.LayoutWidth = info.Bounds[2]
		} else {
			info.Layout.LayoutWidth = info.Layout.NaturalWidth
		}
	}
}

// extractProperties extracts component-specific properties
func extractProperties(vnode rtui.VNode, info *ElementInfo) {
	// Check for Button-specific properties
	if _, ok := vnode.(interface{ Label() string }); ok {
		if focusable, ok := vnode.(interface{ HasFocus() bool }); ok {
			info.Properties["HasFocus"] = focusable.HasFocus()
		}
		if disabled, ok := vnode.(interface{ Disabled() bool }); ok {
			info.Properties["Disabled"] = disabled.Disabled()
		}
	}

	// Check for BoxModel properties
	if boxModel, ok := vnode.(interface {
		GetPadding() [4]int
		GetMargin() [4]int
	}); ok {
		padding := boxModel.GetPadding()
		margin := boxModel.GetMargin()
		if padding != [4]int{} {
			info.Properties["Padding"] = fmt.Sprintf("[T:%d R:%d B:%d L:%d]",
				padding[0], padding[1], padding[2], padding[3])
		}
		if margin != [4]int{} {
			info.Properties["Margin"] = fmt.Sprintf("[T:%d R:%d B:%d L:%d]",
				margin[0], margin[1], margin[2], margin[3])
		}
	}

	// Extract all props for debugging
	if props := vnode.Props(); props != nil && len(props) > 0 {
		info.Properties["Props"] = fmt.Sprintf("%+v", props)
	}
}

// FormatElementInfo formats ElementInfo for display
func FormatElementInfo(info ElementInfo) string {
	result := fmt.Sprintf("Element: %s\n", info.Type)

	if info.Tag != "" {
		result += fmt.Sprintf("  Tag: %s\n", info.Tag)
	}
	if info.Key != "" {
		result += fmt.Sprintf("  Key: %s\n", info.Key)
	}

	result += fmt.Sprintf("\nPosition:\n")
	result += fmt.Sprintf("  X: %d\n", info.Position.X)
	result += fmt.Sprintf("  Y: %d\n", info.Position.Y)

	result += fmt.Sprintf("\nSize:\n")
	result += fmt.Sprintf("  Width: %d\n", info.Size.Width)
	result += fmt.Sprintf("  Height: %d\n", info.Size.Height)

	result += fmt.Sprintf("\nLayout:\n")
	result += fmt.Sprintf("  Natural Width: %d\n", info.Layout.NaturalWidth)
	result += fmt.Sprintf("  Layout Width: %d", info.Layout.LayoutWidth)

	if info.Layout.LayoutWidth > info.Layout.NaturalWidth {
		result += " ✅"
	}
	result += "\n"

	if info.Layout.Padding > 0 {
		result += fmt.Sprintf("  Padding: +%d\n", info.Layout.Padding)
	}

	if info.Layout.Flex > 0 {
		result += fmt.Sprintf("  Flex: %d\n", info.Layout.Flex)
	}

	if info.Layout.Align != "" {
		result += fmt.Sprintf("  Align: %s\n", info.Layout.Align)
	}

	if info.Bounds != [4]int{} {
		result += fmt.Sprintf("\nBounds:\n")
		result += fmt.Sprintf("  [x: %d, y: %d, w: %d, h: %d]\n",
			info.Bounds[0], info.Bounds[1], info.Bounds[2], info.Bounds[3])
	}

	if len(info.Properties) > 0 {
		result += "\nProperties:\n"
		for key, value := range info.Properties {
			result += fmt.Sprintf("  %s: %v\n", key, value)
		}
	}

	return result
}

// textAlignToString converts TextAlign to string representation
func textAlignToString(align runtime.TextAlign) string {
	switch align {
	case runtime.TextAlignLeft:
		return "Left"
	case runtime.TextAlignCenter:
		return "Center"
	case runtime.TextAlignRight:
		return "Right"
	default:
		return "Unknown"
	}
}
