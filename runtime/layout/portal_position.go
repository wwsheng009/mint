// Package layout provides Portal positioning logic
package layout

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// Portal Positioning System
// =============================================================================
//
// Portal positioning supports three modes:
// 1. Anchor-based: Portal positioned relative to an Anchor element
// 2. Viewport-based (PositionFixed): Portal positioned relative to viewport
// 3. Root-based: Portal positioned relative to PortalRoot (default)
//
// Props:
//   - "position": types.PositionType (Relative/Absolute/Fixed)
//   - "anchor": types.Anchor (AnchorTopLeft/AnchorCenter/etc)
//   - "anchorId": string (ID of the anchor element in the main tree)
//   - "top"/"left"/"right"/"bottom": int (offsets from anchor/viewport)
//
// Example (Tooltip):
//   Portal{
//     Props: {
//       "portalRoot": "main-root",
//       "position": type.PositionAbsolute,
//       "anchor": types.AnchorBottomLeft,
//       "anchorId": "button-123",  // Tooltip positioned below button
//       "top": 5,  // 5px offset below button
//       "left": 10,  // 10px offset from button left
//     },
//   }
//
// Example (Modal centered with PositionFixed):
//   Portal{
//     Props: {
//       "portalRoot": "modal-root",
//       "position": type.PositionFixed,
//       "anchor": types.AnchorCenter,
//       "width": 400,
//       "height": 300,
//     },
//   }

// =============================================================================
// PortalPositionConfig - Portal positioning configuration extracted from Fiber
// =============================================================================

// PortalPositionConfig stores positioning configuration for a Portal
type PortalPositionConfig struct {
	// Position type (Relative/Absolute/Fixed)
	Position types.PositionType

	// Anchor for alignment (when Position is Absolute/Fixed)
	Anchor types.Anchor

	// Anchor element ID (for Anchor-based positioning)
	AnchorID string

	// Offsets (top/left/right/bottom)
	Top    *int
	Left   *int
	Right  *int
	Bottom *int

	// Viewport size (for PositionFixed calculations)
	ViewportWidth  int
	ViewportHeight int

	// Portal size (for centering calculations)
	PortalWidth  int
	PortalHeight int

	// Anchor position (for Anchor-based positioning)
	AnchorX int
	AnchorY int

	// Anchor size (for Anchor-based positioning)
	AnchorWidth  int
	AnchorHeight int
}

// ParsePortalPositionConfig extracts positioning config from a map of props
func ParsePortalPositionConfig(props map[string]interface{}, viewportWidth, viewportHeight, portalWidth, portalHeight int) PortalPositionConfig {
	config := PortalPositionConfig{
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
		PortalWidth:    portalWidth,
		PortalHeight:   portalHeight,
	}

	if props == nil {
		return config
	}

	// Parse position type
	if pos, ok := props["position"].(types.PositionType); ok {
		config.Position = pos
	}

	// Parse anchor
	if anchor, ok := props["anchor"].(types.Anchor); ok {
		config.Anchor = anchor
	}

	// Parse anchor ID
	if anchorID, ok := props["anchorId"].(string); ok {
		config.AnchorID = anchorID
	}

	// Parse offsets
	if top, ok := props["top"].(int); ok {
		config.Top = &top
	}
	if left, ok := props["left"].(int); ok {
		config.Left = &left
	}
	if right, ok := props["right"].(int); ok {
		config.Right = &right
	}
	if bottom, ok := props["bottom"].(int); ok {
		config.Bottom = &bottom
	}

	return config
}

// SetAnchorPosition sets the anchor element position (for Anchor-based positioning)
func (c *PortalPositionConfig) SetAnchorPosition(x, y, width, height int) {
	c.AnchorX = x
	c.AnchorY = y
	c.AnchorWidth = width
	c.AnchorHeight = height
}

// =============================================================================
// PortalPositionCalculator - Calculates Portal position based on config
// =============================================================================

// PortalPositionCalculator calculates the final (x, y) position for a Portal
type PortalPositionCalculator struct{}

// NewPortalPositionCalculator creates a new position calculator
func NewPortalPositionCalculator() *PortalPositionCalculator {
	return &PortalPositionCalculator{}
}

// CalculatePosition computes the final position (x, y) for a Portal
func (pc *PortalPositionCalculator) CalculatePosition(config PortalPositionConfig) (x, y int) {
	switch config.Position {
	case types.PositionFixed:
		return pc.calculateFixedPosition(config)
	case types.PositionAbsolute:
		// Check if using anchor positioning (either via AnchorID or explicit AnchorX/Y)
		if config.AnchorID != "" {
			// Anchor-based positioning with lookup
			return pc.calculateAnchorBasedPosition(config)
		}
		// Also check if anchor position is explicitly set (for testing)
		if config.AnchorX != 0 || config.AnchorY != 0 {
			return pc.calculateAnchorBasedPosition(config)
		}
		// Fall through to default (Root-based)
		fallthrough
	case types.PositionRelative:
		// Default: Root-based positioning (relative to PortalRoot)
		return pc.calculateRootBasedPosition(config)
	default:
		return 0, 0
	}
}

// calculateFixedPosition calculates position for PositionFixed (viewport-relative)
func (pc *PortalPositionCalculator) calculateFixedPosition(config PortalPositionConfig) (x, y int) {
	// Start with viewport coordinates
	vw := config.ViewportWidth
	vh := config.ViewportHeight
	pw := config.PortalWidth
	ph := config.PortalHeight

	// Calculate based on anchor
	switch config.Anchor {
	case types.AnchorTopLeft:
		x = pc.getValue(config.Left, 0)
		y = pc.getValue(config.Top, 0)
	case types.AnchorTop:
		x = (vw - pw) / 2
		if config.Left != nil {
			x = *config.Left
		} else if config.Right != nil {
			x = vw - pw - *config.Right
		}
		y = pc.getValue(config.Top, 0)
	case types.AnchorTopRight:
		x = vw - pw - pc.getValue(config.Right, 0)
		y = pc.getValue(config.Top, 0)
	case types.AnchorLeft:
		x = pc.getValue(config.Left, 0)
		y = (vh - ph) / 2
		if config.Top != nil {
			y = *config.Top
		} else if config.Bottom != nil {
			y = vh - ph - *config.Bottom
		}
	case types.AnchorCenter:
		// Default center
		x = (vw - pw) / 2
		y = (vh - ph) / 2
		// Apply offsets if specified
		if config.Left != nil {
			x = *config.Left
		} else if config.Right != nil {
			x = vw - pw - *config.Right
		}
		if config.Top != nil {
			y = *config.Top
		} else if config.Bottom != nil {
			y = vh - ph - *config.Bottom
		}
	case types.AnchorRight:
		x = vw - pw - pc.getValue(config.Right, 0)
		y = (vh - ph) / 2
		if config.Top != nil {
			y = *config.Top
		} else if config.Bottom != nil {
			y = vh - ph - *config.Bottom
		}
	case types.AnchorBottomLeft:
		x = pc.getValue(config.Left, 0)
		y = vh - ph - pc.getValue(config.Bottom, 0)
	case types.AnchorBottom:
		x = (vw - pw) / 2
		if config.Left != nil {
			x = *config.Left
		} else if config.Right != nil {
			x = vw - pw - *config.Right
		}
		y = vh - ph - pc.getValue(config.Bottom, 0)
	case types.AnchorBottomRight:
		x = vw - pw - pc.getValue(config.Right, 0)
		y = vh - ph - pc.getValue(config.Bottom, 0)
	}

	return x, y
}

// calculateAnchorBasedPosition calculates position relative to an Anchor element
func (pc *PortalPositionCalculator) calculateAnchorBasedPosition(config PortalPositionConfig) (x, y int) {
	ax := config.AnchorX
	ay := config.AnchorY
	aw := config.AnchorWidth
	ah := config.AnchorHeight
	pw := config.PortalWidth
	ph := config.PortalHeight

	// Calculate position based on anchor alignment to anchor element
	switch config.Anchor {
	case types.AnchorTopLeft:
		// Portal's top-left at top-left of anchor
		x = ax + pc.getValue(config.Left, 0)
		y = ay + pc.getValue(config.Top, 0)
	case types.AnchorTop:
		// Portal's top-center at top-center of anchor
		x = ax + (aw-pw)/2 + pc.getValue(config.Left, 0)
		y = ay + pc.getValue(config.Top, 0)
	case types.AnchorTopRight:
		// Portal's top-right at top-right of anchor
		x = ax + aw - pw + pc.getValue(config.Left, 0)
		y = ay + pc.getValue(config.Top, 0)
	case types.AnchorLeft:
		// Portal's center-left at left-center of anchor
		x = ax + pc.getValue(config.Left, 0)
		y = ay + (ah-ph)/2 + pc.getValue(config.Top, 0)
	case types.AnchorCenter:
		// Portal's center at center of anchor
		x = ax + (aw-pw)/2 + pc.getValue(config.Left, 0)
		y = ay + (ah-ph)/2 + pc.getValue(config.Top, 0)
	case types.AnchorRight:
		// Portal's center-right at right-center of anchor
		// Support both left and right offsets (right takes precedence)
		if config.Left != nil {
			x = ax + aw - pw + *config.Left
		} else if config.Right != nil {
			// right offset moves the portal to the left
			x = ax + aw - pw - *config.Right
		} else {
			x = ax + aw - pw
		}
		y = ay + (ah-ph)/2 + pc.getValue(config.Top, 0)
	case types.AnchorBottomLeft:
		// Portal's bottom-left at bottom-left of anchor
		x = ax + pc.getValue(config.Left, 0)
		y = ay + ah - ph + pc.getValue(config.Top, 0)
	case types.AnchorBottom:
		// Portal's bottom-center at bottom-center of anchor
		x = ax + (aw-pw)/2 + pc.getValue(config.Left, 0)
		y = ay + ah - ph + pc.getValue(config.Top, 0)
	case types.AnchorBottomRight:
		// Portal's bottom-right at bottom-right of anchor
		x = ax + aw - pw + pc.getValue(config.Left, 0)
		y = ay + ah - ph + pc.getValue(config.Top, 0)
	}

	return x, y
}

// calculateRootBasedPosition calculates position relative to PortalRoot (default)
func (pc *PortalPositionCalculator) calculateRootBasedPosition(config PortalPositionConfig) (x, y int) {
	// Default offset from PortalRoot
	x = pc.getValue(config.Left, 0)
	y = pc.getValue(config.Top, 0)

	return x, y
}

// getValue returns the value from pointer or default
func (pc *PortalPositionCalculator) getValue(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// =============================================================================
// Helper Functions
// =============================================================================

// FindAnchorPosition finds the position (x, y, width, height) of an anchor element
// in the layout tree by its ID
func FindAnchorPosition(root *LayoutBox, anchorID string) (x, y, width, height int, found bool) {
	if root == nil || anchorID == "" {
		return 0, 0, 0, 0, false
	}

	var result struct {
		X, Y          int
		Width, Height int
		Found         bool
	}

	var search func(box *LayoutBox)
	search = func(box *LayoutBox) {
		if box == nil || result.Found {
			return
		}

		// Check if this is the anchor element
		if box.ID == anchorID {
			result.X = box.X
			result.Y = box.Y
			result.Width = box.Width
			result.Height = box.Height
			result.Found = true
			return
		}

		// Search children
		for _, child := range box.Children {
			search(child)
			if result.Found {
				return
			}
		}
	}

	search(root)

	return result.X, result.Y, result.Width, result.Height, result.Found
}

// =============================================================================
// String Methods for Debugging
// =============================================================================

// String returns a string representation of PortalPositionConfig
func (c *PortalPositionConfig) String() string {
	left := "nil"
	right := "nil"
	top := "nil"
	bottom := "nil"
	if c.Left != nil {
		left = fmt.Sprintf("%d", *c.Left)
	}
	if c.Right != nil {
		right = fmt.Sprintf("%d", *c.Right)
	}
	if c.Top != nil {
		top = fmt.Sprintf("%d", *c.Top)
	}
	if c.Bottom != nil {
		bottom = fmt.Sprintf("%d", *c.Bottom)
	}

	return fmt.Sprintf("PortalPositionConfig{Position:%s, Anchor:%s, AnchorID:%s, Left:%s, Right:%s, Top:%s, Bottom:%s, Viewport:%dx%d, Portal:%dx%d, Anchor@(%d,%d-%dx%d)}",
		c.Position, c.Anchor, c.AnchorID, left, right, top, bottom,
		c.ViewportWidth, c.ViewportHeight,
		c.PortalWidth, c.PortalHeight,
		c.AnchorX, c.AnchorY, c.AnchorWidth, c.AnchorHeight)
}
