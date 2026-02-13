package event

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
)

// mockNode 是 layout.Node 的模拟实现
type mockNode struct {
	id       string
	nodeType string
	x, y     int
	width    int
	height   int
	children []layout.Node
}

func (m *mockNode) ID() string        { return m.id }
func (m *mockNode) Type() string      { return m.nodeType }
func (m *mockNode) Children() []layout.Node {
	if m.children == nil {
		return []layout.Node{}
	}
	return m.children
}
func (m *mockNode) GetPosition() (int, int) { return m.x, m.y }
func (m *mockNode) SetPosition(x, y int)   { m.x, m.y = x, y }
func (m *mockNode) GetSize() (int, int)     { return m.width, m.height }
func (m *mockNode) SetSize(w, h int)       { m.width, m.height = w, h }
func (m *mockNode) GetWidth() int           { return m.width }
func (m *mockNode) GetHeight() int          { return m.height }

// TestHitMap_Build 测试 HitMap 构建
func TestHitMap_Build(t *testing.T) {
	// 创建测试布局树
	//
	// root (0,0, 100x100)
	//   ├── child1 (0,0, 50x50)
	//   │   └── grandchild1 (10,10, 30x30)
	//   └── child2 (50,0, 50x50)

	grandchild1 := &mockNode{
		id:       "grandchild1",
		nodeType: "leaf",
		x:        10,
		y:        10,
		width:    30,
		height:   30,
	}

	child1 := &mockNode{
		id:       "child1",
		nodeType: "container",
		x:        0,
		y:        0,
		width:    50,
		height:   50,
		children: []layout.Node{grandchild1},
	}

	child2 := &mockNode{
		id:       "child2",
		nodeType: "leaf",
		x:        50,
		y:        0,
		width:    50,
		height:   50,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    100,
		height:   100,
		children: []layout.Node{child1, child2},
	}

	// 构建 HitMap
	hitMap := BuildHitMap(root)

	// 验证基本属性
	if hitMap == nil {
		t.Fatal("BuildHitMap returned nil")
	}

	if hitMap.Size() != 4 {
		t.Errorf("Expected 4 entries, got %d", hitMap.Size())
	}

	if hitMap.IsEmpty() {
		t.Error("HitMap should not be empty")
	}

	// 验证根节点
	rootEntry := hitMap.FindByID(stringToNodeID("root"))
	if rootEntry == nil {
		t.Fatal("root entry not found")
	}

	if rootEntry.Bounds.X != 0 || rootEntry.Bounds.Y != 0 {
		t.Errorf("root bounds incorrect, got (0,0), want (%d,%d)",
			rootEntry.Bounds.X, rootEntry.Bounds.Y)
	}

	if rootEntry.Bounds.Width != 100 || rootEntry.Bounds.Height != 100 {
		t.Errorf("root size incorrect, got %dx%d, want 100x100",
			rootEntry.Bounds.Width, rootEntry.Bounds.Height)
	}

	// 验证 Z-order
	if rootEntry.ZOrder != 0 {
		t.Errorf("root ZOrder should be 0, got %d", rootEntry.ZOrder)
	}

	// 验证子节点
	child1Entry := hitMap.FindByID(stringToNodeID("child1"))
	if child1Entry == nil {
		t.Fatal("child1 entry not found")
	}

	if child1Entry.ZOrder != 1 {
		t.Errorf("child1 ZOrder should be 1, got %d", child1Entry.ZOrder)
	}

	grandchild1Entry := hitMap.FindByID(stringToNodeID("grandchild1"))
	if grandchild1Entry == nil {
		t.Fatal("grandchild1 entry not found")
	}

	if grandchild1Entry.ZOrder != 2 {
		t.Errorf("grandchild1 ZOrder should be 2, got %d", grandchild1Entry.ZOrder)
	}
}

// TestHitMap_HitTest 测试命中测试
func TestHitMap_HitTest(t *testing.T) {
	// 创建简单的布局树
	child := &mockNode{
		id:       "child",
		nodeType: "leaf",
		x:        10,
		y:        10,
		width:    30,
		height:   20,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    50,
		height:   50,
		children: []layout.Node{child},
	}

	hitMap := BuildHitMap(root)

	// 测试命中 root 范围内的点（注意：会命中 Z-order 最高的节点）
	t.Run("HitRoot", func(t *testing.T) {
		entry := hitMap.HitTest(25, 25)
		if entry == nil {
			t.Error("Expected to hit something, but got nil")
			return
		}

		// 由于 child 的 Z-order 更高，会命中 child 而不是 root
		// 这才是正确的行为
		if entry.NodeID != stringToNodeID("child") {
			t.Errorf("Expected child (higher Z-order), got %d", entry.NodeID)
		}
	})

	// 测试命中 child 范围内的点
	t.Run("HitChild", func(t *testing.T) {
		entry := hitMap.HitTest(15, 15)
		if entry == nil {
			t.Error("Expected to hit child, but got nil")
			return
		}

		if entry.NodeID != stringToNodeID("child") {
			t.Errorf("Expected child, got %d", entry.NodeID)
		}
	})

	// 测试未命中的点
	t.Run("Miss", func(t *testing.T) {
		entry := hitMap.HitTest(100, 100)
		if entry != nil {
			t.Errorf("Expected nil (miss), got %d", entry.NodeID)
		}
	})

	// 测试边界上的点（应该命中）
	t.Run("Boundary", func(t *testing.T) {
		// child 的边界是 (10,10) 到 (40,30)
		// 测试左上角
		entry := hitMap.HitTest(10, 10)
		if entry == nil || entry.NodeID != stringToNodeID("child") {
			t.Error("Boundary point (10,10) should hit child")
		}

		// 测试右下角
		entry = hitMap.HitTest(39, 29)
		if entry == nil || entry.NodeID != stringToNodeID("child") {
			t.Error("Boundary point (39,29) should hit child")
		}
	})
}

// TestHitMap_ZOrder 测试 Z-order 排序和命中优先级
func TestHitMap_ZOrder(t *testing.T) {
	// 创建重叠的节点
	// child1 和 child2 都在 (10,10)，但 child2 的 Z-order 更高
	child1 := &mockNode{
		id:       "child1",
		nodeType: "leaf1",
		x:        10,
		y:        10,
		width:    30,
		height:   20,
	}

	child2 := &mockNode{
		id:       "child2",
		nodeType: "leaf2",
		x:        10,
		y:        10,
		width:    30,
		height:   20,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    50,
		height:   50,
		children: []layout.Node{child1, child2}, // child2 后添加，Z-order 更高
	}

	hitMap := BuildHitMap(root)

	// 测试重叠区域的命中
	// 应该命中 Z-order 更高的 child2
	t.Run("HigherZOrderWins", func(t *testing.T) {
		entry := hitMap.HitTest(15, 15)
		if entry == nil {
			t.Fatal("Expected to hit something, but got nil")
		}

		if entry.NodeID != stringToNodeID("child2") {
			t.Errorf("Expected child2 (higher Z-order), got %d", entry.NodeID)
		}
	})
}

// TestHitMap_FindByID 测试按 ID 查找
func TestHitMap_FindByID(t *testing.T) {
	node := &mockNode{
		id:       "test-node",
		nodeType: "test",
		x:        5,
		y:        5,
		width:    20,
		height:   15,
	}

	hitMap := BuildHitMap(node)

	// 测试查找存在的节点
	t.Run("Found", func(t *testing.T) {
		entry := hitMap.FindByID(stringToNodeID("test-node"))
		if entry == nil {
			t.Fatal("Expected to find test-node")
		}

		if entry.Bounds.X != 5 || entry.Bounds.Y != 5 {
			t.Errorf("Bounds incorrect, got (%d,%d), want (5,5)",
				entry.Bounds.X, entry.Bounds.Y)
		}
	})

	// 测试查找不存在的节点
	t.Run("NotFound", func(t *testing.T) {
		entry := hitMap.FindByID(stringToNodeID("non-existent"))
		if entry != nil {
			t.Errorf("Expected nil for non-existent node, got %v", entry)
		}
	})
}

// TestHitMap_LocalXY 测试局部坐标转换
func TestHitMap_LocalXY(t *testing.T) {
	// 节点位于 (10, 20)，尺寸 50x30
	node := &mockNode{
		id:       "test-node",
		nodeType: "test",
		x:        10,
		y:        20,
		width:    50,
		height:   30,
	}

	hitMap := BuildHitMap(node)
	entry := hitMap.FindByID(stringToNodeID("test-node"))

	if entry == nil {
		t.Fatal("test-node not found in HitMap")
	}

	// 测试局部坐标转换
	t.Run("InsideNode", func(t *testing.T) {
		// 屏幕坐标 (20, 30) 应该转换为局部坐标 (10, 10)
		localX, localY := entry.LocalXY(20, 30)

		if localX != 10 || localY != 10 {
			t.Errorf("LocalXY incorrect, got (%d,%d), want (10,10)",
				localX, localY)
		}
	})

	// 测试边界外的坐标
	t.Run("OutsideNode", func(t *testing.T) {
		// 屏幕坐标 (5, 5) 在节点外部
		// 应该转换为负坐标
		localX, localY := entry.LocalXY(5, 5)

		if localX != -5 || localY != -15 {
			t.Errorf("LocalXY incorrect for outside point, got (%d,%d), want (-5,-15)",
				localX, localY)
		}
	})
}

// TestHitMap_FindAllAt 测试查找所有包含某点的节点
func TestHitMap_FindAllAt(t *testing.T) {
	// 创建嵌套节点
	inner := &mockNode{
		id:       "inner",
		nodeType: "leaf",
		x:        20,
		y:        20,
		width:    10,
		height:   10,
	}

	outer := &mockNode{
		id:       "outer",
		nodeType: "container",
		x:        10,
		y:        10,
		width:    30,
		height:   30,
		children: []layout.Node{inner},
	}

	hitMap := BuildHitMap(outer)

	// 测试重叠区域
	t.Run("NestedNodes", func(t *testing.T) {
		// 点 (25, 25) 同时在 outer 和 inner 内
		results := hitMap.FindAllAt(25, 25)

		if len(results) != 2 {
			t.Fatalf("Expected 2 nodes, got %d", len(results))
		}

		// 结果应该按 Z-order 排序
		if results[0].NodeID != stringToNodeID("outer") {
			t.Errorf("First result should be outer, got %d", results[0].NodeID)
		}

		if results[1].NodeID != stringToNodeID("inner") {
			t.Errorf("Second result should be inner, got %d", results[1].NodeID)
		}
	})
}

// TestHitMap_DetailedHitTest 测试详细命中测试
func TestHitMap_DetailedHitTest(t *testing.T) {
	node := &mockNode{
		id:       "test-node",
		nodeType: "test",
		x:        10,
		y:        20,
		width:    50,
		height:   30,
	}

	hitMap := BuildHitMap(node)

	// 测试命中情况
	t.Run("Hit", func(t *testing.T) {
		result := hitMap.HitTestDetailed(30, 35)

		if !result.Found {
			t.Fatal("Expected Found=true, got false")
		}

		if result.Entry == nil {
			t.Fatal("Expected non-nil Entry")
		}

		if result.Entry.NodeID != stringToNodeID("test-node") {
			t.Errorf("Expected test-node, got %d", result.Entry.NodeID)
		}

		// 验证局部坐标
		// 屏幕坐标 (30, 35) - 节点位置 (10, 20) = (20, 15)
		if result.LocalX != 20 || result.LocalY != 15 {
			t.Errorf("LocalXY incorrect, got (%d,%d), want (20,15)",
				result.LocalX, result.LocalY)
		}
	})

	// 测试未命中情况
	t.Run("Miss", func(t *testing.T) {
		result := hitMap.HitTestDetailed(100, 100)

		if result.Found {
			t.Error("Expected Found=false, got true")
		}

		if result.Entry != nil {
			t.Error("Expected nil Entry for miss")
		}
	})
}

// TestHitMap_InvalidNodes 测试无效节点的处理
func TestHitMap_InvalidNodes(t *testing.T) {
	// 创建无效节点（宽或高为0）
	invalidChild := &mockNode{
		id:       "invalid-child",
		nodeType: "invalid",
		x:        0,
		y:        0,
		width:    0,  // 无效尺寸
		height:   0,
	}

	validChild := &mockNode{
		id:       "valid-child",
		nodeType: "valid",
		x:        10,
		y:        10,
		width:    20,
		height:   20,
	}

	root := &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    50,
		height:   50,
		children: []layout.Node{invalidChild, validChild},
	}

	hitMap := BuildHitMap(root)

	// 无效节点不应该被包含在 HitMap 中
	if entry := hitMap.FindByID(stringToNodeID("invalid-child")); entry != nil {
		t.Error("Invalid node should not be in HitMap")
	}

	// 有效节点应该被包含
	if entry := hitMap.FindByID(stringToNodeID("valid-child")); entry == nil {
		t.Error("Valid node should be in HitMap")
	}

	// 验证总节点数（应该是 root + valid-child，不包括 invalid-child）
	if hitMap.Size() != 2 {
		t.Errorf("Expected 2 entries, got %d", hitMap.Size())
	}
}

// TestHitMap_NilRoot 测试 nil 根节点
func TestHitMap_NilRoot(t *testing.T) {
	hitMap := BuildHitMap(nil)

	if hitMap == nil {
		t.Fatal("BuildHitMap should return empty HitMap for nil root")
	}

	if !hitMap.IsEmpty() {
		t.Error("HitMap should be empty for nil root")
	}

	if hitMap.Size() != 0 {
		t.Errorf("Expected 0 entries, got %d", hitMap.Size())
	}
}

// TestHitMap_Dump 测试调试输出
func TestHitMap_Dump(t *testing.T) {
	node := &mockNode{
		id:       "test-node",
		nodeType: "test",
		x:        5,
		y:        10,
		width:    20,
		height:   15,
	}

	hitMap := BuildHitMap(node)
	dump := hitMap.Dump()

	if dump == "" {
		t.Error("Dump should not be empty")
	}

	// 计算预期的 NodeID (hash of "test-node")
	expectedNodeID := stringToNodeID("test-node")

	// 验证关键信息存在
	requiredStrings := []string{
		"=== HitMap ===",
		fmt.Sprintf("%d", expectedNodeID), // NodeID 现在是数字
		"{5 10 20 15}",  // Rect 格式: {X Y Width Height}
		"Z:0",
	}

	for _, s := range requiredStrings {
		if !contains(dump, s) {
			t.Errorf("Dump should contain %q", s)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
