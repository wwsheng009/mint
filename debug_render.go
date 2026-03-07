package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== Debug Render Output ===\n")

	// Create a renderer with double buffering
	renderer := paint.NewRenderer(40, 10)

	// Frame 1: Draw with arrow
	fmt.Println("Frame 1: 'Button →' at (5, 3)")
	backBuf := renderer.GetBackBuffer()
	fmt.Printf("  backBuf address: %p\n", backBuf)
	backBuf.SetString(5, 3, "Button →", style.Style{})

	fmt.Printf("  Back buffer before render: [5-14]\n")
	printCells(backBuf, 3, 5, 14)

	frame1Output := renderer.Render()
	fmt.Printf("  Output: %q\n", frame1Output)

	// Frame 2: Check what back buffer points to
	fmt.Println("\nFrame 2: 'Button-' at (5, 3)")
	backBuf2 := renderer.GetBackBuffer()
	fmt.Printf("  backBuf2 address: %p\n", backBuf2)
	fmt.Printf("  Is same address? %v\n", backBuf == backBuf2)

	// Print back buffer before SetString
	fmt.Printf("  Back buffer before SetString: [5-14]\n")
	printCells(backBuf2, 3, 5, 14)

	backBuf2.SetString(5, 3, "Button-", style.Style{})

	fmt.Printf("  Back buffer after SetString: [5-14]\n")
	printCells(backBuf2, 3, 5, 14)

	// Check front buffer
	frontBuf := renderer.GetFrontBuffer()
	fmt.Printf("  Front buffer before render: [5-14]\n")
	printCells(frontBuf, 3, 5, 14)

	frame2Output := renderer.Render()
	fmt.Printf("  Output: %q\n", frame2Output)

	fmt.Printf("\nFront buffer after render: [5-14]\n")
	printCells(frontBuf, 3, 5, 14)

	// Analyze output
	if strings.Contains(frame2Output, "\x1b[4;12H") {
		fmt.Println("  ⚠️  Output includes cursor move to position 12")
	}
	if strings.Contains(frame2Output, "  ") && strings.Count(frame2Output, "  ") > 1 {
		fmt.Println("  ⚠️  Output has multiple spaces")
	} else if strings.Contains(frame2Output, " ") && strings.Count(frame2Output, " ") == 1 {
		spaceIdx := strings.Index(frame2Output, " ")
		fmt.Printf("  ⚠️  Output has 1 space at index %d\n", spaceIdx)
	}
}

func printCells(buf *paint.Buffer, y, startX, endX int) {
	for x := startX; x < endX; x++ {
		cell := buf.GetContent(x, y)
		cluster := cell.Cluster
		if cluster == "" {
			if cell.IsContinuation {
				cluster = "[cont]"
			} else {
				cluster = "[empty]"
			}
		} else if strings.Contains(cluster, " ") && len(cluster) == 1 {
			cluster = " "
		}
		fmt.Printf("    [%d]: %q w=%d cont=%v\n", x, cluster, cell.Width, cell.IsContinuation)
	}
}
