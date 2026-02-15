// border_output_test.go - Save border rendering results to files
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/internal/log"
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
				return ui.Bordered().Child(ui.Text("Single Border")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Single with Label",
			filename: "02_single_label.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Label("Title").Child(ui.Text("Content")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Double Style",
			filename: "03_double.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("double").Child(ui.Text("Double Border")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Double with Label",
			filename: "04_double_label.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("double").Label("Settings").Child(ui.Text("Content")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Rounded Style",
			filename: "05_rounded.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("rounded").Child(ui.Text("Rounded Border")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Rounded with Label",
			filename: "06_rounded_label.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("rounded").Label("Info").Child(ui.Text("Content")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Dashed Style",
			filename: "07_dashed.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("dashed").Child(ui.Text("Dashed Border")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Dashed with Label",
			filename: "08_dashed_label.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Style("dashed").Label("ASCII").Child(ui.Text("Content")).Build()
			},
			width:  30,
			height: 10,
		},
		{
			name:     "Multi-line Content",
			filename: "09_multiline.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Label("Multiple Lines").Child(
					ui.VStack(
						ui.Text("Line 1: First content"),
						ui.Text("Line 2: Second content"),
						ui.Text("Line 3: Third content"),
					),
				).Build()
			},
			width:  40,
			height: 15,
		},
		{
			name:     "Wide Characters",
			filename: "10_wide_chars.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Label("Wide Characters").Child(
					ui.VStack(
						ui.Text("English: Hello"),
						ui.Text("Chinese: 你好世界"),
						ui.Text("Japanese: こんにちは"),
						ui.Text("Emoji: 😀🎉🚀"),
					),
				).Build()
			},
			width:  40,
			height: 15,
		},
		{
			name:     "Nested Borders",
			filename: "11_nested.txt",
			buildVNode: func() ui.VNode {
				return ui.Bordered().Label("Outer").Child(
					ui.VStack(
						ui.Text("Above nested"),
						ui.Bordered().Style("double").Label("Inner").Child(
							ui.Text("Nested content"),
						).Build(),
						ui.Text("Below nested"),
					),
				).Build()
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
						ui.Bordered().Label("single").Child(ui.Text("A")).Build(),
						ui.Text("  "),
						ui.Bordered().Style("double").Label("double").Child(ui.Text("B")).Build(),
						ui.Text("  "),
						ui.Bordered().Style("rounded").Label("rounded").Child(ui.Text("C")).Build(),
						ui.Text("  "),
						ui.Bordered().Style("dashed").Label("dashed").Child(ui.Text("D")).Build(),
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

	// Enable debug logging
	log.PaintLogger.SetEnabled(true)

	app, err := ui.RunTestWithSandbox(
		func() ui.VNode {
			return ui.Bordered().Label("Demo").Child(
				ui.VStack(
					ui.Text("This is a demo"),
					ui.Text("with multiple lines"),
					ui.Text("showing the border"),
				),
			).Build()
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
