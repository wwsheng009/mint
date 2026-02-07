// Package dimension provides design tokens for TUI component dimensions
//
// This package implements the spacing and dimension specification based on
// the TUI design system, following a 4-based rhythm adapted for terminal cell density.
//
// Reference: docs/theme/space_1.md
package dimension

import "sync"

// Spacing tokens following the 4-based rhythm adapted for terminal cells
//
// | Token | Cell | Usage |
// | ----- | ---- | ----- |
// | XS    | 0    | No spacing |
// | S     | 1    | Tight spacing, inline gaps |
// | M     | 2    | Normal spacing, padding |
// | L     | 3    | Large spacing, section gaps |
// | XL    | 4    | Extra large spacing |
// | XXL   | 6    | Huge spacing |
const (
	XS int = 0
	S  int = 1
	M  int = 2
	L  int = 3
	XL int = 4
	XXL int = 6
)

// =============================================================================
// Component Dimension Tokens
// =============================================================================

// Container components

// Panel / Card dimensions
const (
	PanelPadding       = S  // 1 - padding (range 1-2 recommended)
	PanelMargin        = S  // 1 - margin
	PanelMinHeight     = L  // 3 - minimum height
	PanelTitleBarHeight = S  // 1 - title bar height
	PanelBorderWidth    = S  // 1 - border width
)

// Modal / Dialog dimensions
const (
	ModalPadding                = M  // 2
	ModalMarginToScreenEdge     = M  // 2 - minimum margin to screen edge
	ModalMinWidth              = 30 // minimum width
	ModalMaxWidthPercent       = 70 // maximum width as percent of screen
	ModalFooterHeight          = L  // 3 - footer height
)

// Data display components

// Table dimensions
const (
	TableRowHeight  = S  // 1 - row height
	TableHeaderHeight = S  // 1 - header height
	TableCellPadding = S  // 1 - cell padding (left/right)
	TableColumnMinWidth = 6  // minimum column width
	TableColumnMaxWidth = 40 // maximum column width
	TableRowGap         = XS // 0 - gap between rows
)

// List dimensions
const (
	ListItemHeight = S // 1 - item height
	ListPaddingLeft = S // 1 - padding left
	ListGap         = XS // 0 - gap between items
)

// Input components

// Input dimensions
const (
	InputHeight       = S // 1 - height
	InputPaddingLR    = S // 1 - padding left/right
	InputMinWidth     = 10 // minimum width
	InputLabelGap     = S // 1 - gap between label and input
)

// TextArea dimensions
const (
	TextAreaMinHeight = L // 3 - minimum height
	TextAreaPadding   = S // 1 - padding
	TextAreaLineGap    = XS // 0 - gap between lines
)

// Button dimensions
const (
	ButtonHeight       = S          // 1 - height
	ButtonPaddingLR    = M          // 2 - padding left/right
	ButtonMinWidth     = 6          // minimum width
	ButtonGap          = S          // 1 - gap between buttons
)

// Selection components

// Checkbox / Radio dimensions
const (
	CheckBoxWidth  = 3 // `[X]` box width
	CheckBoxGap    = S // 1 - gap to label
)

// Switch dimensions
const (
	SwitchWidth  = 6 // width
	SwitchHeight = S // 1 - height
)

// Navigation components

// Tabs dimensions
const (
	TabHeight = S // 1 - tab height
	TabPadding = M // 2 - padding
	TabGap     = S // 1 - gap between tabs
)

// Sidebar dimensions
const (
	SidebarWidthMin = 18 // minimum width
	SidebarWidthMax = 24 // maximum width
	SidebarItemHeight = S // 1 - item height
	SidebarPadding     = S // 1 - padding
)

// Feedback components

// Alert dimensions
const (
	AlertPadding  = S // 1 - padding
	AlertIconGap   = S // 1 - gap between icon and text
	AlertMinHeight = L // 3 - minimum height
)

// Tooltip dimensions
const (
	TooltipPadding = S // 1 - padding
	TooltipMaxWidth = 40 // maximum width
)

// Progress components

// ProgressBar dimensions
const (
	ProgressBarHeight = S // 1 - height
	ProgressBarMinWidth = 10 // minimum width
)

// Chart dimensions
const (
	ChartMinHeight = XXL // 6 - minimum height
	ChartPadding    = S   // 1 - padding
)

// =============================================================================
// Layout Rules
// =============================================================================

// Vertical spacing between components
const (
	VerticalSpacingComponents = S // 1 - minimum vertical spacing between components
	VerticalSpacingSections   = M // 2 - spacing between sections
)

// Horizontal spacing rules
const (
	FieldLabelToControlGap = S // 1 - gap between field label and control
	ButtonGroupGap          = S // 1 - gap between buttons in a group
	FormColumnGap            = L // 3 - gap between form columns
)

// =============================================================================
// Theme Provider (for future extensibility)
// =============================================================================

// Provider holds dimension configuration that can be customized
type Provider struct {
	mu sync.RWMutex

	// Spacing tokens
	SpacingXS   int
	SpacingS    int
	SpacingM    int
	SpacingL    int
	SpacingXL   int
	SpacingXXL  int

	// Component dimensions (can be customized)
	PanelPadding     int
	ModalPadding     int
	ButtonPadding    int
	InputPadding     int
	etc int // ... more as needed
}

// DefaultProvider returns the default dimension provider
func DefaultProvider() *Provider {
	return &Provider{
		SpacingXS:   XS,
		SpacingS:    S,
		SpacingM:    M,
		SpacingL:    L,
		SpacingXL:   XL,
		SpacingXXL:  XXL,

		PanelPadding: PanelPadding,
		ModalPadding: ModalPadding,
		// ... add more as needed
	}
}

// Global provider instance
var globalProvider = DefaultProvider()

// GetProvider returns the global dimension provider
func GetProvider() *Provider {
	globalProvider.mu.RLock()
	defer globalProvider.mu.RUnlock()
	return globalProvider
}

// SetProvider sets a custom dimension provider
func SetProvider(p *Provider) {
	globalProvider.mu.Lock()
	defer globalProvider.mu.Unlock()
	globalProvider = p
}

// =============================================================================
// Helper Functions
// =============================================================================

// GetSpacing returns the spacing value for a given token
func GetSpacing(token int) int {
	if token <= 0 {
		return 0
	}

	p := GetProvider()

	switch {
	case token <= XS:
		return p.SpacingXS
	case token <= S:
		return p.SpacingS
	case token <= M:
		return p.SpacingM
	case token <= L:
		return p.SpacingL
	case token <= XL:
		return p.SpacingXL
	default:
		return p.SpacingXXL
	}
}
