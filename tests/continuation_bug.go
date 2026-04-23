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
	fmt.Println("Testing Continuation Cell Clearing Bug")
	fmt.Println("======================================")

	buf := paint.NewBuffer(20, 5)

	// First frame: Write text with wide arrow
	fmt.Println("\nFrame 1: Writing 'Test → Next' (arrow width 2)")
	buf.SetString(0, 0, "Test → Next", style.Style{})
	printBuffer(buf)

	// Get cell states
	fmt.Println("\nCell states after Frame 1:")
	for x := 0; x < 10; x++ {
		cell := buf.GetContent(x, 0)
		fmt.Printf("  [%d]: Cluster=%q, Width=%d, IsContinuation=%v\n",
			x, cell.Cluster, cell.Width, cell.IsContinuation)
	}

	// Second frame: Write shorter text
	fmt.Println("\nFrame 2: Writing 'Done-' (arrow width 1)")
	buf.SetString(0, 0, "Done-", style.Style{})
	printBuffer(buf)

	// Get cell states
	fmt.Println("\nCell states after Frame 2:")
	for x := 0; x < 10; x++ {
		cell := buf.GetContent(x, 0)
		fmt.Printf("  [%d]: Cluster=%q, Width=%d, IsContinuation=%v\n",
			x, cell.Cluster, cell.Width, cell.IsContinuation)
	}

	// Find any remaining arrow parts
	bugFound := false
	fmt.Println("\nChecking for leftover characters:")
	for x := 0; x < buf.Width; x++ {
		cell := buf.GetContent(x, 0)
		if cell.IsContinuation {
			fmt.Printf("  ❌ BUG: Continuation cell found at [%d,0] without head!\n", x)
			bugFound = true
		}
		if strings.Contains(cell.Cluster, "→") {
			fmt.Printf("  ❌ BUG: Arrow character found at [%d,0] after overwrite!\n", x)
			bugFound = true
		}
	}

	if bugFound {
		fmt.Println("\n❌ BUG REPRODUCED: continuation cell not cleared!")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ PASS: No leftover characters found")
		os.Exit(0)
	}
}

func printBuffer(buf *paint.Buffer) {
	fmt.Println("Buffer content:")
	for y := 0; y < buf.Height; y++ {
		var line strings.Builder
		for x := 0; x < buf.Width; x++ {
			cell := buf.GetContent(x, y)
			if cell.IsContinuation {
				line.WriteString("◼") // Mark continuation
			} else if cell.Cluster == "" || cell.Cluster == " " {
				line.WriteString(" ")
			} else {
				line.WriteString(cell.Cluster)
			}
		}
		fmt.Printf("  %s\n", line.String())
	}
}
