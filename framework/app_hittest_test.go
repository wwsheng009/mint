package framework

import (
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// mockLayoutNode 是一个简单的 layout.Node 模拟实现
type mockLayoutNode struct {
	id       string
	nodeType string
	children []layout.Node
	x, y     int
	width    int
	height   int
}

func (m *mockLayoutNode) ID() string                 { return m.id }
func (m *mockLayoutNode) Type() string               { return m.nodeType }
func (m *mockLayoutNode) Children() []layout.Node  { return m.children }
func (m *mockLayoutNode) GetPosition() (int, int) { return m.x, m.y }
func (m *mockLayoutNode) SetPosition(x, y int)     { m.x, m.y = x, y }
func (m *mockLayoutNode) GetSize() (int, int)      { return m.width, m.height }
func (m *mockLayoutNode) SetSize(w, h int)          { m.width, m.height = w, h }
func (m *mockLayoutNode) GetWidth() int             { return m.width }
func (m *mockLayoutNode) GetHeight() int            { return m.height }

// mockPaintableNode 是一个实现 Paintable 的模拟节点
type mockPaintableNode struct {
	*mockLayoutNode
}

func (m *mockPaintableNode) SetChildren(children []layout.Node) {
	m.children = children
}

// TestApp_HitMapIntegration 测试 App 与 HitMap 的集成
func TestApp_HitMapIntegration(t *testing.T) {
	// 跳过测试如果环境变量 TUI_SKIP_INTEGRATION_TESTS 设置
	if os.Getenv("TUI_SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test (TUI_SKIP_INTEGRATION_TESTS=true)")
	}

	// 创建一个简单的布局树作为 App 的 root
	child1 := &mockLayoutNode{
		id:       "child-1",
		nodeType: "label",
		x:        10,
		y:        10,
		width:    20,
		height:   5,
		children: nil,
	}

	child2 := &mockLayoutNode{
		id:       "child-2",
		nodeType: "button",
		x:        40,
		y:        10,
		width:    15,
		height:   5,
		children: nil,
	}

	root := &mockPaintableNode{
		mockLayoutNode: &mockLayoutNode{
			id:       "root",
			nodeType: "container",
			x:        0,
			y:        0,
			width:    80,
			height:   25,
			children: []layout.Node{child1, child2},
		},
	}

	// 创建 App 并设置 root
	app := NewApp()
	app.root = root

	// 强制渲染（这会构建 HitMap）
	os.Setenv("TUI_DEBUG_HITMAP", "false") // 禁用调试输出

	// 由于 render() 方法需要终端，我们无法直接调用它
	// 但我们可以测试 GetHitMap() 方法
	t.Run("GetHitMap_ReturnsNilBeforeRender", func(t *testing.T) {
		// 在渲染前，hitMap 应该为 nil
		if app.hitMap != nil {
			t.Error("hitMap should be nil before render")
		}
	})

	t.Run("GetHitMap_AfterMockBuild", func(t *testing.T) {
		// 手动构建 HitMap（模拟 render() 后的状态）
		app.hitMap = runtimeevent.BuildHitMap(root)

		// 验证 GetHitMap 返回正确的 HitMap
		hitMap := app.GetHitMap()
		if hitMap == nil {
			t.Fatal("GetHitMap returned nil")
		}

		// 验证 HitMap 包含所有节点
		expectedSize := 3 // root + child1 + child2
		if hitMap.Size() != expectedSize {
			t.Errorf("HitMap size incorrect, got %d, want %d", hitMap.Size(), expectedSize)
		}

		// 验证每个节点都在 HitMap 中
		expectedIDs := []string{"root", "child-1", "child-2"}
		for _, expectedID := range expectedIDs {
			found := false
			entries := hitMap.AllEntries()
			for _, entry := range entries {
				if entry.Node.ID() == expectedID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Node %s not found in HitMap", expectedID)
			}
		}
	})
}

// TestApp_HitMapHitTest 测试通过 App 获取的 HitMap 的命中测试功能
func TestApp_HitMapHitTest(t *testing.T) {
	if os.Getenv("TUI_SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test (TUI_SKIP_INTEGRATION_TESTS=true)")
	}

	// 创建布局树
	child := &mockLayoutNode{
		id:       "button",
		nodeType: "button",
		x:        20,
		y:        15,
		width:    30,
		height:   8,
		children: nil,
	}

	root := &mockPaintableNode{
		mockLayoutNode: &mockLayoutNode{
			id:       "container",
			nodeType: "container",
			x:        10,
			y:        10,
			width:    50,
			height:   30,
			children: []layout.Node{child},
		},
	}

	// 创建 App
	app := NewApp()
	app.root = root

	// 手动构建 HitMap
	app.hitMap = runtimeevent.BuildHitMap(root)

	t.Run("HitTest_InsideChild", func(t *testing.T) {
		hitMap := app.GetHitMap()
		if hitMap == nil {
			t.Fatal("HitMap is nil")
		}

		// 测试命中 child 节点内的点
		entry := hitMap.HitTest(30, 20)
		if entry == nil {
			t.Error("Expected to hit child, but got nil")
			return
		}

		if entry.Node.ID() != "button" {
			t.Errorf("Expected to hit 'button', got %s", entry.Node.ID())
		}
	})

	t.Run("HitTest_InsideRoot", func(t *testing.T) {
		hitMap := app.GetHitMap()
		if hitMap == nil {
			t.Fatal("HitMap is nil")
		}

		// 测试命中 root 但不在 child 内的点
		entry := hitMap.HitTest(12, 12)
		if entry == nil {
			t.Error("Expected to hit root, but got nil")
			return
		}

		if entry.Node.ID() != "container" {
			t.Errorf("Expected to hit 'container', got %s", entry.Node.ID())
		}
	})

	t.Run("HitTest_Outside", func(t *testing.T) {
		hitMap := app.GetHitMap()
		if hitMap == nil {
			t.Fatal("HitMap is nil")
		}

		// 测试命中范围外的点
		entry := hitMap.HitTest(100, 100)
		if entry != nil {
			t.Errorf("Expected nil (miss), got %s", entry.Node.ID())
		}
	})
}
