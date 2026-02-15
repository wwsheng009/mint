package testing

import (
	"testing"

	"github.com/wwsheng009/mint/framework/action"
	rt "github.com/wwsheng009/mint/runtime"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// TestActionSystemWithTestableApp 测试 Action 系统与 TestableApp 的集成
func TestActionSystemWithTestableApp(t *testing.T) {
	// 创建 Router
	router := action.NewRouter(nil)

	// 创建测试应用
	app := NewTestableApp(nil, router)

	t.Run("InjectKeySequence", func(t *testing.T) {
		// 测试键盘序列注入
		app.InjectKeySequence("abc")
		// 如果没有 panic，说明基本功能正常
	})

	t.Run("InjectEnter", func(t *testing.T) {
		app.InjectEnter()
		// ActionEnter 应该被创建并分发
	})

	t.Run("InjectTab", func(t *testing.T) {
		app.InjectTab()
		// ActionNavigateNext 应该被创建并分发
	})

	t.Run("InjectEscape", func(t *testing.T) {
		app.InjectEscape()
		// ActionCancel 应该被创建并分发
	})

	t.Run("InjectAction", func(t *testing.T) {
		act := action.NewAction(action.ActionClick)
		app.InjectAction(act)
		// Action 应该被分发
	})

	t.Run("InjectText", func(t *testing.T) {
		app.InjectText("input1", "test")
		// ActionInputText 应该被创建并分发
	})
}

// TestActionTargetWithTestableApp 测试 ActionTarget 与 TestableApp 的集成
func TestActionTargetWithTestableApp(t *testing.T) {
	// 创建 Router
	router := action.NewRouter(nil)

	// 创建一个简单的 ActionTarget
	target := &mockButton{
		id:     "test-button",
		clicked: false,
	}

	// 注册目标
	router.RegisterTarget(runtimeevent.StringToNodeID(target.id), target)

	// 创建测试应用
	app := NewTestableApp(nil, router)

	t.Run("ClickButton", func(t *testing.T) {
		// 注入鼠标点击
		app.InjectMouseClickByID(target.id, 0, 0)

		// 验证按钮被点击
		if !target.clicked {
			t.Errorf("Button should be clicked after InjectMouseClickByID")
		}
	})

	t.Run("Navigate", func(t *testing.T) {
		target.clicked = false // 重置

		// 注入导航 Action
		app.InjectAction(action.NewAction(action.ActionNavigateDown))

		// 导航操作不应该触发点击
		if target.clicked {
			t.Errorf("NavigateDown should not trigger button click")
		}
	})
}

// mockButton 模拟按钮组件
type mockButton struct {
	id      string
	clicked bool
}

func (b *mockButton) HandleAction(act *action.Action) bool {
	if act.Type == action.ActionClick {
		b.clicked = true
		return true
	}
	return false
}

func (b *mockButton) GetSupportedActions() []action.ActionType {
	return []action.ActionType{action.ActionClick}
}

func (b *mockButton) CanHandleAction(act *action.Action) bool {
	return act.Type == action.ActionClick
}

// TestActionRouterWithTree 测试 Router 树形结构支持
func TestActionRouterWithTree(t *testing.T) {
	t.Run("BuildTreeAndDispatch", func(t *testing.T) {
		// 创建树形结构
		root := &rt.LayoutNode{
			ID:   "root",
			Type: rt.NodeTypeColumn,
			Children: []*rt.LayoutNode{
				{
					ID:   "button1",
					Type: rt.NodeTypeFlex,
					Children: nil,
				},
				{
					ID:   "button2",
					Type: rt.NodeTypeFlex,
					Children: nil,
				},
			},
		}

		// 创建 Router 并设置根
		router := action.NewRouter(root)

		// 验证根节点设置
		if router.Root == nil {
			t.Errorf("Router root should not be nil")
		}

		if router.Root.ID != "root" {
			t.Errorf("Root ID should be 'root', got '%s'", router.Root.ID)
		}

		// 验证子节点数量
		if len(router.Root.Children) != 2 {
			t.Errorf("Root should have 2 children, got %d", len(router.Root.Children))
		}
	})

	t.Run("GetRoot", func(t *testing.T) {
		// 创建树
		root := &rt.LayoutNode{
			ID:   "root",
			Type: rt.NodeTypeColumn,
			Children: []*rt.LayoutNode{
				{ID: "child1", Type: rt.NodeTypeFlex},
				{ID: "child2", Type: rt.NodeTypeFlex},
			},
		}

		router := action.NewRouter(root)

		// 测试获取根节点
		gotRoot := router.GetRoot()
		if gotRoot == nil {
			t.Errorf("GetRoot() should not return nil")
		}
		if gotRoot.ID != "root" {
			t.Errorf("GetRoot().ID = %s, want 'root'", gotRoot.ID)
		}

		// 测试设置新根节点
		newRoot := &rt.LayoutNode{ID: "new-root", Type: rt.NodeTypeRow}
		router.SetRoot(newRoot)
		if router.GetRoot().ID != "new-root" {
			t.Errorf("SetRoot() failed, got ID = %s, want 'new-root'", router.GetRoot().ID)
		}
	})
}

// TestMiddlewareChainWithTestableApp 测试中间件链
func TestMiddlewareChainWithTestableApp(t *testing.T) {
	// 创建带有中间件的 Router
	router := action.NewRouter(nil)

	// 添加指标中间件
	metrics := action.NewMetricsMiddleware()
	chain := action.NewMiddlewareChain(metrics)
	router.SetMiddleware(chain)

	app := NewTestableApp(nil, router)

	t.Run("MetricsCollection", func(t *testing.T) {
		// 重置指标
		metrics.Reset()

		// 注入多个 Action
		app.InjectAction(action.NewAction(action.ActionClick))
		app.InjectAction(action.NewAction(action.ActionEnter))
		app.InjectAction(action.NewAction(action.ActionNavigateDown))

		// 验证指标
		clickCount := metrics.GetActionCount(action.ActionClick)
		if clickCount != 1 {
			t.Errorf("ActionClick count should be 1, got %d", clickCount)
		}

		enterCount := metrics.GetActionCount(action.ActionEnter)
		if enterCount != 1 {
			t.Errorf("ActionEnter count should be 1, got %d", enterCount)
		}

		downCount := metrics.GetActionCount(action.ActionNavigateDown)
		if downCount != 1 {
			t.Errorf("ActionNavigateDown count should be 1, got %d", downCount)
		}

		// 验证总计数
		allCounts := metrics.GetAllActionCounts()
		if len(allCounts) != 3 {
			t.Errorf("Should have 3 action types, got %d", len(allCounts))
		}
	})

	t.Run("MiddlewareChains", func(t *testing.T) {
		// 测试预配置的中间件链
		defaultChain := action.DefaultMiddlewareChain()
		if defaultChain == nil {
			t.Error("DefaultMiddlewareChain should not be nil")
		}

		debugChain := action.DebugMiddlewareChain()
		if debugChain == nil {
			t.Error("DebugMiddlewareChain should not be nil")
		}

		prodChain := action.ProductionMiddlewareChain()
		if prodChain == nil {
			t.Error("ProductionMiddlewareChain should not be nil")
		}

		// 验证中间件数量
		defaultCount := len(defaultChain.Middlewares())
		if defaultCount == 0 {
			t.Errorf("Default chain should have middlewares, got %d", defaultCount)
		}

		debugCount := len(debugChain.Middlewares())
		if debugCount == 0 {
			t.Errorf("Debug chain should have middlewares, got %d", debugCount)
		}

		prodCount := len(prodChain.Middlewares())
		if prodCount == 0 {
			t.Errorf("Production chain should have middlewares, got %d", prodCount)
		}

		// 生产环境中间件应该少于调试环境
		if prodCount >= defaultCount {
			t.Errorf("Production chain should have fewer middlewares than default")
		}
	})
}
