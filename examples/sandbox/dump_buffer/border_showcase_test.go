// border_showcase.go - Comprehensive border rendering showcase
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Border Style Variations
// =============================================================================

// Border 1: Single style (default) - continuous line characters
func ShowcaseSingleStyle() ui.VNode {
	return ui.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{
		ui.Text("Single Border"),
	})
}

// Border 2: Single style with label
func ShowcaseSingleWithLabel() ui.VNode {
	return ui.NewVStack().SingleBorder("Title").SetChildrenList([]ui.VNode{
		ui.Text("Single with Label"),
	})
}

// Border 3: Double style border
func ShowcaseDoubleStyle() ui.VNode {
	return ui.NewVStack().DoubleBorder().SetChildrenList([]ui.VNode{
		ui.Text("Double Border"),
	})
}

// Border 4: Double style with label
func ShowcaseDoubleWithLabel() ui.VNode {
	return ui.NewVStack().DoubleBorder("Settings").SetChildrenList([]ui.VNode{
		ui.Text("Double with Label"),
	})
}

// Border 5: Rounded corners style
func ShowcaseRoundedStyle() ui.VNode {
	return ui.NewVStack().RoundedBorder().SetChildrenList([]ui.VNode{
		ui.Text("Rounded Corners"),
	})
}

// Border 6: Rounded style with label
func ShowcaseRoundedWithLabel() ui.VNode {
	return ui.NewVStack().RoundedBorder("Info").SetChildrenList([]ui.VNode{
		ui.Text("Rounded with Label"),
	})
}

// Border 7: Dashed (ASCII) style
func ShowcaseDashedStyle() ui.VNode {
	return ui.NewVStack().DashedBorder().SetChildrenList([]ui.VNode{
		ui.Text("Dashed Border"),
	})
}

// Border 8: Dashed style with label
func ShowcaseDashedWithLabel() ui.VNode {
	return ui.NewVStack().DashedBorder("ASCII").SetChildrenList([]ui.VNode{
		ui.Text("Dashed with Label"),
	})
}

// =============================================================================
// Multi-line Content
// =============================================================================

func ShowcaseMultiLineContent() ui.VNode {
	return ui.NewVStack().SingleBorder("Multiple Lines").SetChildrenList([]ui.VNode{
		ui.VStack(
			ui.Text("Line 1: First content"),
			ui.Text("Line 2: Second content"),
			ui.Text("Line 3: Third content"),
		),
	})
}

func ShowcaseMultiLineDouble() ui.VNode {
	return ui.NewVStack().DoubleBorder("Data Grid").SetChildrenList([]ui.VNode{
		ui.VStack(
			ui.Text("┌─────────┬─────────┐"),
			ui.Text("│ Column1 │ Column2 │"),
			ui.Text("├─────────┼─────────┤"),
			ui.Text("│ Data A  │ Data B  │"),
			ui.Text("└─────────┴─────────┘"),
		),
	})
}

// =============================================================================
// Wide Character Content
// =============================================================================

func ShowcaseWideCharacters() ui.VNode {
	return ui.NewVStack().SingleBorder("Wide Characters").SetChildrenList([]ui.VNode{
		ui.VStack(
			ui.Text("English: Hello"),
			ui.Text("Chinese: 你好世界"),
			ui.Text("Japanese: こんにちは"),
			ui.Text("Emoji: 😀🎉🚀"),
		),
	})
}

// =============================================================================
// Nested Borders
// =============================================================================

func ShowcaseNestedBorders() ui.VNode {
	return ui.NewVStack().SingleBorder("Outer Border").SetChildrenList([]ui.VNode{
		ui.VStack(
			ui.Text("Content above nested"),
			ui.NewVStack().DoubleBorder("Inner").SetChildrenList([]ui.VNode{
				ui.Text("Nested content"),
			}),
			ui.Text("Content below nested"),
		),
	})
}

func ShowcaseTripleNested() ui.VNode {
	return ui.NewVStack().SingleBorder("Level 1").SetChildrenList([]ui.VNode{
		ui.NewVStack().DoubleBorder("Level 2").SetChildrenList([]ui.VNode{
			ui.NewVStack().RoundedBorder("Level 3").SetChildrenList([]ui.VNode{
				ui.Text("Deeply nested content"),
			}),
		}),
	})
}

// =============================================================================
// All Styles Grid
// =============================================================================

func ShowcaseAllStylesGrid() ui.VNode {
	return ui.VStack(
		ui.Text("Border Style Showcase"),
		ui.Text(""),
		ui.HStack(
			ui.NewVStack().SingleBorder("single").SetChildrenList([]ui.VNode{ui.Text("Style 1")}),
			ui.Text("  "),
			ui.NewVStack().DoubleBorder("double").SetChildrenList([]ui.VNode{ui.Text("Style 2")}),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewVStack().RoundedBorder("rounded").SetChildrenList([]ui.VNode{ui.Text("Style 3")}),
			ui.Text("  "),
			ui.NewVStack().DashedBorder("dashed").SetChildrenList([]ui.VNode{ui.Text("Style 4")}),
		),
	)
}

// =============================================================================
// Color Variations
// =============================================================================

func ShowcaseBorderColors() ui.VNode {
	return ui.VStack(
		ui.Text("Border Color Variations"),
		ui.Text(""),
		ui.HStack(
			ui.NewVStack().SingleBorder("red").SetChildrenList([]ui.VNode{ui.Text("Red")}),
			ui.Text(" "),
			ui.NewVStack().SingleBorder("green").SetChildrenList([]ui.VNode{ui.Text("Green")}),
			ui.Text(" "),
			ui.NewVStack().SingleBorder("blue").SetChildrenList([]ui.VNode{ui.Text("Blue")}),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewVStack().SingleBorder("yellow").SetChildrenList([]ui.VNode{ui.Text("Yellow")}),
			ui.Text(" "),
			ui.NewVStack().SingleBorder("magenta").SetChildrenList([]ui.VNode{ui.Text("Magenta")}),
			ui.Text(" "),
			ui.NewVStack().SingleBorder("cyan").SetChildrenList([]ui.VNode{ui.Text("Cyan")}),
		),
	)
}

// =============================================================================
// Tests
// =============================================================================

func TestShowcaseSingleStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseSingleStyle,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Single Style Border"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Single Style Border ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseSingleWithLabel(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseSingleWithLabel,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Single Border with Label"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Single Border with Label ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseDoubleStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseDoubleStyle,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Double Style Border"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Double Style Border ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseDoubleWithLabel(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseDoubleWithLabel,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Double Border with Label"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Double Border with Label ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseRoundedStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseRoundedStyle,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Rounded Style Border"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Rounded Style Border ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseRoundedWithLabel(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseRoundedWithLabel,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Rounded Border with Label"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Rounded Border with Label ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseDashedStyle(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseDashedStyle,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Dashed Style Border"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Dashed Style Border ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseDashedWithLabel(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseDashedWithLabel,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Dashed Border with Label"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Dashed Border with Label ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseMultiLineContent(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseMultiLineContent,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Multi-line Content"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Multi-line Content ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseMultiLineDouble(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseMultiLineDouble,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Multi-line Double Border"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Multi-line Double Border ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseWideCharacters(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseWideCharacters,
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Wide Characters"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Wide Characters ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseNestedBorders(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseNestedBorders,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Nested Borders"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Nested Borders ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseTripleNested(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseTripleNested,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Triple Nested Borders"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Triple Nested Borders ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseAllStylesGrid(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseAllStylesGrid,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("All Styles Grid"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== All Styles Grid ===")
	t.Log(output)
	t.Log("=== End ===")
}

func TestShowcaseBorderColors(t *testing.T) {
	app, err := ui.RunTestWithSandbox(ShowcaseBorderColors,
		ui.WithWidth(60),
		ui.WithHeight(20),
		ui.WithTitle("Border Colors"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)
	output := app.GetRenderString()
	t.Log("=== Border Colors ===")
	t.Log(output)
	t.Log("=== End ===")
}

// Master test that runs all showcases
func TestBorderShowcaseAll(t *testing.T) {
	tests := []struct {
		name string
		fn   func() ui.VNode
		w, h int
	}{
		{"Single Style", ShowcaseSingleStyle, 30, 10},
		{"Single + Label", ShowcaseSingleWithLabel, 30, 10},
		{"Double Style", ShowcaseDoubleStyle, 30, 10},
		{"Double + Label", ShowcaseDoubleWithLabel, 30, 10},
		{"Rounded Style", ShowcaseRoundedStyle, 30, 10},
		{"Rounded + Label", ShowcaseRoundedWithLabel, 30, 10},
		{"Dashed Style", ShowcaseDashedStyle, 30, 10},
		{"Dashed + Label", ShowcaseDashedWithLabel, 30, 10},
		{"Multi-line Content", ShowcaseMultiLineContent, 40, 15},
		{"Multi-line Double", ShowcaseMultiLineDouble, 40, 15},
		{"Wide Characters", ShowcaseWideCharacters, 40, 15},
		{"Nested Borders", ShowcaseNestedBorders, 50, 20},
		{"Triple Nested", ShowcaseTripleNested, 50, 20},
		{"All Styles Grid", ShowcaseAllStylesGrid, 50, 20},
		{"Border Colors", ShowcaseBorderColors, 60, 20},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app, err := ui.RunTestWithSandbox(tc.fn,
				ui.WithWidth(tc.w),
				ui.WithHeight(tc.h),
				ui.WithTitle(tc.name),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer app.Close()

			time.Sleep(100 * time.Millisecond)
			output := app.GetRenderString()
			t.Logf("=== %s ===", tc.name)
			t.Log(output)
			t.Log("=== End ===")
		})
	}
}
