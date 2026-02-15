package compute

import (
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestDebugFiberTreeStructure 调试 Fiber 树结构
func TestDebugFiberTreeStructure(t *testing.T) {
	// 创建简单的 Bounded 结构
	// Bordered (无key)
	//   └── HStack (无key)
	//         └── Text (无key)

	textNode := rtui.Element("text").Prop("content", "Hello").Build()
	hStackNode := rtui.Element("hstack").Children(textNode).Build()
	borderNode := rtui.Element("bordered").Children(hStackNode).Build()

	// 创建 Fiber 树
	fiberTree := reconciler.CreateFiberFromVNode(borderNode)
	if fiberTree == nil {
		t.Fatal("Failed to create fiber tree")
	}

	// 打印 Fiber 树结构
	t.Log("=== Fiber Tree Structure ===")
	printFiberTree(fiberTree, 0, t)

	// 创建 Engine 并 Layout
	engine := NewEngine()
	engine.SetDebug(true)

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  20,
		MinHeight: 0,
		MaxHeight: 10,
	}

	layout, err := engine.Layout(borderNode, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 打印 ComputedBox 树结构
	t.Log("\n=== ComputedBox Tree Structure ===")
	printComputedBoxTree(layout.Root, 0, t)

	// 验证 NodeID
	t.Log("\n=== NodeID Verification ===")
	var verifyNodeIDs func(box *ComputedBox, parentBox *ComputedBox)
	verifyNodeIDs = func(box *ComputedBox, parentBox *ComputedBox) {
		if box == nil || box.GetVNode() == nil {
			return
		}

		tag := getVNodeTag(box.GetVNode())
		key := box.GetVNode().Key()

		t.Logf("Box: tag=%q, key=%q, NodeID=%d", tag, key, box.NodeID)

		// 检查 NodeID 是否为 0
		if box.NodeID == 0 {
			t.Errorf("ERROR: NodeID is 0 for tag=%q", tag)
		}

		// 检查与父节点的 NodeID 是否重复
		if parentBox != nil {
			if box.NodeID == parentBox.NodeID {
				t.Errorf("ERROR: Child (tag=%q) has same NodeID as parent (tag=%q): %d",
					tag, getVNodeTag(parentBox.GetVNode()), box.NodeID)
			}
		}

		for _, child := range box.Children {
			verifyNodeIDs(child, box)
		}
	}
	verifyNodeIDs(layout.Root, nil)
}

func printFiberTree(fiber *rtui.Fiber, depth int, t *testing.T) {
	if fiber == nil {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	tag := ""
	if tagger, ok := fiber.VNode.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	key := fiber.VNode.Key()

	t.Logf("%sFiber: nodeID=%d, tag=%q, key=%q, diffKey=%q, siblingIndex=%d, pathSegment=%q",
		indent, fiber.NodeID, tag, key, fiber.DiffKey, fiber.SiblingIndex, fiber.PathSegment)

	printFiberTree(fiber.Child, depth+1, t)
	printFiberTree(fiber.Sibling, depth, t)
}

func printComputedBoxTree(box *ComputedBox, depth int, t *testing.T) {
	if box == nil || box.GetVNode() == nil {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	tag := getVNodeTag(box.GetVNode())
	key := box.GetVNode().Key()
	childFiberNodeID := uint64(0)
	if box.ChildFiber != nil {
		childFiberNodeID = box.ChildFiber.NodeID
	}

	t.Logf("%sBox: nodeID=%d, tag=%q, key=%q, childFiberNodeID=%d, children=%d",
		indent, box.NodeID, tag, key, childFiberNodeID, len(box.Children))

	for _, child := range box.Children {
		printComputedBoxTree(child, depth+1, t)
	}
}

func getVNodeTag(vnode rtui.VNode) string {
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return "unknown"
}
