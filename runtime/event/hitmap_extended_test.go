package event

import (
	"fmt"
	"sync"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
)

// ============================================================================
// Phase 1-7: 扩展测试 - 性能、边界、并发、复杂场景
// ============================================================================

// TestHitMap_Performance_LargeBuild 大规模节点构建性能测试
func TestHitMap_Performance_LargeBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// 创建包含 1000 个节点的布局树
	const nodeCount = 1000

	// 构建一个扁平的节点列表（避免中间容器节点）
	nodes := make([]layout.Node, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodes[i] = &mockNode{
			id:       fmt.Sprintf("node-%d", i),
			nodeType: "leaf",
			x:        (i % 50) * 20,
			y:        (i / 50) * 20,
			width:    10,
			height:   10,
		}
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    1000,
		height:   1000,
		children: nodes,
	}

	// 测试构建性能（运行多次取平均）
	for i := 0; i < 10; i++ {
		hitMap := BuildHitMap(root)
		expectedSize := nodeCount + 1 // root + children
		if hitMap.Size() != expectedSize {
			t.Errorf("Expected %d nodes, got %d", expectedSize, hitMap.Size())
		}
	}
}

// TestHitMap_Performance_HitTest 命中测试性能测试
func TestHitMap_Performance_HitTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	const nodeCount = 1000
	root := buildBalancedTree(nodeCount, 0, 0, 1000, 1000)
	hitMap := BuildHitMap(root)

	// 测试命中测试性能（大量命中测试操作）
	for i := 0; i < 10000; i++ {
		// 测试不同位置的命中
		x := (i * 7) % 1000
		y := (i * 13) % 1000
		hitMap.HitTest(x, y)
	}
}

// TestHitMap_Performance_FindByID ID查找性能测试
func TestHitMap_Performance_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	const nodeCount = 1000
	root := buildBalancedTree(nodeCount, 0, 0, 1000, 1000)
	hitMap := BuildHitMap(root)

	// 测试 ID 查找性能（大量查找操作）
	for i := 0; i < 10000; i++ {
		nodeID := fmt.Sprintf("node-%d", i%nodeCount)
		hitMap.FindByID(stringToNodeID(nodeID))
	}
}

// buildBalancedTree 构建一个平衡的布局树
func buildBalancedTree(count, x, y, width, height int) layout.Node {
	if count == 0 {
		return nil
	}

	if count == 1 {
		return &mockNode{
			id:       fmt.Sprintf("node-%d", x+y*width),
			nodeType: "leaf",
			x:        x,
			y:        y,
			width:    width,
			height:   height,
		}
	}

	// 分裂为左右两个子树
	leftCount := count / 2
	rightCount := count - leftCount

	leftWidth := width / 2
	rightWidth := width - leftWidth

	leftTree := buildBalancedTree(leftCount, x, y, leftWidth, height)
	rightTree := buildBalancedTree(rightCount, x+leftWidth, y, rightWidth, height)

	// 创建父节点（抽象容器，不参与命中测试）
	children := []layout.Node{}
	if leftTree != nil {
		children = append(children, leftTree)
	}
	if rightTree != nil {
		children = append(children, rightTree)
	}

	return &mockNode{
		id:       fmt.Sprintf("container-%d", len(children)),
		nodeType: "container",
		x:        x,
		y:        y,
		width:    width,
		height:   height,
		children: children,
	}
}

// TestHitMap_BoundaryConditions 边界条件测试
func TestHitMap_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		x, y, w, h  int
		testX, testY int
		shouldHit  bool
		description string
	}{
		{
			name: "ZeroSize",
			x: 10, y: 10, w: 0, h: 0,
			testX: 10, testY: 10,
			shouldHit: false,
			description: "零尺寸节点不应该命中",
		},
		{
			name: "NegativePosition",
			x: -10, y: -10, w: 20, h: 20,
			testX: -5, testY: -5,
			shouldHit: true,
			description: "负坐标节点应该可以命中",
		},
		{
			name: "LargeCoordinates",
			x: 100000, y: 100000, w: 100, h: 100,
			testX: 100050, testY: 100050,
			shouldHit: true,
			description: "大坐标节点应该可以命中",
		},
		{
			name: "SinglePixel",
			x: 50, y: 50, w: 1, h: 1,
			testX: 50, testY: 50,
			shouldHit: true,
			description: "单像素节点应该可以命中",
		},
		{
			name: "ExtremeWidth",
			x: 0, y: 0, w: 99999, h: 10,
			testX: 50000, testY: 5,
			shouldHit: true,
			description: "极大宽度节点应该可以命中",
		},
		{
			name: "ExactCorner_TopLeft",
			x: 100, y: 100, w: 50, h: 50,
			testX: 100, testY: 100,
			shouldHit: true,
			description: "精确点击左上角应该命中",
		},
		{
			name: "ExactCorner_BottomRight",
			x: 100, y: 100, w: 50, h: 50,
			testX: 149, testY: 149,
			shouldHit: true,
			description: "精确点击右下角（不包括边界）应该命中",
		},
		{
			name: "OutsideByOne",
			x: 100, y: 100, w: 50, h: 50,
			testX: 150, testY: 150,
			shouldHit: false,
			description: "超出一个像素不应该命中",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &mockNode{
				id:       tt.name,
				nodeType: "test",
				x:        tt.x,
				y:        tt.y,
				width:    tt.w,
				height:   tt.h,
			}

			hitMap := BuildHitMap(node)
			entry := hitMap.HitTest(tt.testX, tt.testY)

			hit := entry != nil

			if hit != tt.shouldHit {
				t.Errorf("%s: expected hit=%v, got hit=%v (node at (%d,%d) size %dx%d, test point (%d,%d))",
					tt.description, tt.shouldHit, hit,
					tt.x, tt.y, tt.w, tt.h, tt.testX, tt.testY)
			}
		})
	}
}

// TestHitMap_ConcurrentAccess 并发访问测试
func TestHitMap_ConcurrentAccess(t *testing.T) {
	// 创建一个包含多个节点的 HitMap
	nodes := make([]layout.Node, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = &mockNode{
			id:       fmt.Sprintf("node-%d", i),
			nodeType: "test",
			x:        (i % 10) * 100,
			y:        (i / 10) * 100,
			width:    50,
			height:   50,
		}
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    1000,
		height:   1000,
		children: nodes,
	}

	hitMap := BuildHitMap(root)

	// 并发测试
	const numGoroutines = 10
	const operationsPerGoroutine = 1000

	var wg sync.WaitGroup
	errors := make(chan string, numGoroutines*operationsPerGoroutine)

	// 启动多个 goroutine 并发访问 HitMap
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 执行各种操作
				switch j % 4 {
				case 0:
					// HitTest
					x := (goroutineID*operationsPerGoroutine + j) % 1000
					y := (goroutineID*operationsPerGoroutine + j*2) % 1000
					hitMap.HitTest(x, y)

				case 1:
					// FindByID
					nodeID := fmt.Sprintf("node-%d", j%100)
					hitMap.FindByID(stringToNodeID(nodeID))

				case 2:
					// FindAllAt
					x := (goroutineID*operationsPerGoroutine + j) % 1000
					y := (goroutineID*operationsPerGoroutine + j*3) % 1000
					hitMap.FindAllAt(x, y)

				case 3:
					// Size
					hitMap.Size()
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Error(err)
	}

	// 验证 HitMap 仍然正常工作
	if hitMap.Size() != 101 { // root + 100 children
		t.Errorf("HitMap size changed during concurrent access, got %d, want 101", hitMap.Size())
	}
}

// TestHitMap_DeepNesting 深层嵌套测试
func TestHitMap_DeepNesting(t *testing.T) {
	// 创建深度为 100 的嵌套结构
	const depth = 100

	var current layout.Node
	current = &mockNode{
		id:       fmt.Sprintf("level-%d", depth),
		nodeType: "leaf",
		x:        0,
		y:        0,
		width:    10,
		height:   10,
	}

	// 从叶子向上构建
	for i := depth - 1; i >= 0; i-- {
		parent := &mockNode{
			id:       fmt.Sprintf("level-%d", i),
			nodeType: "container",
			x:        0,
			y:        0,
			width:    10,
			height:   10,
			children: []layout.Node{current},
		}
		current = parent
	}

	hitMap := BuildHitMap(current)

	// 验证所有节点都被添加到 HitMap
	expectedSize := depth + 1 // level-0 到 level-100
	if hitMap.Size() != expectedSize {
		t.Errorf("Expected %d nodes in deeply nested structure, got %d", expectedSize, hitMap.Size())
	}

	// 验证每一层都能找到
	for i := 0; i <= depth; i++ {
		nodeID := fmt.Sprintf("level-%d", i)
		entry := hitMap.FindByID(stringToNodeID(nodeID))
		if entry == nil {
			t.Errorf("Level %d not found in HitMap", i)
		}
	}

	// 验证命中测试
	entry := hitMap.HitTest(5, 5)
	if entry == nil {
		t.Error("Expected to hit something in deeply nested structure")
	} else {
		// 应该命中最内层的节点（最高 Z-order）
		if entry.NodeID != stringToNodeID(fmt.Sprintf("level-%d", depth)) {
			t.Errorf("Expected to hit innermost node, got %d", entry.NodeID)
		}
	}
}

// TestHitMap_OverlappingNodes 重叠节点测试
func TestHitMap_OverlappingNodes(t *testing.T) {
	// 创建部分重叠的节点
	node1 := &mockNode{
		id:       "node1",
		nodeType: "type1",
		x:        0,
		y:        0,
		width:    100,
		height:   100,
	}

	node2 := &mockNode{
		id:       "node2",
		nodeType: "type2",
		x:        50,
		y:        0,
		width:    100,
		height:   100,
	}

	node3 := &mockNode{
		id:       "node3",
		nodeType: "type3",
		x:        0,
		y:        50,
		width:    100,
		height:   100,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    200,
		height:   200,
		children: []layout.Node{node1, node2, node3},
	}

	hitMap := BuildHitMap(root)

	// 测试重叠区域的命中
	t.Run("Overlap_1and2", func(t *testing.T) {
		// node1 和 node2 重叠区域
		entry := hitMap.HitTest(75, 25)
		if entry == nil {
			t.Fatal("Expected to hit something")
		}
		// 应该命中 node2（更高 Z-order）
		if entry.NodeID != stringToNodeID("node2") {
			t.Errorf("Expected node2 (higher Z-order), got %d", entry.NodeID)
		}
	})

	t.Run("Overlap_1and3", func(t *testing.T) {
		// node1 和 node3 重叠区域
		entry := hitMap.HitTest(25, 75)
		if entry == nil {
			t.Fatal("Expected to hit something")
		}
		// 应该命中 node3（更高 Z-order）
		if entry.NodeID != stringToNodeID("node3") {
			t.Errorf("Expected node3 (higher Z-order), got %d", entry.NodeID)
		}
	})

	t.Run("AllThreeOverlap", func(t *testing.T) {
		// 三个节点都重叠的中心区域
		entry := hitMap.HitTest(75, 75)
		if entry == nil {
			t.Fatal("Expected to hit something")
		}
		// 应该命中 node3（最高 Z-order）
		if entry.NodeID != stringToNodeID("node3") {
			t.Errorf("Expected node3 (highest Z-order), got %d", entry.NodeID)
		}
	})

	t.Run("OnlyNode1", func(t *testing.T) {
		// 只有 node1 的区域
		entry := hitMap.HitTest(25, 25)
		if entry == nil {
			t.Fatal("Expected to hit something")
		}
		if entry.NodeID != stringToNodeID("node1") {
			t.Errorf("Expected node1, got %d", entry.NodeID)
		}
	})
}

// TestHitMap_WideShallowTree 宽而浅的树测试
func TestHitMap_WideShallowTree(t *testing.T) {
	// 创建一个宽而浅的树（1 个 root，1000 个直接子节点）
	const childCount = 1000

	children := make([]layout.Node, childCount)
	for i := 0; i < childCount; i++ {
		children[i] = &mockNode{
			id:       fmt.Sprintf("child-%d", i),
			nodeType: "leaf",
			x:        (i % 50) * 20,
			y:        (i / 50) * 20,
			width:    10,
			height:   10,
		}
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    1000,
		height:   1000,
		children: children,
	}

	hitMap := BuildHitMap(root)

	// 验证节点数量
	expectedSize := childCount + 1 // root + children
	if hitMap.Size() != expectedSize {
		t.Errorf("Expected %d nodes, got %d", expectedSize, hitMap.Size())
	}

	// 验证随机子节点可以找到
	testCases := []struct{ x, y int }{
		{25, 25},   // child-7 (大概位置)
		{505, 305}, // child-? (大概位置)
		{905, 905}, // 最后一个子节点
	}

	for _, tc := range testCases {
		entry := hitMap.HitTest(tc.x, tc.y)
		if entry == nil {
			t.Errorf("Expected to hit something at (%d,%d)", tc.x, tc.y)
		}
	}
}

// TestHitMap_EmptyChildren 空子节点列表测试
func TestHitMap_EmptyChildren(t *testing.T) {
	node := &mockNode{
		id:       "leaf",
		nodeType: "leaf",
		x:        10,
		y:        10,
		width:    50,
		height:   50,
		children: []layout.Node{}, // 显式空的子节点列表
	}

	hitMap := BuildHitMap(node)

	if hitMap.Size() != 1 {
		t.Errorf("Expected 1 node, got %d", hitMap.Size())
	}

	entry := hitMap.HitTest(25, 25)
	if entry == nil || entry.NodeID != stringToNodeID("leaf") {
		t.Error("Should hit leaf node")
	}
}

// TestHitMap_MultipleRoots 多个根节点测试
func TestHitMap_MultipleRoots(t *testing.T) {
	// 测试从不同的根节点构建多个独立的 HitMap
	root1 := &mockNode{
		id:       "root1",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    100,
		height:   100,
	}

	root2 := &mockNode{
		id:       "root2",
		nodeType: "root",
		x:        100,
		y:        0,
		width:    100,
		height:   100,
	}

	hitMap1 := BuildHitMap(root1)
	hitMap2 := BuildHitMap(root2)

	// 验证两个 HitMap 是独立的
	if hitMap1.Size() != 1 {
		t.Errorf("hitMap1 should have 1 node, got %d", hitMap1.Size())
	}

	if hitMap2.Size() != 1 {
		t.Errorf("hitMap2 should have 1 node, got %d", hitMap2.Size())
	}

	// 验证各自的范围
	if entry := hitMap1.HitTest(50, 50); entry == nil || entry.NodeID != stringToNodeID("root1") {
		t.Error("hitMap1 should find root1")
	}

	if entry := hitMap2.HitTest(150, 50); entry == nil || entry.NodeID != stringToNodeID("root2") {
		t.Error("hitMap2 should find root2")
	}

	// 验证不会互相干扰
	if entry := hitMap1.HitTest(150, 50); entry != nil {
		t.Error("hitMap1 should not find root2's area")
	}

	if entry := hitMap2.HitTest(50, 50); entry != nil {
		t.Error("hitMap2 should not find root1's area")
	}
}

// TestHitMap_LocalXY_EdgeCases 局部坐标边界情况
func TestHitMap_LocalXY_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		nodeX, nodeY   int
		nodeW, nodeH   int
		screenX, screenY int
		expectedLocalX, expectedLocalY int
	}{
		{
			name: "Inside",
			nodeX: 10, nodeY: 20, nodeW: 50, nodeH: 50,
			screenX: 30, screenY: 40,
			expectedLocalX: 20, expectedLocalY: 20,
		},
		{
			name: "TopLeftBoundary",
			nodeX: 10, nodeY: 20, nodeW: 50, nodeH: 50,
			screenX: 10, screenY: 20,
			expectedLocalX: 0, expectedLocalY: 0,
		},
		{
			name: "BottomRightBoundary",
			nodeX: 10, nodeY: 20, nodeW: 50, nodeH: 50,
			screenX: 59, screenY: 69,
			expectedLocalX: 49, expectedLocalY: 49,
		},
		{
			name: "FarOutsideTop",
			nodeX: 100, nodeY: 100, nodeW: 50, nodeH: 50,
			screenX: 100, screenY: 50,
			expectedLocalX: 0, expectedLocalY: -50,
		},
		{
			name: "FarOutsideLeft",
			nodeX: 100, nodeY: 100, nodeW: 50, nodeH: 50,
			screenX: 50, screenY: 100,
			expectedLocalX: -50, expectedLocalY: 0,
		},
		{
			name: "FarOutsideBottom",
			nodeX: 100, nodeY: 100, nodeW: 50, nodeH: 50,
			screenX: 100, screenY: 200,
			expectedLocalX: 0, expectedLocalY: 100,
		},
		{
			name: "FarOutsideRight",
			nodeX: 100, nodeY: 100, nodeW: 50, nodeH: 50,
			screenX: 200, screenY: 100,
			expectedLocalX: 100, expectedLocalY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &mockNode{
				id:       "test",
				nodeType: "test",
				x:        tt.nodeX,
				y:        tt.nodeY,
				width:    tt.nodeW,
				height:   tt.nodeH,
			}

			hitMap := BuildHitMap(node)
			entry := hitMap.FindByID(stringToNodeID("test"))

			if entry == nil {
				t.Fatal("Node not found in HitMap")
			}

			localX, localY := entry.LocalXY(tt.screenX, tt.screenY)

			if localX != tt.expectedLocalX || localY != tt.expectedLocalY {
				t.Errorf("LocalXY(%d,%d) = (%d,%d), want (%d,%d)",
					tt.screenX, tt.screenY,
					localX, localY,
					tt.expectedLocalX, tt.expectedLocalY)
			}
		})
	}
}

// TestHitMap_FindAllAt_SeveralOverlaps 多重重叠节点测试
func TestHitMap_FindAllAt_SeveralOverlaps(t *testing.T) {
	// 创建多个完全重叠的节点
	const overlapCount = 10

	children := make([]layout.Node, overlapCount)
	for i := 0; i < overlapCount; i++ {
		children[i] = &mockNode{
			id:       fmt.Sprintf("node-%d", i),
			nodeType: "overlap",
			x:        100,
			y:        100,
			width:    50,
			height:   50,
		}
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    200,
		height:   200,
		children: children,
	}

	hitMap := BuildHitMap(root)

	// 测试重叠中心点的查找
	results := hitMap.FindAllAt(125, 125)

	if len(results) != overlapCount+1 { // +1 for root
		t.Errorf("Expected %d nodes at overlap point, got %d", overlapCount+1, len(results))
	}

	// 验证结果按 Z-order 排序（从低到高）
	// root 的 Z-order 是 0（最低），所以应该在最前面
	if results[0].NodeID != stringToNodeID("root") {
		t.Errorf("Result[0] should be root (lowest Z-order), got %d", results[0].NodeID)
	}

	// 其余节点按照添加顺序（Z-order 递增）
	for i := 0; i < overlapCount; i++ {
		expectedID := fmt.Sprintf("node-%d", i)
		if results[i+1].NodeID != stringToNodeID(expectedID) {
			t.Errorf("Result[%d] should be %s, got %d", i+1, expectedID, results[i+1].NodeID)
		}
	}
}

// TestHitMap_HitTestDetailed_Comprehensive 详细命中测试综合测试
func TestHitMap_HitTestDetailed_Comprehensive(t *testing.T) {
	child := &mockNode{
		id:       "child",
		nodeType: "leaf",
		x:        100,
		y:        100,
		width:    50,
		height:   30,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    200,
		height:   200,
		children: []layout.Node{child},
	}

	hitMap := BuildHitMap(root)

	tests := []struct {
		name           string
		x, y           int
		expectedFound  bool
		expectedTarget string
		expectedLocalX int
		expectedLocalY int
	}{
		{
			name: "InsideChild",
			x: 125, y: 115,
			expectedFound: true,
			expectedTarget: "child",
			expectedLocalX: 25,
			expectedLocalY: 15,
		},
		{
			name: "InsideRootOnly",
			x: 50, y: 50,
			expectedFound: true,
			expectedTarget: "root",
			expectedLocalX: 50,
			expectedLocalY: 50,
		},
		{
			name: "OutsideAll",
			x: 250, y: 250,
			expectedFound: false,
			expectedTarget: "",
			expectedLocalX: 0,
			expectedLocalY: 0,
		},
		{
			name: "ChildBoundary_TopLeft",
			x: 100, y: 100,
			expectedFound: true,
			expectedTarget: "child",
			expectedLocalX: 0,
			expectedLocalY: 0,
		},
		{
			name: "ChildBoundary_BottomRight",
			x: 149, y: 129,
			expectedFound: true,
			expectedTarget: "child",
			expectedLocalX: 49,
			expectedLocalY: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hitMap.HitTestDetailed(tt.x, tt.y)

			if result.Found != tt.expectedFound {
				t.Errorf("Found: got %v, want %v", result.Found, tt.expectedFound)
			}

			if tt.expectedFound {
				if result.Entry == nil {
					t.Fatal("Expected non-nil Entry when Found=true")
				}

				if result.Entry.NodeID != stringToNodeID(tt.expectedTarget) {
					t.Errorf("Target: got %d, want %s", result.Entry.NodeID, tt.expectedTarget)
				}

				if result.LocalX != tt.expectedLocalX || result.LocalY != tt.expectedLocalY {
					t.Errorf("LocalXY: got (%d,%d), want (%d,%d)",
						result.LocalX, result.LocalY,
						tt.expectedLocalX, tt.expectedLocalY)
				}
			} else {
				if result.Entry != nil {
					t.Error("Expected nil Entry when Found=false")
				}
			}
		})
	}
}
