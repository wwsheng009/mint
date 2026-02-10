package action

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
)

// MockCaptureHandler 是用于测试的捕获处理器
type MockCaptureHandler struct {
	PriorityValue int
	ShouldStop    bool
	Called        bool
}

func (m *MockCaptureHandler) HandleCapture(act *Action, target *runtime.LayoutNode) bool {
	m.Called = true
	return m.ShouldStop
}

func (m *MockCaptureHandler) Priority() int {
	return m.PriorityValue
}

// MockBubbleHandler 是用于测试的冒泡处理器
type MockBubbleHandler struct {
	ShouldStop bool
	Called     bool
}

func (m *MockBubbleHandler) HandleBubble(act *Action, target *runtime.LayoutNode) bool {
	m.Called = true
	return m.ShouldStop
}

// MockActionTarget 是用于测试的目标组件
type MockActionTarget struct {
	SupportedActions []ActionType
	CanHandle        bool
	HandleActionCalled bool
	HandledValue      bool
}

func (m *MockActionTarget) HandleAction(act *Action) bool {
	m.HandleActionCalled = true
	return m.HandledValue
}

func (m *MockActionTarget) GetSupportedActions() []ActionType {
	return m.SupportedActions
}

func (m *MockActionTarget) CanHandleAction(act *Action) bool {
	return m.CanHandle
}

// TestRouter_NewRouter 测试 Router 构造函数
func TestRouter_NewRouter(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	router := NewRouter(root)

	if router == nil {
		t.Fatal("NewRouter() should not return nil")
	}

	if router.Root != root {
		t.Errorf("Root should be set to the provided node")
	}

	if router.CaptureHandlers == nil {
		t.Error("CaptureHandlers should be initialized")
	}

	if router.BubbleHandlers == nil {
		t.Error("BubbleHandlers should be initialized")
	}

	if router.TargetHandlers == nil {
		t.Error("TargetHandlers should be initialized")
	}
}

// TestRouter_AddCaptureHandler 测试添加捕获处理器
func TestRouter_AddCaptureHandler(t *testing.T) {
	router := NewRouter(nil)

	handler1 := &MockCaptureHandler{PriorityValue: 10}
	handler2 := &MockCaptureHandler{PriorityValue: 20}
	handler3 := &MockCaptureHandler{PriorityValue: 5}

	router.AddCaptureHandler(handler1, "handler1")
	router.AddCaptureHandler(handler2, "handler2")
	router.AddCaptureHandler(handler3, "handler3")

	// 检查数量
	if len(router.CaptureHandlers) != 3 {
		t.Errorf("Expected 3 capture handlers, got %d", len(router.CaptureHandlers))
	}

	// 检查排序（应该是从高到低：20, 10, 5）
	if router.CaptureHandlers[0].Priority != 20 {
		t.Errorf("First handler should have priority 20, got %d", router.CaptureHandlers[0].Priority)
	}
	if router.CaptureHandlers[1].Priority != 10 {
		t.Errorf("Second handler should have priority 10, got %d", router.CaptureHandlers[1].Priority)
	}
	if router.CaptureHandlers[2].Priority != 5 {
		t.Errorf("Third handler should have priority 5, got %d", router.CaptureHandlers[2].Priority)
	}
}

// TestRouter_RemoveCaptureHandler 测试移除捕获处理器
func TestRouter_RemoveCaptureHandler(t *testing.T) {
	router := NewRouter(nil)

	handler1 := &MockCaptureHandler{PriorityValue: 10}
	handler2 := &MockCaptureHandler{PriorityValue: 20}

	router.AddCaptureHandler(handler1, "handler1")
	router.AddCaptureHandler(handler2, "handler2")

	// 移除一个
	router.RemoveCaptureHandler("handler1")

	if len(router.CaptureHandlers) != 1 {
		t.Errorf("Expected 1 capture handler after removal, got %d", len(router.CaptureHandlers))
	}

	if router.CaptureHandlers[0].ID != "handler2" {
		t.Errorf("Remaining handler should be handler2")
	}
}

// TestRouter_CapturePhase 测试捕获阶段
func TestRouter_CapturePhase(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	router := NewRouter(root)

	// 添加捕获处理器
	handler1 := &MockCaptureHandler{PriorityValue: 10, ShouldStop: false}
	handler2 := &MockCaptureHandler{PriorityValue: 20, ShouldStop: true}

	router.AddCaptureHandler(handler1, "handler1")
	router.AddCaptureHandler(handler2, "handler2")

	act := NewAction(ActionClick)
	act.TargetID = "root"

	result := router.Dispatch(act)

	// 验证结果
	if !result.Handled {
		t.Error("Result should be marked as handled")
	}
	if !result.Stopped {
		t.Error("Result should be marked as stopped")
	}
	if result.Phase != ActionPhaseCapture {
		t.Errorf("Phase should be Capture, got %s", result.Phase)
	}

	// 验证只有 handler2 被调用（优先级高且停止传播）
	if !handler2.Called {
		t.Error("handler2 should be called")
	}
	if handler1.Called {
		t.Error("handler1 should not be called (stopped by handler2)")
	}
}

// TestRouter_CapturePhase_NoStop 测试捕获阶段不停止
func TestRouter_CapturePhase_NoStop(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	router := NewRouter(root)

	handler1 := &MockCaptureHandler{PriorityValue: 10, ShouldStop: false}
	handler2 := &MockCaptureHandler{PriorityValue: 20, ShouldStop: false}

	router.AddCaptureHandler(handler1, "handler1")
	router.AddCaptureHandler(handler2, "handler2")

	act := NewAction(ActionClick)
	act.TargetID = "root"

	result := router.Dispatch(act)

	// 两个处理器都应该被调用
	if !handler1.Called {
		t.Error("handler1 should be called")
	}
	if !handler2.Called {
		t.Error("handler2 should be called")
	}

	// 结果应该不是 handled（因为没有处理器停止）
	if result.Handled {
		t.Error("Result should not be marked as handled")
	}
	if result.Stopped {
		t.Error("Result should not be marked as stopped")
	}
}

// TestRouter_TargetPhase 测试目标阶段
func TestRouter_TargetPhase(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	target := &runtime.LayoutNode{
		ID: "target1",
		Parent: root,
		Children: []*runtime.LayoutNode{},
	}
	root.Children = append(root.Children, target)

	router := NewRouter(root)

	// 注册目标处理器
	targetHandler := &MockActionTarget{
		CanHandle:   true,
		HandledValue: true,
		SupportedActions: []ActionType{ActionClick},
	}
	router.RegisterTarget("target1", targetHandler)

	act := NewAction(ActionClick)
	act.TargetID = "target1"

	result := router.Dispatch(act)

	// 验证结果
	if !result.Handled {
		t.Error("Result should be marked as handled")
	}
	if result.Phase != ActionPhaseTarget {
		t.Errorf("Phase should be Target, got %s", result.Phase)
	}

	// 验证目标被调用
	if !targetHandler.HandleActionCalled {
		t.Error("Target handler should be called")
	}
}

// TestRouter_TargetPhase_CannotHandle 测试目标无法处理
func TestRouter_TargetPhase_CannotHandle(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	target := &runtime.LayoutNode{
		ID: "target1",
		Parent: root,
		Children: []*runtime.LayoutNode{},
	}
	root.Children = append(root.Children, target)

	router := NewRouter(root)

	// 注册目标处理器（但不能处理该 Action）
	targetHandler := &MockActionTarget{
		CanHandle:   false,
		HandledValue: true,
		SupportedActions: []ActionType{},
	}
	router.RegisterTarget("target1", targetHandler)

	act := NewAction(ActionClick)
	act.TargetID = "target1"

	result := router.Dispatch(act)

	// 验证结果
	if result.Handled {
		t.Error("Result should not be marked as handled")
	}

	// 验证目标没有被调用
	if targetHandler.HandleActionCalled {
		t.Error("Target handler should not be called (CanHandleAction returned false)")
	}
}

// TestRouter_BubblePhase 测试冒泡阶段
func TestRouter_BubblePhase(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	target := &runtime.LayoutNode{
		ID: "target1",
		Parent: root,
		Children: []*runtime.LayoutNode{},
	}
	root.Children = append(root.Children, target)

	router := NewRouter(root)

	// 注册冒泡处理器
	bubbleHandler := &MockBubbleHandler{ShouldStop: true}
	router.AddBubbleHandler(bubbleHandler, "bubble1")

	act := NewAction(ActionClick)
	act.TargetID = "target1"

	result := router.Dispatch(act)

	// 验证结果
	if !result.Handled {
		t.Error("Result should be marked as handled")
	}
	if !result.Stopped {
		t.Error("Result should be marked as stopped")
	}
	if result.Phase != ActionPhaseBubble {
		t.Errorf("Phase should be Bubble, got %s", result.Phase)
	}

	// 验证冒泡处理器被调用
	if !bubbleHandler.Called {
		t.Error("Bubble handler should be called")
	}
}

// TestRouter_StopPropagation 测试停止传播
func TestRouter_StopPropagation(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	router := NewRouter(root)

	// 添加捕获处理器（会停止传播）
	captureHandler := &MockCaptureHandler{PriorityValue: 10, ShouldStop: true}
	router.AddCaptureHandler(captureHandler, "capture1")

	// 添加冒泡处理器（不应该被调用）
	bubbleHandler := &MockBubbleHandler{ShouldStop: false}
	router.AddBubbleHandler(bubbleHandler, "bubble1")

	act := NewAction(ActionClick)
	act.TargetID = "root"

	result := router.Dispatch(act)

	// 验证捕获阶段停止传播
	if result.Phase != ActionPhaseCapture {
		t.Errorf("Should stop at Capture phase, got %s", result.Phase)
	}

	// 验证捕获处理器被调用
	if !captureHandler.Called {
		t.Error("Capture handler should be called")
	}

	// 验证冒泡处理器没有被调用
	if bubbleHandler.Called {
		t.Error("Bubble handler should not be called (propagation stopped)")
	}
}

// TestRouter_FindNodeByID 测试节点查找
func TestRouter_FindNodeByID(t *testing.T) {
	// 构建测试树：
	//     root
	//    /    \
	//   a      b
	//  / \
	// a1  a2

	b := &runtime.LayoutNode{ID: "b"}
	a1 := &runtime.LayoutNode{ID: "a1"}
	a2 := &runtime.LayoutNode{ID: "a2"}
	a := &runtime.LayoutNode{
		ID: "a",
		Children: []*runtime.LayoutNode{a1, a2},
	}

	root := &runtime.LayoutNode{
		ID: "root",
		Children: []*runtime.LayoutNode{a, b},
	}

	a.Parent = root
	b.Parent = root
	a1.Parent = a
	a2.Parent = a

	router := NewRouter(root)

	// 测试查找各个节点
	tests := []struct {
		id   string
		want bool
	}{
		{"root", true},
		{"a", true},
		{"b", true},
		{"a1", true},
		{"a2", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			node := router.findNodeByID(tt.id)
			found := (node != nil)
			if found != tt.want {
				t.Errorf("findNodeByID(%s) found = %v, want %v", tt.id, found, tt.want)
			}
			if found && node.ID != tt.id {
				t.Errorf("findNodeByID(%s) returned node with ID %s", tt.id, node.ID)
			}
		})
	}
}

// TestRouter_BuildTargetRegistry 测试构建目标注册表
func TestRouter_BuildTargetRegistry(t *testing.T) {
	// 构建测试树，其中一些节点实现了 ActionTarget
	target1 := &MockActionTarget{
		SupportedActions: []ActionType{ActionClick},
		CanHandle: true,
	}

	target2 := &MockActionTarget{
		SupportedActions: []ActionType{ActionNavigateDown},
		CanHandle: true,
	}

	node1 := &runtime.LayoutNode{
		ID: "node1",
	}
	node1.Component = &runtime.ComponentRef{
		Instance: target1,
	}

	node2 := &runtime.LayoutNode{
		ID: "node2",
	}
	node2.Component = &runtime.ComponentRef{
		Instance: target2,
	}

	nonTargetNode := &runtime.LayoutNode{
		ID: "nonTarget",
	}

	root := &runtime.LayoutNode{
		ID: "root",
		Children: []*runtime.LayoutNode{node1, node2, nonTargetNode},
	}

	router := NewRouter(root)

	// 构建注册表
	router.BuildTargetRegistry()

	// 验证注册表
	if len(router.TargetHandlers) != 2 {
		t.Errorf("Expected 2 target handlers, got %d", len(router.TargetHandlers))
	}

	// 检查是否注册了正确的目标
	if _, exists := router.TargetHandlers["node1"]; !exists {
		t.Error("node1 should be registered")
	}
	if _, exists := router.TargetHandlers["node2"]; !exists {
		t.Error("node2 should be registered")
	}
	if _, exists := router.TargetHandlers["nonTarget"]; exists {
		t.Error("nonTarget should not be registered (not an ActionTarget)")
	}
}

// TestRouter_Dispatch_NoTarget 测试没有目标的分发
func TestRouter_Dispatch_NoTarget(t *testing.T) {
	root := &runtime.LayoutNode{
		ID: "root",
	}

	router := NewRouter(root)

	// 添加冒泡处理器
	bubbleHandler := &MockBubbleHandler{ShouldStop: false}
	router.AddBubbleHandler(bubbleHandler, "bubble1")

	act := NewAction(ActionQuit)
	// 没有 TargetID

	result := router.Dispatch(act)

	// 验证冒泡处理器仍然被调用
	if !bubbleHandler.Called {
		t.Error("Bubble handler should be called even without TargetID")
	}

	if result.Phase != ActionPhaseBubble {
		t.Errorf("Phase should be Bubble, got %s", result.Phase)
	}
}

// TestRouter_ActionPhase_String 测试阶段字符串表示
func TestRouter_ActionPhase_String(t *testing.T) {
	tests := []struct {
		phase ActionPhase
		want  string
	}{
		{ActionPhaseNone, "None"},
		{ActionPhaseCapture, "Capture"},
		{ActionPhaseTarget, "Target"},
		{ActionPhaseBubble, "Bubble"},
		{ActionPhase(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.want {
				t.Errorf("ActionPhase.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRouter_RegisterAndUnregister 测试注册和注销
func TestRouter_RegisterAndUnregister(t *testing.T) {
	router := NewRouter(nil)

	target := &MockActionTarget{
		SupportedActions: []ActionType{ActionClick},
		CanHandle: true,
	}

	// 注册
	router.RegisterTarget("test", target)

	if _, exists := router.TargetHandlers["test"]; !exists {
		t.Error("Target should be registered")
	}

	// 注销
	router.UnregisterTarget("test")

	if _, exists := router.TargetHandlers["test"]; exists {
		t.Error("Target should be unregistered")
	}
}
