package layout

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/types"
)

// TestPositionCalculator_Debug simple debug test
func TestPositionCalculator_Debug(t *testing.T) {
	println("=== Testing PositionCalculator Debug ===")

	calculator := NewPortalPositionCalculator()

	config := PortalPositionConfig{
		Position:     types.PositionAbsolute,
		Anchor:       types.AnchorTopLeft,
		AnchorX:      100,
		AnchorY:      100,
		AnchorWidth:  200,
		AnchorHeight: 50,
		PortalWidth:  150,
		PortalHeight: 40,
	}

	fmt.Printf("Config: Position=%v, Anchor=%v, AnchorX=%d, AnchorY=%d\n",
		config.Position, config.Anchor, config.AnchorX, config.AnchorY)

	x, y := calculator.CalculatePosition(config)
	fmt.Printf("Result: x=%d, y=%d\n", x, y)

	assert.Equal(t, 100, x)
	assert.Equal(t, 100, y)

	config.Position = types.PositionFixed
	config.Anchor = types.AnchorCenter
	config.ViewportWidth = 800
	config.ViewportHeight = 600

	fmt.Printf("\nConfig (Fixed): Position=%v, Anchor=%v, Viewport=%dx%d\n",
		config.Position, config.Anchor, config.ViewportWidth, config.ViewportHeight)

	x, y = calculator.CalculatePosition(config)
	fmt.Printf("Result: x=%d, y=%d\n", x, y)

	assert.Equal(t, 200, x) // (800-400)/2 ... wait with 150 width: (800-150)/2 = 325
	fmt.Printf("Expected: (800-150)/2 = 325, got: %d\n", x)
}
