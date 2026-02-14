package compute

import (
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestDemo1LikeLayoutWithNoKeys 测试类似于 demo1 的布局，没有 keys 的情况
// 根据 demo1_full_featured/main.go 的结构：
// - Boutered 包裹 VStack/HStack
// - 嵌套的 Bordered 组件
// - 没有 keys 的静态 UI
func TestDemo1LikeLayoutWithNoKeys(t *testing.T) {
	engine := NewEngine(EngineConfig{})

	// 模拟 demo1 的布局结构：
	// Header:
	//   Bordered (没有 key)
	//     HStack (没有 key)
	//       Text (没有 key)
	//       Button (没有 key)
	//       Text (没有 key)

	textVNode := func(content string) rtui.VNode {
		return rtui.Element("text").Prop("content", content).Build()
	}

	buttonVNode := func(label string) rtui.VNode {
		return rtui.Element("button").Prop("label", label).Build()
	}

	rootVNode := rtui.Bordered().
		Child(
			rtui.Element("hstack").Children(
				textVNode("TUI Engine Demo"),
				textVNode("              "),
				buttonVNode("[Open Modal]"),
				textVNode(" "),
				textVNode("Clicks: 0"),
			).Build(),
		).
		Build()

	// 创建 Fiber 树（第一次渲染）
	fiberTree := rtui.CreateFiberFromVNode(rootVNode)

	// 执行 layout
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}
	layout, err := engine.Layout(rootVNode, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 验证 NodeID 不重复
	nodeIDs := collectNodeIDs(layout.Root)
	t.Logf("Collected NodeIDs: %v", nodeIDs)

	// 检查是否有重复的 NodeID
	if hasDuplicateNodeIDs(nodeIDs) {
		t.Errorf("Found duplicate NodeIDs: %v", nodeIDs)
		// 打印详细的树结构用于调试
		printComputedBoxTree(layout.Root, 0)
	} else {
		t.Log("✅ No duplicate NodeIDs found")
	}
}

// TestDemo1LikeLayoutWithReorder 测试 demo1 风格布局的重新排序
func TestDemo1LikeLayoutWithReorder(t *testing.T) {
	engine := NewEngine(EngineConfig{})

	textVNode := func(content string) rtui.VNode {
		return rtui.Element("text").Prop("content", content).Build()
	}

	// 初始布局
	initialVNode := rtui.Bordered().
		Child(
			rtui.Element("hstack").Children(
				textVNode("Item 1"),
				textVNode("Item 2"),
				textVNode("Item 3"),
			).Build(),
		).
		Build()

	// 第一次渲染：创建 Fiber 树
	fiberTree := rtui.CreateFiberFromVNode(initialVNode)

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	layout1, err := engine.Layout(initialVNode, fiberTree, constraints)
	if err != nil {
		t.Fatalf("First layout failed: %v", err)
	}

	// 记录第一次渲染的 NodeIDs
	initialNodeIDs := collectNodeIDs(layout1.Root)
	t.Logf("Initial NodeIDs: %v", initialNodeIDs)

	// 重新排序 children (没有 key，这是静态 UI)
	reorderedVNode := rtui.Bordered().
		Child(
			rtui.Element("hstack").Children(
				textVNode("Item 3"),
				textVNode("Item 2"),
				textVNode("Item 1"),
			).Build(),
		).
		Build()

	// 第二次渲染：重新 reconcile
	fiberTree2 := rtui.CreateFiberFromVNode(reorderedVNode)

	layout2, err := engine.Layout(reorderedVNode, fiberTree2, constraints)
	if err != nil {
		t.Fatalf("Second layout failed: %v", err)
	}

	// 记录第二次渲染的 NodeIDs
	reorderedNodeIDs := collectNodeIDs(layout2.Root)
	t.Logf("Reordered NodeIDs: %v", reorderedNodeIDs)

	// 检查是否有重复的 NodeIDs
	if hasDuplicateNodeIDs(reorderedNodeIDs) {
		t.Errorf("Found duplicate NodeIDs after reorder: %v", reorderedNodeIDs)
		printComputedBoxTree(layout2.Root, 0)
	} else {
		t.Log("✅ No duplicate NodeIDs after reorder")
	}
}

// TestNestedBorderedWithoutKeys 测试嵌套的 Bordered 组件（没有 keys）
func TestNestedBorderedWithoutKeys(t *testing.T) {
	engine := NewEngine(EngineConfig{})

	textVNode := func(content string) rtui.VNode {
		return rtui.Element("text").Prop("content", content).Build()
	}

	buttonVNode := func(label string) rtui.VNode {
		return rtui.Element("button").Prop("label", label).Build()
	}

	vStackVNode := func(children ...rtui.VNode) rtui.VNode {
		return rtui.Element("vstack").Children(children...).Build()
	}

	flexProp := func(vnode rtui.VNode, factor int) rtui.VNode {
		vnode.SetProp("flex", factor)
		return vnode
	}

	// 模拟 Body 部分的结构：
	// HStack
	//   Flex(Bordered(VStack))  // sidebar
	//   Flex(Bordered(VStack))  // contentArea

	hStackNode := rtui.Element("hstack").
		Children(
			flexProp(
				rtui.Bordered().
					Child(
						vStackVNode(
							textVNode("Menu"),
							buttonVNode("Add"),
							buttonVNode("Quit"),
						),
					).
					Build(),
				1,
			),
			flexProp(
				rtui.Bordered().
					Child(
						vStackVNode(
							textVNode("Content 1"),
							textVNode("Content 2"),
						),
					).
					Build(),
				1,
			),
		).
		Prop("gap", 0).
		Build()

	// 创建 Fiber 树
	fiberTree := rtui.CreateFiberFromVNode(rootVNode)

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	layout, err := engine.Layout(rootVNode, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 验证 NodeID 不重复且父节点和子节点的 NodeID 不同
	nodeIDs := collectNodeIDs(layout.Root)
	t.Logf("NodeIDs in nested layout: %v", nodeIDs)

	if hasDuplicateNodeIDs(nodeIDs) {
		t.Errorf("Found duplicate NodeIDs in nested layout")
		printComputedBoxTree(layout.Root, 0)
	} else {
		t.Log("✅ No duplicate NodeIDs in nested layout")
	}

	// 验证父子节点的 NodeID 不同
	verifyParentChildNodeIDsDifferent(t, layout.Root)
}

// collectNodeIDs 递归收集所有 ComputedBox 的 NodeID
func collectNodeIDs(root *ComputedBox) map[uint64]int {
	result := make(map[uint64]int)
	var traverse func(box *ComputedBox)
	traverse = func(box *ComputedBox) {
		if box == nil {
			return
		}
		if box.NodeID != 0 {
			result[box.NodeID]++
		}
		for _, child := range box.Children {
			traverse(child)
		}
	}
	traverse(root)
	return result
}

// hasDuplicateNodeIDs 检查是否有重复的 NodeID
func hasDuplicateNodeIDs(nodeIDs map[uint64]int) bool {
	for _, count := range nodeIDs {
		if count > 1 {
			return true
		}
	}
	return false
}

// verifyParentChildNodeIDsDifferent 验证父节点和子节点的 NodeID 不同
func verifyParentChildNodeIDsDifferent(t *testing.T, root *ComputedBox) {
	var traverse func(box *ComputedBox)
	traverse = func(box *ComputedBox) {
		if box == nil || len(box.Children) == 0 {
			return
		}

		for _, child := range box.Children {
			if child.NodeID != 0 && box.NodeID != 0 {
				if child.NodeID == box.NodeID {
					t.Errorf("Child has same NodeID as parent: parent.NodeID=%d child.NodeID=%d (parent tag=%s child tag=%s)",
						box.NodeID, child.NodeID, getVNodeTag(box.VNode), getVNodeTag(child.VNode))
				} else {
					t.Logf("✅ Parent (NodeID=%d) and Child (NodeID=%d) have different IDs (parent tag=%s child tag=%s)",
						box.NodeID, child.NodeID, getVNodeTag(box.VNode), getVNodeTag(child.VNode))
				}
			}
			traverse(child)
		}
	}
	traverse(root)
}

// printComputedBoxTree 打印 ComputedBox 树结构用于调试
func printComputedBoxTree(box *ComputedBox, indent int) {
	if box == nil {
		return
	}

	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	tag := getVNodeTag(box.VNode)
	t.Logf("%sBox: NodeID=%d, tag=%s, children=%d", prefix, box.NodeID, tag, len(box.Children))
	for _, child := range box.Children {
		printComputedBoxTree(child, indent+1)
	}
}

// verifyBoxNodeIDs 验证 ComputedBox 树中的 NodeID（从旧测试移过来）
func verifyBoxNodeIDs(t *testing.T, box *ComputedBox, expectedNodeIDs map[string]uint64) {
	t.Helper()
	t.Log("=== Verifying ComputedBox Tree ===")
	var traverse func(box *ComputedBox)
	traverse = func(box *ComputedBox) {
		if box == nil {
			return
		}

		tag := getVNodeTag(box.VNode)
		expectedNodeID, hasExpected := expectedNodeIDs[tag]

		t.Logf("Box: NodeID=%d, tag=%s, expectedNodeID=%d (hasExpected=%v)",
			box.NodeID, tag, expectedNodeID, hasExpected)

		// 验证 NodeID 是否正确（如果有期望值）
		if hasExpected && box.NodeID != expectedNodeID {
			t.Errorf("NodeID mismatch for tag=%s: got=%d, expected=%d",
				tag, box.NodeID, expectedNodeID)
		}

		// 验证子节点
		for _, child := range box.Children {
			traverse(child)
		}
	}
	traverse(box)
}

// verifyFiberNodeIDs 验证 Fiber 树中的 NodeID（从旧测试移过来）
func verifyFiberNodeIDs(t *testing.T, fiber *reconciler.Fiber, expectedNodeIDs map[string]uint64) {
	t.Helper()
	t.Log("=== Verifying Fiber Tree ===")
	var traverse func(fiber *reconciler.Fiber)
	traverse = func(fiber *reconciler.Fiber) {
		if fiber == nil {
			return
		}

		expectedNodeID, hasExpected := expectedNodeIDs[fiber.Tag]

		t.Logf("Fiber: NodeID=%d, tag=%s, expectedNodeID=%d (hasExpected=%v)",
			fiber.NodeID, fiber.Tag, expectedNodeID, hasExpected)

		// 验证 NodeID 是否正确（如果有期望值）
		if hasExpected && fiber.NodeID != expectedNodeID {
			t.Errorf("NodeID mismatch for tag=%s: got=%d, expected=%d",
				fiber.Tag, fiber.NodeID, expectedNodeID)
		}

		// 验证子节点
		traverse(fiber.Child)
		// 验证兄弟节点
		traverse(fiber.Sibling)
	}
	traverse(fiber)
}
