//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func main() {
	fmt.Println("=== Arrow Width Bug Test ===\n")

	// Create a renderer with double buffering
	renderer := paint.NewRenderer(40, 10)

	// Frame 1: Draw with arrow → (width 2)
	fmt.Println("Frame 1: Drawing 'Button →' at (5, 3)")
	backBuf := renderer.GetBackBuffer()
	backBuf.SetString(5, 3, "Button →", style.Style{})

	frame1Output := renderer.Render()

	// Check buffer state after Frame 1
	fmt.Println("\nBuffer state after Frame 1:")
	printBufferLine(backBuf, 3, 20)
	printCellDetails(backBuf, 3, 5, 15)

	fmt.Printf("Frame 1 output: %q\n", frame1Output)

	// Frame 2: Draw with hyphen - (width 1)
	fmt.Println("\n--- Frame 2: Drawing 'Button-' at (5, 3) ---")
	backBuf.SetString(5, 3, "Button-", style.Style{})

	// Check buffer state before render
	fmt.Println("\nBuffer state before render:")
	printBufferLine(backBuf, 3, 20)
	printCellDetails(backBuf, 3, 5, 15)

	frame2Output := renderer.Render()

	// Check buffer state after render
	frontBuf := renderer.GetFrontBuffer()
	fmt.Println("\nBuffer state after render:")
	printBufferLine(frontBuf, 3, 20)
	printCellDetails(frontBuf, 3, 5, 15)

	fmt.Printf("Frame 2 output: %q\n", frame2Output)

	// Check for bugs
	fmt.Println("\n=== Bug Detection ===")
	bugFound := false

	// Check if arrow residue exists
	for x := 12; x < 14; x++ {
		cell := frontBuf.GetContent(x, 3)
		if strings.Contains(cell.Cluster, "→") {
			fmt.Printf("❌ BUG: Arrow residue found at [%d,3]\n", x)
			bugFound = true
		}
	}

	// Check for continuation cells
	for x := 13; x < 17; x++ {
		cell := frontBuf.GetContent(x, 3)
		if cell.IsContinuation {
			fmt.Printf("❌ BUG: Continuation cell found at [%d,3]\n", x)
			bugFound = true
		}
	}

	// Check if output cleared the arrow
	if !strings.Contains(frame2Output, "\x1b[4;14H") {
		fmt.Println("⚠️  WARNING: Output doesn't include position for clearing arrow")
		fmt.Println("   (This might be OK if the renderer handles clearing differently)")
	}

	if bugFound {
		fmt.Println("\n❌ BUG CONFIRMED: Arrow residue or continuation cell detected")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ PASS: No bugs detected")
		os.Exit(0)
	}
}

func printBufferLine(buf *paint.Buffer, y, endX int) {
	var sb strings.Builder
	sb.WriteString("  [")
	for x := 0; x < endX; x++ {
		cell := buf.GetContent(x, y)
		if cell.IsContinuation {
			sb.WriteString("◼")
		} else if cell.Cluster == "" {
			sb.WriteString("·")
		} else {
			sb.WriteString(cell.Cluster)
		}
	}
	sb.WriteString("]")
	fmt.Println(sb.String())
}

func printCellDetails(buf *paint.Buffer, y, startX, endX int) {
	fmt.Printf("  Cell details from %d to %d:\n", startX, endX)
	for x := startX; x < endX; x++ {
		cell := buf.GetContent(x, y)
		cluster := cell.Cluster
		if cluster == "" {
			cluster = "∅"
		}
		fmt.Printf("    [%d]: %q w=%d cont=%v\n", x, cluster, cell.Width, cell.IsContinuation)
	}
}
