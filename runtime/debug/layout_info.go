package debug

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// LayoutInfo contains layout information for a single component
type LayoutInfo struct {
	// Component identification
	Type     string // VNode type (button, text, hstack, etc.)
	Tag      string // Tag if available
	Key      string // Key if available
	Label    string // Label for buttons/text
	Path     string // Path from root (e.g., "root.0.child.2")

	// Position and size
	X      int    // X position (absolute)
	Y      int    // Y position (absolute)
	Width  int    // Width
	Height int    // Height

	// Constraints
	MinWidth  int // Minimum width constraint
	MaxWidth  int // Maximum width constraint
	MinHeight int // Minimum height constraint
	MaxHeight int // Maximum height constraint

	// Layout properties
	Flex        int    // Flex factor (0 if not flexible)
	Gap         int    // Gap for layout containers
	Align       string // Main-axis alignment
	CrossAlign  string // Cross-axis alignment
	Padding     [4]int // Padding [top, right, bottom, left]
	Margin      [4]int // Margin [top, right, bottom, left]
	IsContainer bool   // True if this is a layout container

	// Children
	Children []LayoutInfo // Child layout info
}

// LayoutTree represents the complete layout tree
type LayoutTree struct {
	Root LayoutInfo
}

// GetLayoutTree extracts layout information from a ComputedLayout
func GetLayoutTree(layout *compute.ComputedLayout) *LayoutTree {
	if layout == nil || layout.Root == nil {
		return &LayoutTree{}
	}

	return &LayoutTree{
		Root: extractLayoutInfo(layout.Root, ""),
	}
}

// extractLayoutInfo recursively extracts layout info from a ComputedBox
func extractLayoutInfo(box *compute.ComputedBox, path string) LayoutInfo {
	if box == nil {
		return LayoutInfo{}
	}

	info := LayoutInfo{
		X:      box.Box.X,
		Y:      box.Box.Y,
		Width:  box.Box.Width,
		Height: box.Box.Height,
	}

	// Extract VNode information
	if box.VNode != nil {
		info.Type = box.VNode.Type().String()

		// Get tag
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
			info.Tag = tagger.Tag()
		}

		// Get key
		info.Key = box.VNode.Key()

		// Get label (for buttons)
		if labeler, ok := box.VNode.(interface{ Label() string }); ok {
			info.Label = labeler.Label()
		}

		// Get layout properties
		if layoutInfo := rtui.GetLayoutInfo(box.VNode); layoutInfo.Flex > 0 {
			info.Flex = layoutInfo.Flex
		}
		if layoutInfo := rtui.GetLayoutInfo(box.VNode); layoutInfo.Gap > 0 {
			info.Gap = layoutInfo.Gap
		}
		if layoutInfo := rtui.GetLayoutInfo(box.VNode); layoutInfo.Align != rtui.AlignStart {
			info.Align = alignToString(layoutInfo.Align)
		}
		if layoutInfo := rtui.GetLayoutInfo(box.VNode); layoutInfo.CrossAlign != rtui.AlignStart {
			info.CrossAlign = alignToString(layoutInfo.CrossAlign)
		}
		if boxModel, ok := box.VNode.(interface {
			Padding() [4]int
			Margin() [4]int
		}); ok {
			info.Padding = boxModel.Padding()
			info.Margin = boxModel.Margin()
		}

		// Check if this is a container
		if _, ok := box.VNode.(*rtui.LayoutNode); ok {
			info.IsContainer = true
		}
		if _, ok := box.VNode.(interface{ Tag() string }); ok {
			tag := box.VNode.(interface{ Tag() string }).Tag()
			if tag == "hstack" || tag == "vstack" || tag == "wrap" {
				info.IsContainer = true
			}
		}
	}

	// Extract children
	for i, child := range box.Children {
		childPath := fmt.Sprintf("%s.%d", path, i)
		childInfo := extractLayoutInfo(child, childPath)
		childInfo.Path = childPath
		info.Children = append(info.Children, childInfo)
	}

	return info
}

// alignToString converts Align to string
func alignToString(align rtui.Align) string {
	switch align {
	case rtui.AlignStart:
		return "Start"
	case rtui.AlignCenter:
		return "Center"
	case rtui.AlignEnd:
		return "End"
	case rtui.AlignSpaceBetween:
		return "SpaceBetween"
	case rtui.AlignSpaceAround:
		return "SpaceAround"
	default:
		return fmt.Sprintf("Unknown(%d)", align)
	}
}

// FormatLayoutTree formats the layout tree as a string
func FormatLayoutTree(tree *LayoutTree) string {
	var builder strings.Builder
	builder.WriteString("Layout Tree:\n")
	builder.WriteString("═" + strings.Repeat("═", 78) + "═\n")
	formatLayoutInfo(&tree.Root, &builder, 0)
	builder.WriteString("═" + strings.Repeat("═", 78) + "═\n")
	return builder.String()
}

// formatLayoutInfo recursively formats layout info
func formatLayoutInfo(info *LayoutInfo, builder *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	// Header
	if depth == 0 {
		builder.WriteString(fmt.Sprintf("%s📍 Root (%s)", indent, info.Type))
	} else {
		builder.WriteString(fmt.Sprintf("%s├─ %s (%s)", indent, info.Path, info.Type))
	}

	// Tag/Key/Label
	if info.Tag != "" {
		builder.WriteString(fmt.Sprintf(" tag=%s", info.Tag))
	}
	if info.Key != "" {
		builder.WriteString(fmt.Sprintf(" key=%s", info.Key))
	}
	if info.Label != "" {
		builder.WriteString(fmt.Sprintf(" label=%q", info.Label))
	}

	// Position and Size
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("%s   Position: (%d, %d) Size: %dx%d\n",
		indent, info.X, info.Y, info.Width, info.Height))

	// Flex
	if info.Flex > 0 {
		builder.WriteString(fmt.Sprintf("%s   Flex: %d\n", indent, info.Flex))
	}

	// Container properties
	if info.IsContainer {
		parts := []string{}
		if info.Gap > 0 {
			parts = append(parts, fmt.Sprintf("gap=%d", info.Gap))
		}
		if info.Align != "" {
			parts = append(parts, fmt.Sprintf("align=%s", info.Align))
		}
		if info.CrossAlign != "" {
			parts = append(parts, fmt.Sprintf("crossAlign=%s", info.CrossAlign))
		}
		if len(parts) > 0 {
			builder.WriteString(fmt.Sprintf("%s   Layout: %s\n", indent, strings.Join(parts, ", ")))
		}
	}

	// Padding/Margin
	if info.Padding != [4]int{} {
		builder.WriteString(fmt.Sprintf("%s   Padding: [%d %d %d %d]\n",
			indent, info.Padding[0], info.Padding[1], info.Padding[2], info.Padding[3]))
	}
	if info.Margin != [4]int{} {
		builder.WriteString(fmt.Sprintf("%s   Margin: [%d %d %d %d]\n",
			indent, info.Margin[0], info.Margin[1], info.Margin[2], info.Margin[3]))
	}

	// Children
	for i := range info.Children {
		formatLayoutInfo(&info.Children[i], builder, depth+1)
		if i < len(info.Children)-1 && depth == 0 {
			// Add separator between top-level items
			// builder.WriteString("\n")
		}
	}
}

// GetComponentInfo finds a component by path and returns its layout info
func GetComponentInfo(tree *LayoutTree, path string) (LayoutInfo, bool) {
	if path == "root" || path == "" {
		return tree.Root, true
	}

	return findComponentByPath(&tree.Root, path)
}

// findComponentByPath recursively searches for a component by path
func findComponentByPath(info *LayoutInfo, path string) (LayoutInfo, bool) {
	if info.Path == path {
		return *info, true
	}

	for _, child := range info.Children {
		if found, ok := findComponentByPath(&child, path); ok {
			return found, true
		}
	}

	return LayoutInfo{}, false
}

// FindComponentsByType finds all components of a specific type
func FindComponentsByType(tree *LayoutTree, vtype string) []LayoutInfo {
	var results []LayoutInfo
	findByType(&tree.Root, vtype, &results)
	return results
}

// findByType recursively finds components by type
func findByType(info *LayoutInfo, vtype string, results *[]LayoutInfo) {
	if info.Type == vtype || info.Tag == vtype {
		*results = append(*results, *info)
	}

	for _, child := range info.Children {
		findByType(&child, vtype, results)
	}
}

// GetComponentAtPoint finds the component at a specific position
func GetComponentAtPoint(tree *LayoutTree, x, y int) (LayoutInfo, bool) {
	return findAtPoint(&tree.Root, x, y)
}

// findAtPoint recursively finds component at position
func findAtPoint(info *LayoutInfo, x, y int) (LayoutInfo, bool) {
	// Check if point is within this component's bounds
	if x >= info.X && x < info.X+info.Width &&
		y >= info.Y && y < info.Y+info.Height {

		// Check children first (they're on top)
		for i := len(info.Children) - 1; i >= 0; i-- {
			if child, ok := findAtPoint(&info.Children[i], x, y); ok {
				return child, true
			}
		}

		// No child matched, return this component
		return *info, true
	}

	return LayoutInfo{}, false
}
