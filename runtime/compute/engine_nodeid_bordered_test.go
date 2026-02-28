package compute

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
)

// TestNodeIDExtractionWithBordered 测试 Bordered 组件的 NodeID 提取
// NOTE: This test is skipped because it tests the legacy compute.Engine behavior
func TestNodeIDExtractionWithBordered(t *testing.T) {
	t.Skip("VNode is deprecated - this test tests legacy behavior that has been migrated to runtime/layout")
	// 模拟示例中的 Bounded 包裹 HStack 的结构
	// Bordered (父)
	//   └── HStack (子，无key)
	//         └── text节点

	innerText := rtui.Element("text").Prop("content", "Inner").Build()
	innerHStack := rtui.Element("hstack").Children(innerText).Build()

	bordered := rtui.Element("bordered").Children(innerHStack).Build()

	// 创建 Fiber 树
	fiberTree := reconciler.CreateFiberFromVNode(bordered)
	if fiberTree == nil {
		t.Fatal("Failed to create fiber tree")
	}

	engine := NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 10,
	}

	layout, err := engine.Layout(bordered, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 收集所有 NodeID 的映射
	nodeIDs := make(map[string]uint64)
	nodePaths := make(map[uint64]string)

	var collectNodeIDs func(box *ComputedBox, path string)
	collectNodeIDs = func(box *ComputedBox, path string) {
		if box == nil || box.VNode == nil {
			return
		}

		boxPath := path + "/" + box.VNode.Type().String()
		if box.VNode.Key() != "" {
			boxPath += "[" + box.VNode.Key() + "]"
		}

		t.Logf("Node: path=%q, NodeID=%d", boxPath, box.NodeID)

		// 检查 NodeID 是否重复
		if existingPath, exists := nodeIDs[boxPath]; exists {
			t.Errorf("Duplicate path %q (previously seen at %q)", boxPath, existingPath)
		}
		nodeIDs[boxPath] = box.NodeID

		// 检查 NodeID 是否被多个节点共享
		if existingPath, exists := nodePaths[box.NodeID]; exists {
			t.Errorf("NodeID=%d is shared by multiple nodes: %q and %q",
				box.NodeID, existingPath, boxPath)
		}
		nodePaths[box.NodeID] = boxPath

		for _, child := range box.Children {
			collectNodeIDs(child, boxPath)
		}
	}
	collectNodeIDs(layout.Root, "ROOT")

	// 验证 NodeID 唯一性
	if len(nodeIDs) != len(nodePaths) {
		t.Errorf("NodeID count mismatch: %d unique paths but only %d unique NodeIDs",
			len(nodeIDs), len(nodePaths))
	}

	t.Logf("Total nodes: %d, Unique NodeIDs: %d", len(nodeIDs), len(nodePaths))

	// 预期至少 3 个节点：bordered, hstack, text
	if len(nodeIDs) < 3 {
		t.Errorf("Expected at least 3 nodes, got %d", len(nodeIDs))
	}

	// 验证特定节点：父节点和子节点的 NodeID 不应该相同
	// 注意：rtui.Element("bordered") 创建的是 Element 类型，所以路径是 "ROOT/Element"
	borderedNodeID := nodeIDs["ROOT/Element"]
	hStackNodeID := nodeIDs["ROOT/Element/Element"]
	textNodeID := nodeIDs["ROOT/Element/Element/Element"]

	if borderedNodeID == hStackNodeID && borderedNodeID != 0 {
		t.Errorf("Parent (bordered) NodeID=%d equals child (hstack) NodeID=%d",
			borderedNodeID, hStackNodeID)
	}

	if hStackNodeID == textNodeID && hStackNodeID != 0 {
		t.Errorf("Parent (hstack) NodeID=%d equals child (text) NodeID=%d",
			hStackNodeID, textNodeID)
	}

	t.Logf("NodeIDs: bordered=%d, hstack=%d, text=%d", borderedNodeID, hStackNodeID, textNodeID)
}

// TestNodeIDExtractionNestedBordered 测试嵌套 Bordered 组件
// NOTE: This test is skipped because it tests the legacy compute.Engine behavior
func TestNodeIDExtractionNestedBordered(t *testing.T) {
	t.Skip("VNode is deprecated - this test tests legacy behavior that has been migrated to runtime/layout")
	// 模拟示例中的嵌套 Bounded 结构
	// Outer Element
	//   └── VStack
	//         ├── Inner Element 1
	//         │       └── HStack
	//         └── Inner Element 2
	//                 └── HStack

	innerHStack1 := rtui.Element("hstack").Children(
		rtui.Element("text").Prop("content", "Panel 1").Build(),
	).Build()

	bordered1 := rtui.Element("bordered").Children(innerHStack1).Build()

	innerHStack2 := rtui.Element("hstack").Children(
		rtui.Element("text").Prop("content", "Panel 2").Build(),
	).Build()

	bordered2 := rtui.Element("bordered").Children(innerHStack2).Build()

	vStack := rtui.Element("vstack").Children(bordered1, bordered2).Build()

	outerBordered := rtui.Element("bordered").Children(vStack).Build()

	// 创建 Fiber 树
	fiberTree := reconciler.CreateFiberFromVNode(outerBordered)
	if fiberTree == nil {
		t.Fatal("Failed to create fiber tree")
	}

	engine := NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 30,
	}

	layout, err := engine.Layout(outerBordered, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 收集所有 NodeID
	nodeIDs := make(map[string]uint64)
	nodePaths := make(map[uint64]string)

	var collectNodeIDs func(box *ComputedBox, path string)
	collectNodeIDs = func(box *ComputedBox, path string) {
		if box == nil || box.VNode == nil {
			return
		}

		boxPath := path + "/" + box.VNode.Type().String()
		if box.VNode.Key() != "" {
			boxPath += "[" + box.VNode.Key() + "]"
		}

		t.Logf("Node: path=%q, NodeID=%d", boxPath, box.NodeID)

		// 检查 NodeID 唯一性
		if existingPath, exists := nodeIDs[boxPath]; exists {
			// 相同路径是允许的（多个相同类型的节点），不报错
			t.Logf("Same path %q seen again (previously at %q)", boxPath, existingPath)
		}
		nodeIDs[boxPath] = box.NodeID

		// 检查 NodeID 是否被多个不同路径的节点共享（这才是真正的错误）
		if existingPath, exists := nodePaths[box.NodeID]; exists && existingPath != boxPath {
			t.Errorf("NodeID=%d is shared by different nodes: %q and %q",
				box.NodeID, existingPath, boxPath)
		}
		nodePaths[box.NodeID] = boxPath

		for _, child := range box.Children {
			collectNodeIDs(child, boxPath)
		}
	}
	collectNodeIDs(layout.Root, "ROOT")

	t.Logf("Total nodes: %d, Unique NodeIDs: %d", len(nodeIDs), len(nodePaths))

	// 验证所有节点都有唯一的 NodeID
	// 核心检查：每个 NodeID 只对应一个节点
	uniqueNodeCount := len(nodePaths)
	if uniqueNodeCount < 5 {
		t.Errorf("Expected at least 5 unique nodes, got %d", uniqueNodeCount)
	}

	// 验证父子节点 NodeID 不同
	var verifyParentChild func(box *ComputedBox) int
	verifyParentChild = func(box *ComputedBox) int {
		if box == nil || box.VNode == nil {
			return 0
		}

		errorCount := 0
		for _, child := range box.Children {
			if child != nil && child.VNode != nil {
				if box.NodeID == child.NodeID {
					t.Errorf("Parent NodeID=%d equals child NodeID=%d",
						box.NodeID, child.NodeID)
					errorCount++
				}
				errorCount += verifyParentChild(child)
			}
		}
		return errorCount
	}

	conflicts := verifyParentChild(layout.Root)
	if conflicts > 0 {
		t.Errorf("Found %d parent-child NodeID conflicts", conflicts)
	} else {
		t.Log("Test passed: Nested bordered nodes have unique NodeIDs")
	}
}

// TestDemo1LikeLayout 测试类似于 demo1 的完整布局结构
func TestDemo1LikeLayout(t *testing.T) {
	// 模拟 demo1 的顶层简化结构
	// VStack
	//   ├── Element (Header)
	//   │       └── HStack
	//   │               ├── Text
	//   │               ├── Text
	//   │               ├── Button
	//   │               ├── Text
	//   │               └── Text
	//   └── HStack
	//           ├── Element (Sidebar)
	//           │       └── VStack
	//           └── Element (Content)
	//                   └── VStack

	// Header
	headerHStack := rtui.Element("hstack").Children(
		rtui.Element("text").Prop("content", "Title").Build(),
		rtui.Element("text").Prop("content", "Spacer").Build(),
		rtui.Element("text").Prop("content", "[Button]").Build(),
		rtui.Element("text").Prop("content", "Spacer2").Build(),
		rtui.Element("text").Prop("content", "Count: 0").Build(),
	).Build()
	headerBordered := rtui.Element("bordered").Children(headerHStack).Build()

	// Sidebar
	sidebarVStack := rtui.Element("vstack").Children(
		rtui.Element("text").Prop("content", "Item1").Build(),
		rtui.Element("text").Prop("content", "Item2").Build(),
	).Build()
	sidebarBordered := rtui.Element("bordered").Children(sidebarVStack).Build()

	// Content
	contentVStack := rtui.Element("vstack").Children(
		rtui.Element("text").Prop("content", "Content1").Build(),
		rtui.Element("text").Prop("content", "Content2").Build(),
	).Build()
	contentBordered := rtui.Element("bordered").Children(contentVStack).Build()

	// Main body
	bodyHStack := rtui.Element("hstack").Children(sidebarBordered, contentBordered).Build()

	// Root
	root := rtui.Element("vstack").Children(headerBordered, bodyHStack).Build()

	// 创建 Fiber 树
	fiberTree := reconciler.CreateFiberFromVNode(root)
	if fiberTree == nil {
		t.Fatal("Failed to create fiber tree")
	}

	engine := NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	layout, err := engine.Layout(root, fiberTree, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// 收集所有 NodeID 并验证唯一性
	nodePaths := make(map[uint64]string)
	totalNodes := 0

	var traverse func(box *ComputedBox, path string)
	traverse = func(box *ComputedBox, path string) {
		if box == nil || box.VNode == nil {
			return
		}

		boxPath := path + "/" + box.VNode.Type().String()
		totalNodes++

		t.Logf("Node[%d]: path=%q, NodeID=%d", totalNodes, boxPath, box.NodeID)

		// 检查 NodeID 是否被多个不同节点共享
		if existingPath, exists := nodePaths[box.NodeID]; exists {
			t.Errorf("❌ DUPLICATE NodeID=%d: %q and %q",
				box.NodeID, existingPath, boxPath)
		}
		nodePaths[box.NodeID] = boxPath

		for _, child := range box.Children {
			traverse(child, boxPath)
		}
	}
	traverse(layout.Root, "ROOT")

	t.Logf("=== Demo1-like Layout Summary ===")
	t.Logf("Total nodes: %d", totalNodes)
	t.Logf("Unique NodeIDs: %d", len(nodePaths))

	// 核心验证：每个节点都有唯一的 NodeID
	if totalNodes != len(nodePaths) {
		t.Errorf("❌ NODEID DUPLICATION: %d nodes but only %d unique NodeIDs",
			totalNodes, len(nodePaths))
	} else {
		t.Logf("✅ All %d NodeIDs are unique!", totalNodes)
	}

	// 验证所有父子节点 NodeID 不同
	var verifyParentChild func(box *ComputedBox) int
	verifyParentChild = func(box *ComputedBox) int {
		if box == nil || box.VNode == nil {
			return 0
		}

		errorCount := 0
		for _, child := range box.Children {
			if child != nil && child.VNode != nil {
				if box.NodeID == child.NodeID {
					t.Errorf("❌ PARENT-CHILD NODEID CONFLICT: parent %s (NodeID=%d) = child %s (NodeID=%d)",
						box.VNode.Type().String(), box.NodeID,
						child.VNode.Type().String(), child.NodeID)
					errorCount++
				}
				errorCount += verifyParentChild(child)
			}
		}
		return errorCount
	}

	conflicts := verifyParentChild(layout.Root)
	if conflicts > 0 {
		t.Errorf("Found %d parent-child NodeID conflicts", conflicts)
	} else {
		t.Logf("✅ No parent-child NodeID conflicts!")
	}

	// 最终判断
	if totalNodes == len(nodePaths) && conflicts == 0 {
		t.Log("🎉 Test passed: All NodeIDs are correct and unique!")
	}
}
