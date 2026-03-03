// Test Layout Engine Fixed Positioning - 直接测试布局引擎的 Fixed 定位计算
package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/types"
	"github.com/wwsheng009/mint/runtime/layout"
)

// MockFiberNode 模拟一个带 PositionProvider 接口的节点
type MockFiberNode struct {
	id       string
	width    int
	height   int
	position types.PositionType
	anchor   types.Anchor
	layer    layout.Layer
}

func (m *MockFiberNode) ID() string      { return m.id }
func (m *MockFiberNode) Type() string     { return "mock" }
func (m *MockFiberNode) Children() []layout.Node { return nil }
func (m *MockFiberNode) GetPosition() (int, int) { return 0, 0 }
func (m *MockFiberNode) SetPosition(x, y int) {}
func (m *MockFiberNode) GetSize() (int, int) { return m.width, m.height }
func (m *MockFiberNode) SetSize(w, h int) {}
func (m *MockFiberNode) GetWidth() int { return m.width }
func (m *MockFiberNode) GetHeight() int { return m.height }

// Measurable 接口
func (m *MockFiberNode) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{Width: m.width, Height: m.height}
}

// PositionProvider 接口
func (m *MockFiberNode) GetPositionType() layout.PositionType {
	return layout.PositionType(m.position)
}

func (m *MockFiberNode) GetAnchor() layout.Anchor {
	return layout.Anchor(m.anchor)
}

// Layered 接口
func (m *MockFiberNode) GetLayer() layout.Layer {
	return m.layer
}

func (m *MockFiberNode) GetZIndex() int {
	return 100
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Test: Layout Engine Fixed Positioning 直接测试                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	viewportW := 80
	viewportH := 45
	modalW := 38
	modalH := 12

	// 创建 PositionFixed + AnchorCenter 的 Modal 节点
	fixedCenteredModal := &MockFiberNode{
		id:       "modal-fixed-centered",
		width:    modalW,
		height:   modalH,
		position: types.PositionFixed,
		anchor:   types.AnchorCenter,
		layer:    layout.LayerModal,
	}

	fmt.Printf("\n%s 测试配置 %s\n", strings.Repeat("=", 54), strings.Repeat("=", 54))
	fmt.Printf("\nViewport 尺寸: %dx%d\n", viewportW, viewportH)
	fmt.Printf("Modal 尺寸:   %dx%d\n", modalW, modalH)
	fmt.Printf("定位模式:     Position=fixed, Anchor=center\n")

	expectedX := (viewportW - modalW) / 2
	expectedY := (viewportH - modalH) / 2
	fmt.Printf("预期位置:     AbsX=%d, AbsY=%d\n", expectedX, expectedY)

	// 模拟布局引擎计算（查看 types.go 中的 Fixed 定位逻辑）
	fmt.Printf("\n%s 布局计算逻辑 (types.go) %s\n", strings.Repeat("=", 57), strings.Repeat("=", 57))
	fmt.Println("\nif position == PositionFixed:")
	fmt.Printf("  rootW := constraints.MaxWidth  // = %d\n", viewportW)
	fmt.Printf("  rootH := constraints.MaxHeight // = %d\n", viewportH)
	fmt.Printf("  switch AnchorCenter:")
	fmt.Printf("    x = (rootW - width) / 2  // = (%d - %d) / 2 = %d\n", viewportW, modalW, expectedX)
	fmt.Printf("    y = (rootH - height) / 2 // = (%d - %d) / 2 = %d\n", viewportH, modalH, expectedY)

	// 实际运行布局引擎
	fmt.Printf("\n%s 运行布局引擎 %s\n", strings.Repeat("=", 53), strings.Repeat("=", 53))
	
	engine := layout.NewEngine()
	result := engine.Layout(fixedCenteredModal, layout.Constraints{
		MinWidth:  0,
		MaxWidth:  viewportW,
		MinHeight: 0,
		MaxHeight: viewportH,
	})

	if result.Root == nil {
		fmt.Println("\n❌ 布局结果为 nil")
		return
	}

	fmt.Printf("\n=== 布局结果 ===\n")
	fmt.Printf("Position:    X=%d, Y=%d\n", result.Root.X, result.Root.Y)
	fmt.Printf("Absolute:    AbsX=%d, AbsY=%d\n", result.Root.AbsX, result.Root.AbsY)
	fmt.Printf("Size:        %dx%d\n", result.Root.Width, result.Root.Height)
	fmt.Printf("Layer:       %s (%d)\n", getLayerName(result.Root.Layer), result.Root.Layer)

	isCorrect := (result.Root.AbsX == expectedX) && (result.Root.AbsY == expectedY)
	if isCorrect {
		fmt.Printf("\n✅ 布局引擎 Fixed 定位计算正确！Modal 会居中显示\n")
	} else {
		fmt.Printf("\n❌ 布局引擎固定定位计算错误！\n")
		fmt.Printf("   期望: AbsX=%d, AbsY=%d\n", expectedX, expectedY)
		fmt.Printf("   实际: AbsX=%d, AbsY=%d\n", result.Root.AbsX, result.Root.AbsY)
		fmt.Printf("   偏差: ΔX=%d, ΔY=%d\n", result.Root.AbsX-expectedX, result.Root.AbsY-expectedY)
	}
}

func getLayerName(l layout.Layer) string {
	switch l {
	case layout.LayerBase:
		return "LayerBase"
	case layout.LayerOverlay:
		return "LayerOverlay"
	case layout.LayerModal:
		return "LayerModal"
	case layout.LayerTooltip:
		return "LayerTooltip"
	default:
		return fmt.Sprintf("Unknown(%d)", int(l))
	}
}
