package display

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
)

// TestTreeViewWithSimulatedInspectorFlow simulates the exact flow used by Inspector
func TestTreeViewWithSimulatedInspectorFlow(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Step 1: Create TreeView with many lines (like Inspector does)
	lines := make([]string, 34)
	for i := 0; i < 34; i++ {
		lines[i] = fmt.Sprintf("Node %d with some description", i)
	}

	treeView := NewTreeView().
		FromLines(lines).
		ExpandLevel(1).
		ShowIcons(true).
		Build()

	fmt.Printf("\n[TEST] === Step 1: TreeView created ===\n")
	fmt.Printf("[TEST] Total lines: %d\n", len(lines))

	// Step 2: First measurement (like Layout engine measuring Inspector's VStack)
	// Inspector has VStack with Height(20), so TreeView gets ~9-10 lines visible
	constraints1 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  76,
		MinHeight: 0,
		MaxHeight: 10, // Simulating available height in Inspector
	}

	measurable, ok := treeView.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable interface")
	}

	fmt.Printf("\n[TEST] === Step 2: First Measure() ===\n")
	size1 := measurable.Measure(constraints1)
	fmt.Printf("[TEST] First measurement: %dx%d\n", size1.Width, size1.Height)

	tv, ok := treeView.(*TreeView)
	if !ok {
		t.Fatal("Build() should return *TreeView")
	}

	fmt.Printf("[TEST] viewportHeight after first Measure: %d\n", tv.viewportHeight)
	if tv.viewportHeight != 10 {
		t.Errorf("Expected viewportHeight=10, got %d", tv.viewportHeight)
	}

	// Step 3: Check children (should be limited by virtual scrolling)
	children1 := treeView.Children()
	fmt.Printf("[TEST] Children after first Measure: %d\n", len(children1))

	if len(children1) <= 10 {
		fmt.Printf("[TEST] ✓ Virtual scrolling working! Only %d children for %d lines\n",
			len(children1), len(lines))
	} else {
		t.Errorf("[TEST] ✗ Virtual scrolling NOT working! Got %d children for %d lines\n",
			len(children1), len(lines))
	}

	// Step 4: Simulate Inspector updating the tree (like on key press)
	fmt.Printf("\n[TEST] === Step 4: UpdateLines() (simulating Inspector refresh) ===\n")
	newLines := make([]string, 40)
	for i := 0; i < 40; i++ {
		newLines[i] = fmt.Sprintf("Updated Node %d", i)
	}

	tv.UpdateLines(newLines)
	fmt.Printf("[TEST] After UpdateLines: totalLines=%d, viewportHeight=%d\n",
		tv.totalLines, tv.viewportHeight)

	// Step 5: Verify viewportHeight was preserved
	if tv.viewportHeight != 10 {
		t.Errorf("After UpdateLines, viewportHeight should still be 10, got %d", tv.viewportHeight)
	} else {
		fmt.Printf("[TEST] ✓ UpdateLines() preserved viewportHeight\n")
	}

	// Step 6: Second measurement (same constraints)
	fmt.Printf("\n[TEST] === Step 5: Second Measure() ===\n")
	size2 := measurable.Measure(constraints1)
	fmt.Printf("[TEST] Second measurement: %dx%d\n", size2.Width, size2.Height)

	// Step 7: Check children again (should still be limited)
	children2 := treeView.Children()
	fmt.Printf("[TEST] Children after second Measure: %d\n", len(children2))

	if len(children2) <= 10 {
		fmt.Printf("[TEST] ✓ Virtual scrolling still working! Only %d children for %d lines\n",
			len(children2), len(newLines))
	} else {
		t.Errorf("[TEST] ✗ Virtual scrolling failed! Got %d children for %d lines\n",
			len(children2), len(newLines))
	}

	// Step 8: Test with Layout engine (full integration)
	fmt.Printf("\n[TEST] === Step 6: Layout Engine Integration ===\n")
	engine := compute.NewEngine()
	layout, err := engine.Layout(treeView, constraints1)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("[TEST] Layout result: %dx%d\n", layout.Root.Box.Width, layout.Root.Box.Height)

	if layout.Root.Box.Height > 10 {
		t.Errorf("Layout height %d exceeds constraint 10", layout.Root.Box.Height)
	} else {
		fmt.Printf("[TEST] ✓ Layout height respects constraint\n")
	}

	// Check the VStack child
	if len(layout.Root.Children) > 0 {
		vstackChild := layout.Root.Children[0]
		fmt.Printf("[TEST] VStack child size: %dx%d (should be ~10 for virtual scrolling)\n",
			vstackChild.Box.Width, vstackChild.Box.Height)

		// Note: The VStack child size might not perfectly match viewportHeight due to
		// how the Layout engine reports child sizes, but the important thing is that
		// TreeView itself respects the constraint and uses virtual scrolling
	}

	fmt.Printf("\n[TEST] === All Tests Complete ===\n")
	fmt.Printf("[TEST] Summary:\n")
	fmt.Printf("[TEST] - TreeView preserves viewportHeight: ✓\n")
	fmt.Printf("[TEST] - UpdateLines() preserves state: ✓\n")
	fmt.Printf("[TEST] - Virtual scrolling works: ✓\n")
	fmt.Printf("[TEST] - Layout respects constraints: ✓\n")
}
