// Fiber-first Absolute Component Demo
// Demonstrates absolute positioning following the Fiber-first architecture
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/absolute"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Absolute positioned elements using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: TopLeft - 在容器内展示左上角定位
			// =====================================================
			sectionTitle("═══ 1. TopLeft (0,0) ═══"),

			// 用 border 包裹来显示容器边界 (border 需要显式高度)
			newborder.New().
				SetBorderLabel("TopLeft").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[TL] 定位在 (0,0)").Foreground(theme.Highlight()),
					).
						Left(absolute.AbsolutePos(0)).
						Top(absolute.AbsolutePos(0)).
						Build(),
				}),

			// =====================================================
			// Section 2: TopRight - 右上角定位
			// =====================================================
			sectionTitle("═══ 2. TopRight ═══"),

			newborder.New().
				SetBorderLabel("TopRight").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[TR] 右上角").Foreground(theme.Success()),
					).
						Right(absolute.AbsolutePos(0)).
						Top(absolute.AbsolutePos(0)).
						Build(),
				}),

			// =====================================================
			// Section 3: BottomLeft - 左下角定位
			// =====================================================
			sectionTitle("═══ 3. BottomLeft ═══"),

			newborder.New().
				SetBorderLabel("BottomLeft").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[BL] 左下角").Foreground(theme.Warning()),
					).
						Left(absolute.AbsolutePos(0)).
						Bottom(absolute.AbsolutePos(0)).
						Build(),
				}),

			// =====================================================
			// Section 4: BottomRight - 右下角定位
			// =====================================================
			sectionTitle("═══ 4. BottomRight ═══"),

			newborder.New().
				SetBorderLabel("BottomRight").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[BR] 右下角").Foreground(theme.Error()),
					).
						Right(absolute.AbsolutePos(0)).
						Bottom(absolute.AbsolutePos(0)).
						Build(),
				}),

			// =====================================================
			// Section 5: Center - 居中定位
			// =====================================================
			sectionTitle("═══ 5. Center (50%,50%) ═══"),

			newborder.New().
				SetBorderLabel("Center").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[CENTER] 居中").Foreground(theme.Highlight()),
					).
						Left(absolute.RelativePos(50)).
						Top(absolute.RelativePos(50)).
						Anchor(absolute.AnchorCenter).
						Build(),
				}),

			// =====================================================
			// Section 6: 25% 位置
			// =====================================================
			sectionTitle("═══ 6. AtPercent(25,25) ═══"),

			newborder.New().
				SetBorderLabel("25%").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[25%] 四分之一位置").Foreground(theme.Success()),
					).
						Left(absolute.RelativePos(25)).
						Top(absolute.RelativePos(25)).
						Build(),
				}),

			// =====================================================
			// Section 7: 75% 位置
			// =====================================================
			sectionTitle("═══ 7. AtPercent(75,75) ═══"),

			newborder.New().
				SetBorderLabel("75%").
				SetWidth(58).SetHeight(8).
				SetChildren([]rtui.VNode{
					absolute.NewBuilder(
						newtext.New("[75%] 四分之三位置").Foreground(theme.Warning()),
					).
						Left(absolute.RelativePos(75)).
						Top(absolute.RelativePos(75)).
						Build(),
				}),

			// =====================================================
			// Footer
			// =====================================================
			sectionTitle("════════════════════════════════════════════════════════"),
			newtext.New("Absolute 定位演示完成").Foreground(theme.Muted()),
		})
}

// =============================================================================
// Helper Functions
// =============================================================================

func sectionTitle(text string) rtui.VNode {
	return newtext.New(text).Foreground(theme.Highlight())
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Absolute Rendering Demo                      ║")
	fmt.Println("║   (Absolute positioning component)                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// Enable debug logging
	os.Setenv("MINT_DEBUG_LAYOUT", "1")

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(60, 75)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 75},
		AvailableWidth:  60,
		AvailableHeight: 75,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering Absolute positioned elements...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	node.Paint(ctx, buf)
	utils.PrintBuffer(buf, 60, 75)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Absolute Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	printFeatures()

	os.Exit(0)
}

func printFeatures() {
	features := []string{
		"  - AbsolutePos: 固定位置 (单位: 字符)",
		"  - RelativePos: 百分比位置 (0-100)",
		"  - Left/Top/Right/Bottom: 位置偏移",
		"  - Anchor: 对齐锚点",
		"  - Width/Height: 显式尺寸",
		"",
		"Anchors:",
		"  - AnchorTopLeft (默认)",
		"  - AnchorTop, AnchorTopRight",
		"  - AnchorLeft, AnchorCenter, AnchorRight",
		"  - AnchorBottomLeft, AnchorBottom, AnchorBottomRight",
	}

	for _, f := range features {
		fmt.Println(f)
	}
	fmt.Println(strings.Repeat("=", 60))
}
