// border_output_test.go - Save border rendering results to files
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
)

// outputDir is where we save the rendered output files
const outputDir = "./border_outputs"

// TestSaveAllBordersToFile saves all border styles to files for viewing
func TestSaveAllBordersToFile(t *testing.T) {
	// Create output directory
	os.MkdirAll(outputDir, 0755)

	testCases := []struct {
		name      string
		filename  string
		buildVNode func() ui.VNode
		width     int
		height    int
	}{
		{
			name:     "Single Style",
			filename: "01_single.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{ui.Text("Single Border")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Single with Label",
			filename: "02_single_label.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().SingleBorder("Title").SetChildrenList([]ui.VNode{ui.Text("Content")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Double Style",
			filename: "03_double.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().DoubleBorder().SetChildrenList([]ui.VNode{ui.Text("Double Border")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Double with Label",
			filename: "04_double_label.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().DoubleBorder("Settings").SetChildrenList([]ui.VNode{ui.Text("Content")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Rounded Style",
			filename: "05_rounded.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().RoundedBorder().SetChildrenList([]ui.VNode{ui.Text("Rounded Border")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Rounded with Label",
			filename: "06_rounded_label.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().RoundedBorder("Info").SetChildrenList([]ui.VNode{ui.Text("Content")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Dashed Style",
			filename: "07_dashed.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().DashedBorder().SetChildrenList([]ui.VNode{ui.Text("Dashed Border")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Dashed with Label",
			filename: "08_dashed_label.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().DashedBorder("ASCII").SetChildrenList([]ui.VNode{ui.Text("Content")})
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Multi-line Content",
			filename: "09_multiline.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().SingleBorder("Multiple Lines").SetChildrenList([]ui.VNode{
					ui.VStack(
						ui.Text("Line 1: First content"),
						ui.Text("Line 2: Second content"),
						ui.Text("Line 3: Third content"),
					),
				})
			},
			width:  40,
			height: 15,
		},
		{
			name:     "Wide Characters",
			filename: "10_wide_chars.txt",
			buildVNode: func() ui.VNode {
				return ui.NewVStack().SingleBorder("Wide Characters").SetChildrenList([]ui.VNode{
					ui.VStack(
						ui.Text("English: Hello"),
						ui.Text("Chinese: 你好世界"),
						ui.Text("Japanese: こんにちは"),
						ui.Text("Emoji: 😀🎉🚀"),
					),
				})
			},
			width:  40,
			height: 15,
		},
		{
			name:     "Nested Borders (ui.VStack)",
			filename: "11_nested.txt",
			buildVNode: func() ui.VNode {
				return ui.VStackBuilder(
					ui.Text("Above nested"),
					ui.VStackBuilder(
						ui.Text("Nested content"),
					).DoubleBorder("Inner").Build(),
					ui.Text("Below nested"),
				).SingleBorder("Outer").Build()
			},
			width:  50,
			height: 20,
		},
		{
			name:     "All Styles Grid",
			filename: "12_all_styles.txt",
			buildVNode: func() ui.VNode {
				return ui.VStack(
					ui.Text("Border Style Showcase"),
					ui.Text(""),
					ui.HStack(
						ui.NewVStack().SingleBorder("single").SetChildrenList([]ui.VNode{ui.Text("A")}),
						ui.Text("  "),
						ui.NewVStack().DoubleBorder("double").SetChildrenList([]ui.VNode{ui.Text("B")}),
						ui.Text("  "),
						ui.NewVStack().RoundedBorder("rounded").SetChildrenList([]ui.VNode{ui.Text("C")}),
						ui.Text("  "),
						ui.NewVStack().DashedBorder("dashed").SetChildrenList([]ui.VNode{ui.Text("D")}),
					),
				)
			},
			width:  60,
			height: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app, err := ui.RunTestWithSandbox(tc.buildVNode,
				ui.WithWidth(tc.width),
				ui.WithHeight(tc.height),
				ui.WithTitle(tc.name),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer app.Close()

			time.Sleep(100 * time.Millisecond)

			// Save to file
			outputPath := filepath.Join(outputDir, tc.filename)
			err = app.SaveBufferToFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to save to %s: %v", outputPath, err)
			}

			// Also print to console
			output := app.GetRenderString()
			t.Logf("=== %s ===\n%s\n=== End ===\n", tc.name, output)
			t.Logf("Saved to: %s\n", outputPath)
		})
	}

	t.Logf("\n=== All outputs saved to: %s ===\n", outputDir)
}

// TestViewSingleBorder saves and displays a single border for quick viewing
func TestViewSingleBorder(t *testing.T) {
	os.MkdirAll(outputDir, 0755)

	app, err := ui.RunTestWithSandbox(
		func() ui.VNode {
			return ui.NewVStack().SingleBorder("Demo").SetChildrenList([]ui.VNode{
				ui.VStack(
					ui.Text("This is a demo"),
					ui.Text("with multiple lines"),
					ui.Text("showing the border"),
				),
			})
		},
		ui.WithWidth(40),
		ui.WithHeight(15),
		ui.WithTitle("Border Demo"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)

	// Save to file
	outputPath := filepath.Join(outputDir, "demo.txt")
	err = app.SaveBufferToFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	// Print to console
	app.DumpBuffer()
	t.Logf("\nOutput saved to: %s\n", outputPath)
}

// TestAllStylesGrid runs the "All Styles Grid" test case independently
// This test verifies that HStack with multiple Bordered elements renders correctly
// without overlapping or truncation of labels
func TestAllStylesGrid(t *testing.T) {
	os.MkdirAll(outputDir, 0755)
	os.Setenv("TUI_DEBUG_ALL", "true")
	os.Setenv("TUI_LOG_OUTPUT", "both")
	vnode := ui.VStack(
		ui.Text("Border Style Showcase"),
		ui.Text(""),
		ui.HStack(
			ui.NewVStack().SingleBorder("single").SetChildrenList([]ui.VNode{ui.Text("A")}),
			ui.Text("  "),
			ui.NewVStack().DoubleBorder("double").SetChildrenList([]ui.VNode{ui.Text("B")}),
			ui.Text("  "),
			ui.NewVStack().RoundedBorder("rounded").SetChildrenList([]ui.VNode{ui.Text("C")}),
			ui.Text("  "),
			ui.NewVStack().DashedBorder("dashed").SetChildrenList([]ui.VNode{ui.Text("D")}),
		),
	)

	app, err := ui.RunTestWithSandbox(
		func() ui.VNode { return vnode },
		ui.WithWidth(60),
		ui.WithHeight(20),
		ui.WithTitle("All Styles Grid"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	time.Sleep(100 * time.Millisecond)

	// Save to file
	outputPath := filepath.Join(outputDir, "12_all_styles.txt")
	err = app.SaveBufferToFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	// Print to console
	output := app.GetRenderString()
	t.Logf("=== All Styles Grid ===\n%s\n=== End ===\n", output)
	t.Logf("Output saved to: %s\n", outputPath)

	// Verify output contains expected labels (not truncated)
	if !strings.Contains(output, "single") {
		t.Error("Expected output to contain 'single' label (might be truncated)")
	}
	if !strings.Contains(output, "double") {
		t.Error("Expected output to contain 'double' label (might be truncated)")
	}
	if !strings.Contains(output, "rounded") {
		t.Error("Expected output to contain 'rounded' label (might be truncated)")
	}
	if !strings.Contains(output, "dashed") {
		t.Error("Expected output to contain 'dashed' label (might be truncated)")
	}
}
