package inspector

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/ui"
)

// Sidebar provides a detailed information panel for inspected elements
type Sidebar struct {
	width      int
	height     int
	enabled    bool
	collapsed  map[string]bool // Track collapsed sections
	showPaths  bool            // Show element paths
	showProps  bool            // Show properties
}

// NewSidebar creates a new Sidebar instance
func NewSidebar() *Sidebar {
	return &Sidebar{
		width:     40, // Default width
		height:    20, // Default height
		enabled:   true,
		collapsed: make(map[string]bool),
		showPaths: true,
		showProps: true,
	}
}

// SetWidth sets the sidebar width
func (s *Sidebar) SetWidth(width int) {
	s.width = width
}

// SetHeight sets the sidebar height
func (s *Sidebar) SetHeight(height int) {
	s.height = height
}

// Enable enables the sidebar
func (s *Sidebar) Enable() {
	s.enabled = true
}

// Disable disables the sidebar
func (s *Sidebar) Disable() {
	s.enabled = false
}

// IsEnabled returns whether the sidebar is enabled
func (s *Sidebar) IsEnabled() bool {
	return s.enabled
}

// ToggleSection toggles a section's collapsed state
func (s *Sidebar) ToggleSection(section string) {
	s.collapsed[section] = !s.collapsed[section]
}

// SetShowPaths controls whether to show element paths
func (s *Sidebar) SetShowPaths(show bool) {
	s.showPaths = show
}

// SetShowProps controls whether to show properties
func (s *Sidebar) SetShowProps(show bool) {
	s.showProps = show
}

// FormatSidebar formats the sidebar content for a given element
func (s *Sidebar) FormatSidebar(info ElementInfo) string {
	if !s.enabled {
		return ""
	}

	var lines []string

	// Header
	lines = append(lines, "┌─ UI Inspector "+strings.Repeat("─", s.width-18)+"┐")

	// Element type and label
	lines = append(lines, fmt.Sprintf("│ Element: %-40s │", formatTruncate(info.Type, 40)))

	// Type section
	if !s.collapsed["type"] {
		lines = append(lines, "│ ├── Type                              │")
		lines = append(lines, fmt.Sprintf("│ │   VNode Type: %-24s │", info.Type))
		lines = append(lines, fmt.Sprintf("│ │   Tag: %-30s │", info.Tag))
		lines = append(lines, fmt.Sprintf("│ │   Key: %-30s │", info.Key))
	} else {
		lines = append(lines, "│ ├── + Type (collapsed)                │")
	}

	// Position section
	if !s.collapsed["position"] {
		lines = append(lines, "│ ├── Position                          │")
		lines = append(lines, fmt.Sprintf("│ │   X: %-32d │", info.Position.X))
		lines = append(lines, fmt.Sprintf("│ │   Y: %-32d │", info.Position.Y))
	} else {
		lines = append(lines, "│ ├── + position (collapsed)            │")
	}

	// Size section
	if !s.collapsed["size"] {
		lines = append(lines, "│ ├── Size                              │")
		lines = append(lines, fmt.Sprintf("│ │   Width: %-30d │", info.Size.Width))
		lines = append(lines, fmt.Sprintf("│ │   Height: %-29d │", info.Size.Height))
	} else {
		lines = append(lines, "│ ├── + size (collapsed)                │")
	}

	// Layout section
	if !s.collapsed["layout"] {
		lines = append(lines, "│ ├── Layout                            │")
		lines = append(lines, fmt.Sprintf("│ │   Natural Width: %-22d │", info.Layout.NaturalWidth))
		layoutWidthLine := fmt.Sprintf("│ │   Layout Width: %-22d", info.Layout.LayoutWidth)
		if info.Layout.LayoutWidth > info.Layout.NaturalWidth {
			layoutWidthLine += " ✅"
		}
		layoutWidthLine += " │"
		lines = append(lines, layoutWidthLine)
		if info.Layout.Padding > 0 {
			lines = append(lines, fmt.Sprintf("│ │   Padding: +%-27d │", info.Layout.Padding))
		}
		if info.Layout.Flex > 0 {
			lines = append(lines, fmt.Sprintf("│ │   Flex: %-31d │", info.Layout.Flex))
		}
		if info.Layout.Align != "" {
			lines = append(lines, fmt.Sprintf("│ │   Align: %-29s │", info.Layout.Align))
		}
	} else {
		lines = append(lines, "│ ├── + layout (collapsed)               │")
	}

	// Bounds section
	if !s.collapsed["bounds"] {
		lines = append(lines, "│ ├── Bounds                            │")
		lines = append(lines, fmt.Sprintf("│ │   [x: %d, y: %d, w: %d, h: %d]         │",
			info.Bounds[0], info.Bounds[1], info.Bounds[2], info.Bounds[3]))
	} else {
		lines = append(lines, "│ ├── + bounds (collapsed)              │")
	}

	// Constraints section
	if !s.collapsed["constraints"] {
		lines = append(lines, "│ ├── Constraints                       │")
		lines = append(lines, fmt.Sprintf("│ │   MinWidth: %-27d │", info.Constraints.MinWidth))
		lines = append(lines, fmt.Sprintf("│ │   MaxWidth: %-27d │", info.Constraints.MaxWidth))
		lines = append(lines, fmt.Sprintf("│ │   MinHeight: %-26d │", info.Constraints.MinHeight))
		lines = append(lines, fmt.Sprintf("│ │   MaxHeight: %-26d │", info.Constraints.MaxHeight))
	} else {
		lines = append(lines, "│ ├── + constraints (collapsed)         │")
	}

	// Properties section
	if s.showProps && len(info.Properties) > 0 {
		if !s.collapsed["properties"] {
			lines = append(lines, "│ ├── Properties                        │")
			for key, value := range info.Properties {
				propStr := fmt.Sprintf("%s: %v", key, value)
				if len(propStr) > 36 {
					propStr = propStr[:33] + "..."
				}
				lines = append(lines, fmt.Sprintf("│ │   %-36s │", propStr))
			}
		} else {
			lines = append(lines, "│ ├── + properties (collapsed)          │")
		}
	}

	// Path section (if enabled)
	if s.showPaths && info.Path != "" {
		if !s.collapsed["path"] {
			lines = append(lines, "│ ├── Path                              │")
			pathStr := info.Path
			if len(pathStr) > 36 {
				pathStr = pathStr[:33] + "..."
			}
			lines = append(lines, fmt.Sprintf("│ │   %-36s │", pathStr))
		} else {
			lines = append(lines, "│ ├── + path (collapsed)                 │")
		}
	}

	// Footer
	lines = append(lines, "└"+strings.Repeat("─", s.width-2)+"┘")

	return strings.Join(lines, "\n")
}

// FormatCompact formats a compact version of element info
func (s *Sidebar) FormatCompact(info ElementInfo) string {
	var parts []string

	// Type and label
	if info.Label != "" {
		parts = append(parts, fmt.Sprintf("%s(%s)", info.Type, info.Label))
	} else {
		parts = append(parts, info.Type)
	}

	// Position
	parts = append(parts, fmt.Sprintf("@(%d,%d)", info.Position.X, info.Position.Y))

	// Size
	parts = append(parts, fmt.Sprintf("%dx%d", info.Size.Width, info.Size.Height))

	// Layout info
	if info.Layout.Flex > 0 {
		parts = append(parts, fmt.Sprintf("flex:%d", info.Layout.Flex))
	}
	if info.Layout.NaturalWidth != info.Layout.LayoutWidth {
		parts = append(parts, fmt.Sprintf("nat:%d->%d",
			info.Layout.NaturalWidth, info.Layout.LayoutWidth))
	}

	return strings.Join(parts, " ")
}

// FormatTable formats multiple elements in a table format
func (s *Sidebar) FormatTable(elements []ElementInfo) string {
	if len(elements) == 0 {
		return "No elements"
	}

	var lines []string
	lines = append(lines, "┌─ Elements "+strings.Repeat("─", max(0, s.width-30))+"┐")

	for i, elem := range elements {
		prefix := "│ "
		if i < len(elements)-1 {
			prefix += "├─"
		} else {
			prefix += "└─"
		}

		compact := s.FormatCompact(elem)
		lines = append(lines, fmt.Sprintf("%s %s", prefix, formatTruncate(compact, s.width-10)))
	}

	lines = append(lines, "└"+strings.Repeat("─", s.width-2)+"┘")
	return strings.Join(lines, "\n")
}

// GetCopyableText returns all information in a format suitable for copying
func (s *Sidebar) GetCopyableText(info ElementInfo) string {
	var lines []string

	lines = append(lines, "=== UI Inspector Element Info ===")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Type: %s", info.Type))
	lines = append(lines, fmt.Sprintf("Tag: %s", info.Tag))
	lines = append(lines, fmt.Sprintf("Key: %s", info.Key))
	lines = append(lines, fmt.Sprintf("Label: %s", info.Label))
	lines = append(lines, "")
	lines = append(lines, "Position:")
	lines = append(lines, fmt.Sprintf("  X: %d", info.Position.X))
	lines = append(lines, fmt.Sprintf("  Y: %d", info.Position.Y))
	lines = append(lines, "")
	lines = append(lines, "Size:")
	lines = append(lines, fmt.Sprintf("  Width: %d", info.Size.Width))
	lines = append(lines, fmt.Sprintf("  Height: %d", info.Size.Height))
	lines = append(lines, "")
	lines = append(lines, "Layout:")
	lines = append(lines, fmt.Sprintf("  Natural Width: %d", info.Layout.NaturalWidth))
	lines = append(lines, fmt.Sprintf("  Layout Width: %d", info.Layout.LayoutWidth))
	lines = append(lines, fmt.Sprintf("  Padding: %d", info.Layout.Padding))
	lines = append(lines, fmt.Sprintf("  Flex: %d", info.Layout.Flex))
	lines = append(lines, fmt.Sprintf("  Align: %s", info.Layout.Align))
	lines = append(lines, "")
	lines = append(lines, "Bounds:")
	lines = append(lines, fmt.Sprintf("  [%d, %d, %d, %d]",
		info.Bounds[0], info.Bounds[1], info.Bounds[2], info.Bounds[3]))
	lines = append(lines, "")
	lines = append(lines, "Constraints:")
	lines = append(lines, fmt.Sprintf("  MinWidth: %d", info.Constraints.MinWidth))
	lines = append(lines, fmt.Sprintf("  MaxWidth: %d", info.Constraints.MaxWidth))
	lines = append(lines, fmt.Sprintf("  MinHeight: %d", info.Constraints.MinHeight))
	lines = append(lines, fmt.Sprintf("  MaxHeight: %d", info.Constraints.MaxHeight))
	lines = append(lines, "")

	if len(info.Properties) > 0 {
		lines = append(lines, "Properties:")
		for key, value := range info.Properties {
			lines = append(lines, fmt.Sprintf("  %s: %v", key, value))
		}
		lines = append(lines, "")
	}

	if info.Path != "" {
		lines = append(lines, fmt.Sprintf("Path: %s", info.Path))
	}

	return strings.Join(lines, "\n")
}

// BuildVNode builds a VNode representation of the sidebar
func (s *Sidebar) BuildVNode(info ElementInfo) ui.VNode {
	if !s.enabled || info.Type == "" {
		return ui.Text("")
	}

	// Build sidebar content as text
	content := s.FormatSidebar(info)

	// Return as a text node
	return ui.Bordered().
		Child(ui.Text(content)).
		Width(s.width).
		Build()
}

// BuildCompactVNode builds a compact VNode for the sidebar
func (s *Sidebar) BuildCompactVNode(info ElementInfo) ui.VNode {
	if !s.enabled || info.Type == "" {
		return ui.Text("")
	}

	content := s.FormatCompact(info)
	return ui.Text(content)
}

// Helper functions

// formatTruncate truncates a string to fit within max width
func formatTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
