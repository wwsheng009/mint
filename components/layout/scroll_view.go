package layout

import (
	"fmt"
	"os"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ScrollView is a scrollable container that clips its content to the viewport size
// It supports vertical scrolling and displays a scroll position indicator
//
// Usage:
//   scrollView := layout.NewScrollView(content).
//       Width(80).
//       Height(20).
//       ScrollOffset(5).  // Current scroll position
//       Build()
type ScrollView struct {
	*ui.ElementVNode

	// Configuration
	content      ui.VNode  // Content to display
	scrollOffset int       // Vertical scroll offset (in lines)
	width        int       // Viewport width
	height       int       // Total content height (calculated during Build)
	totalLines   int       // Total content height
}

// ScrollViewBuilder builds scrollable containers
type ScrollViewBuilder struct {
	node         *ScrollView
	content      ui.VNode
	width        int
	height       int
	scrollOffset int
	showBorder   bool
}

// NewScrollView creates a new scrollable view builder
func NewScrollView(content ui.VNode) *ScrollViewBuilder {
	return &ScrollViewBuilder{
		content:      content,
		width:        80,
		height:       20,
		scrollOffset: 0,
		showBorder:   false,
		node: &ScrollView{
			ElementVNode: ui.NewElement("scroll-view"),
		},
	}
}

// Width sets the viewport width
func (b *ScrollViewBuilder) Width(n int) *ScrollViewBuilder {
	b.width = n
	return b
}

// Height sets the viewport height (number of visible lines)
func (b *ScrollViewBuilder) Height(n int) *ScrollViewBuilder {
	b.height = n
	return b
}

// ScrollOffset sets the current vertical scroll position
func (b *ScrollViewBuilder) ScrollOffset(offset int) *ScrollViewBuilder {
	b.scrollOffset = offset
	return b
}

// ShowBorder adds a border around the viewport
func (b *ScrollViewBuilder) ShowBorder(show bool) *ScrollViewBuilder {
	b.showBorder = show
	return b
}

// Style sets the visual style
func (b *ScrollViewBuilder) Style(s style.Style) *ScrollViewBuilder {
	b.node.SetStyle(s)
	return b
}

// Key sets key for diffing
func (b *ScrollViewBuilder) Key(key string) *ScrollViewBuilder {
	b.node.SetKey(key)
	return b
}

// Build creates the scrollable view VNode
func (b *ScrollViewBuilder) Build() ui.VNode {
	// Calculate content as text lines
	contentText := b.extractTextContent(b.content)
	lines := strings.Split(contentText, "\n")
	b.node.totalLines = len(lines)

	// Auto-height mode: if height is 0 or not set, render all content
	// Parent container (with ui.Flex) will constrain the actual height
	if b.height <= 0 {
		// Render all content without clipping
		visibleText := contentText

		// DEBUG: Log what we extracted
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			lineCount := strings.Count(contentText, "\n") + 1
			if len(contentText) == 0 {
				fmt.Fprintf(os.Stderr, "[ScrollView] Auto-height mode: NO CONTENT EXTRACTED!\n")
				fmt.Fprintf(os.Stderr, "[ScrollView] Input type: %T\n", b.content)
			} else {
				fmt.Fprintf(os.Stderr, "[ScrollView] Auto-height mode: extracted %d lines\n", lineCount)
				fmt.Fprintf(os.Stderr, "[ScrollView] First 200 chars: %q\n", contentText[:min(200, len(contentText))])
			}
		}

		// Create text node for all content
		textNode := ui.Text(visibleText)

		// Apply style if set
		if b.node.Style().FG != "" || b.node.Style().BG != "" {
			textNode.SetStyle(b.node.Style())
		}

		// Store configuration
		b.node.content = b.content
		b.node.width = b.width
		b.node.height = 0  // Auto-height

		// Mark as flexible for parent layout
		textNode.SetProps(ui.Props{
			"flex":           1,              // Allow to grow
			"scroll-content":  contentText,    // Store original for scrolling
			"scroll-offset":  b.scrollOffset, // Store scroll position
			"total-lines":    b.node.totalLines,
		})

		// Wrap in VStackBuilder to maintain consistent type (LayoutNode)
		result := ui.VStackBuilder(textNode).
			Width(b.width).
			Build()

		// Copy style to wrapper
		if b.node.Style().FG != "" || b.node.Style().BG != "" {
			result.SetStyle(b.node.Style())
		}

		return result
	}

	// Fixed-height mode: clip content to viewport
	// Clamp scroll offset to valid range
	maxOffset := len(lines) - b.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if b.scrollOffset < 0 {
		b.scrollOffset = 0
	}
	if b.scrollOffset > maxOffset {
		b.scrollOffset = maxOffset
	}
	b.node.scrollOffset = b.scrollOffset

	// Calculate visible range
	startLine := b.scrollOffset
	endLine := startLine + b.height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Extract visible lines
	var visibleLines []string
	if startLine < len(lines) {
		visibleLines = lines[startLine:endLine]
	} else {
		visibleLines = []string{}
	}

	// Build visible content
	visibleText := strings.Join(visibleLines, "\n")

	// Add scroll indicator if content is scrollable
	if b.node.totalLines > b.height {
		// Add scroll position indicator at the end
		indicator := " ▼"
		if b.scrollOffset >= maxOffset {
			indicator = " ▲" // At bottom
		} else if b.scrollOffset == 0 {
			indicator = " ▼" // At top
		} else {
			indicator = " ↕" // In middle
		}
		scrollInfo := strings.Repeat(" ", b.width-5) + indicator
		if len(scrollInfo) > b.width {
			scrollInfo = scrollInfo[:b.width]
		}
		visibleText += "\n" + scrollInfo
	}

	// Create text node for visible content
	textNode := ui.Text(visibleText)

	// Apply style if set
	if b.node.Style().FG != "" || b.node.Style().BG != "" {
		textNode.SetStyle(b.node.Style())
	}

	// Store configuration for event handling
	b.node.content = b.content
	b.node.width = b.width
	b.node.height = b.height

	// Wrap with border if requested
	if b.showBorder {
		return rtui.Bordered().
			Child(textNode).
			Width(b.width).
			Height(b.height).
			Build()
	}

	// Otherwise, wrap in VStackBuilder for size constraints
	result := ui.VStackBuilder(textNode).
		Width(b.width).
		Height(b.height).
		Build()

	// Copy style
	if b.node.Style().FG != "" || b.node.Style().BG != "" {
		result.SetStyle(b.node.Style())
	}

	return result
}

// extractTextContent extracts text content from a VNode
// This handles various node types (Text, VStack, HStack, etc.)
func (b *ScrollViewBuilder) extractTextContent(node ui.VNode) string {
	if node == nil {
		return ""
	}

	// Check for "content" property (new style text elements)
	if props := node.Props(); props != nil {
		if content, ok := props["content"]; ok && content != "" {
			if contentStr, ok := content.(string); ok {
				return contentStr
			}
		}
	}

	// Check if it implements Content() method (old style)
	if contentNode, ok := node.(interface{ Content() string }); ok {
		return contentNode.Content()
	}

	// Legacy: Check if it's a TextVNode
	if textNode, ok := node.(*rtui.TextVNode); ok {
		return textNode.Content()
	}

	// For element nodes, recursively extract from children
	var result strings.Builder
	children := node.Children()
	for i, child := range children {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(b.extractTextContent(child))
	}

	return result.String()
}

// =============================================================================
// ScrollView Instance Methods (for runtime state management)
// =============================================================================

// ScrollBy scrolls by the given delta
func (sv *ScrollView) ScrollBy(delta int) int {
	newOffset := sv.scrollOffset + delta

	// Clamp to valid range
	maxOffset := sv.totalLines - sv.height
	if maxOffset < 0 {
		maxOffset = 0
	}

	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset > maxOffset {
		newOffset = maxOffset
	}

	sv.scrollOffset = newOffset
	return newOffset
}

// ScrollTo scrolls to an absolute position
func (sv *ScrollView) ScrollTo(offset int) int {
	maxOffset := sv.totalLines - sv.height
	if maxOffset < 0 {
		maxOffset = 0
	}

	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	sv.scrollOffset = offset
	return offset
}

// ScrollTop scrolls to the top
func (sv *ScrollView) ScrollTop() {
	sv.scrollOffset = 0
}

// ScrollBottom scrolls to the bottom
func (sv *ScrollView) ScrollBottom() {
	maxOffset := sv.totalLines - sv.height
	if maxOffset < 0 {
		sv.scrollOffset = 0
	} else {
		sv.scrollOffset = maxOffset
	}
}

// PageUp scrolls up by one page
func (sv *ScrollView) PageUp() int {
	return sv.ScrollBy(-sv.height)
}

// PageDown scrolls down by one page
func (sv *ScrollView) PageDown() int {
	return sv.ScrollBy(sv.height)
}

// CanScrollUp returns true if can scroll up
func (sv *ScrollView) CanScrollUp() bool {
	return sv.scrollOffset > 0
}

// CanScrollDown returns true if can scroll down
func (sv *ScrollView) CanScrollDown() bool {
	maxOffset := sv.totalLines - sv.height
	return sv.scrollOffset < maxOffset && maxOffset > 0
}

// GetScrollOffset returns current scroll offset
func (sv *ScrollView) GetScrollOffset() int {
	return sv.scrollOffset
}

// GetTotalLines returns total content height
func (sv *ScrollView) GetTotalLines() int {
	return sv.totalLines
}

// GetViewportSize returns viewport height
func (sv *ScrollView) GetViewportSize() int {
	return sv.height
}

// IsScrollable returns true if content is larger than viewport
func (sv *ScrollView) IsScrollable() bool {
	return sv.totalLines > sv.height
}
