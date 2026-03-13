package input

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type testSearchIntent struct{}

func (testSearchIntent) IntentType() string { return "test:search" }

// TestActionHandlerInstance_KeyInput 测试 Instance.HandleAction 处理键盘输入
func TestActionHandlerInstance_KeyInput(t *testing.T) {
	vnode := New().SetPlaceholder("Enter text").SetKey("test-input")
	factory, ok := vnode.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("Input VNode should implement InstanceFactory")
	}

	inst := factory.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() should return non-nil instance")
	}

	inputInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance should be *input.Instance")
	}

	// 测试插入文本
	actions := []struct {
		actionType action.ActionType
		payload    interface{}
		expected   string
	}{
		{action.ActionInputText, "a", "a"},
		{action.ActionInputText, "b", "ab"},
		{action.ActionInputText, "c", "abc"},
	}

	for _, tc := range actions {
		t.Run(string(tc.actionType)+"_"+tc.payload.(string), func(t *testing.T) {
			if !inputInst.HandleAction(action.NewActionWithPayload(tc.actionType, tc.payload)) {
				t.Errorf("HandleAction(%q, %v) should return true", tc.actionType, tc.payload)
			}
			if inputInst.GetValue() != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, inputInst.GetValue())
			}
		})
	}
}

// TestActionHandlerInstance_Backspace 测试删除操作
func TestActionHandlerInstance_Backspace(t *testing.T) {
	vnode := New().SetValue("hello").SetKey("test")
	factory := vnode.(rtui.InstanceFactory)
	inst := factory.CreateInstance().(*Instance)

	// 初始值
	if inst.GetValue() != "hello" {
		t.Fatalf("Initial value should be 'hello', got %q", inst.GetValue())
	}

	// 测试 backspace
	actions := []struct {
		actionType action.ActionType
		expected   string
	}{
		{action.ActionBackspace, "hell"},
		{action.ActionBackspace, "hel"},
		{action.ActionBackspace, "he"},
	}

	for _, tc := range actions {
		t.Run(string(tc.actionType), func(t *testing.T) {
			if !inst.HandleAction(action.NewAction(tc.actionType)) {
				t.Errorf("HandleAction(%q) should return true", tc.actionType)
			}
			if inst.GetValue() != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, inst.GetValue())
			}
		})
	}
}

// TestActionHandlerInstance_Submit 测试 Submit (Enter) 键行为
func TestActionHandlerInstance_Submit(t *testing.T) {
	vnode := New().SetValue("test").SetKey("test")
	factory := vnode.(rtui.InstanceFactory)
	inst := factory.CreateInstance().(*Instance)

	// Enter 不应该调用什么（没有 submitIntent）
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Error("HandleAction(action.Enter) should return true")
	}
}

func TestBuilder_OnSearchSetsSubmitIntent(t *testing.T) {
	vnode := NewBuilder().Search().OnSearch(testSearchIntent{}).Build()
	submitIntent, ok := vnode.Props()[propSubmitIntent].(intent.Intent)
	if !ok {
		t.Fatal("submitIntent should be present")
	}
	if submitIntent.IntentType() != "test:search" {
		t.Fatalf("submitIntent = %q, want %q", submitIntent.IntentType(), "test:search")
	}
}

// TestActionHandlerInstance_DisabledNotReceiveInput 测试禁用状态不接受输入
func TestActionHandlerInstance_DisabledNotReceiveInput(t *testing.T) {
	vnode := New().SetDisabled(true).SetKey("test")
	factory := vnode.(rtui.InstanceFactory)
	inst := factory.CreateInstance().(*Instance)

	// 禁用状态下应该不处理输入
	if inst.HandleAction(action.NewActionWithPayload(action.ActionInputText, "a")) {
		t.Error("HandleAction should return false when disabled")
	}

	if inst.GetValue() != "" {
		t.Errorf("Value should remain empty, got %q", inst.GetValue())
	}
}

// TestActionHandlerInstance_ReadOnly 测试只读状态
func TestActionHandlerInstance_ReadOnly(t *testing.T) {
	vnode := New().SetReadOnly(true).SetKey("test")
	factory := vnode.(rtui.InstanceFactory)
	inst := factory.CreateInstance().(*Instance)

	// 只读状态下应该不处理输入
	if inst.HandleAction(action.NewActionWithPayload(action.ActionInputText, "a")) {
		t.Error("HandleAction should return false when readOnly")
	}

	if inst.GetValue() != "" {
		t.Errorf("Value should remain empty, got %q", inst.GetValue())
	}
}
