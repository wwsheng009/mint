package ui

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
)

func main() {
	// Enable debug
	os.Setenv("xui_DEBUG_LAYOUT", "true")

	fmt.Println("=== Testing Text Node Width ===")

	// Use the Text nodes that work in tests
	text1 := Text("Hello")
	text2 := Text("World")

	root := VStack(text1, text2)

	fmt.Printf("Text1 content: %q\n", getTextNodeContent(text1))
	fmt.Printf("Text2 content: %q\n", getTextNodeContent(text2))

	// Create Fiber tree
	fiberRoot := CreateFiberFromVNode(root)
	if fiberRoot == nil {
		fmt.Println("❌ Fiber root is nil")
		os.Exit(1)
	}

	fmt.Printf("\n2. Fiber Root.NodeID=%d\n", fiberRoot.NodeID)

	// Run Layout
	fmt.Println("3. Running Layout...")
	engine := compute.NewEngine()
	engine.SetDebug(true)

	layout, err := engine.Layout(root, fiberRoot, runtime.UnboundedConstraints())
	if err != nil {
		fmt.Printf("❌ Layout failed: %v\n", err)
		os.Exit(1)
	}

	if layout == nil || layout.Root == nil {
		fmt.Println("❌ Layout root is nil")
		os.Exit(1)
	}

	fmt.Printf("\n=== Box Sizes ===\n")
	fmt.Printf("Root: NodeID=%d, Box=(%d, %d, %dx%d)\n",
		layout.Root.NodeID, layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)

	for i, child := range layout.Root.Children {
		fmt.Printf("Child %d: NodeID=%d, Box=(%d, %d, %dx%d)\n",
			i, child.NodeID, child.Box.X, child.Box.Y, child.Box.Width, child.Box.Height)
	}

	fmt.Printf("\nHitMap entries: %d\n", layout.HitMap.Size())

	if layout.HitMap != nil && layout.HitMap.Size() > 0 {
		fmt.Println("\n✅ HitMap has entries!")
	} else {
		fmt.Println("\n⚠️ HitMap is empty!")
	}
}

func getTextNodeContent(node VNode) string {
	// Try to get text content
	if node == nil {
		return ""
	}
	props := node.Props()
	if props != nil {
		if text, ok := props["text"].(string); ok {
			return text
		}
	}
	return ""
}
