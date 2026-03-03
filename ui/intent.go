package ui

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Intent Handler Registration API
// =============================================================================

// RegisterIntent registers an intent handler for a specific intent type.
// This is a wrapper around rtui.RegisterIntent for convenience.
// Call this before ui.Run() to register global intent handlers.
func RegisterIntent[T intent.Intent](handler intent.TypedHandler[T]) func() {
	return rtui.RegisterIntent(handler)
}

// EmitIntentGlobal emits an intent through the global runtime.
func EmitIntentGlobal(i intent.Intent) intent.IntentResult {
	return rtui.EmitIntentGlobal(i)
}

// =============================================================================
// Common Intent Types (for UseState + On pattern)
// =============================================================================
// 这些 Intent 类型是通用的，无参数，具体行为由 handler 根据 ActionContext 决定
// 适用于组件内部状态管理（配合 UseState + On）
//
// 注意：IntentType 使用 "Simple*" 前缀避免与 runtime/intent/builtin.go 冲突
//
// 这些类型实现了 StayPressedIntent 接口（返回 true），用于控制按钮视觉反馈。

// SimpleToggleIntent 切换意图 - 用于切换布尔值
// 与 runtime/intent.ToggleIntent 不同：不需要 Key 参数，行为由 handler 决定
type SimpleToggleIntent struct{}

func (SimpleToggleIntent) IntentType() string { return "SimpleToggle" }
func (SimpleToggleIntent) StayPressed() bool  { return true }

// SimpleIncrementIntent 简单递增意图 - 通用递增操作
// 与 runtime/intent.IncrementIntent 不同：不需要 Key/Delta 参数，行为由 handler 决定
type SimpleIncrementIntent struct{}

func (SimpleIncrementIntent) IntentType() string { return "SimpleIncrement" }
func (SimpleIncrementIntent) StayPressed() bool  { return true }

// SimpleDecrementIntent 简单递减意图 - 通用递减操作
type SimpleDecrementIntent struct{}

func (SimpleDecrementIntent) IntentType() string { return "SimpleDecrement" }
func (SimpleDecrementIntent) StayPressed() bool  { return true }

// =============================================================================
// On Helper - 简洁的 Intent 注册 API
// =============================================================================

// handlerRegistry 用于跟踪已注册的 handlers（避免重复注册）
// 使用 sync.Map 保证并发安全
//
// 注意：这是包级别的注册表，用于防止同一 Intent 类型被重复注册。
// Intent 处理器应该在应用启动时或组件首次渲染时注册，且只注册一次。
var handlerRegistry sync.Map

// On 注册 Intent 处理器（推荐 API，无闭包风险）
//
// handler 接收 *ActionContext 参数，可以从中读取最新状态，
// 避免闭包捕获旧值的问题。
//
// # 设计原则
//
//  1. handler 不得捕获闭包状态 - 必须从 ActionContext 读取
//  2. Intent = 纯数据 - 不携带逻辑
//  3. 状态通过 ActionContext.SetState/GetState 访问
//
// # 使用示例
//
//	func Counter() ui.VNode {
//		count, setCount, _ := ui.UseStateInt(0)
//
//		// 保存状态到 Context 供 handler 读取
//		ctx := ui.GetCurrentContext()
//		if ctx != nil {
//			ctx.SetState("counter_setCount", setCount)
//		}
//
//		ui.On(ui.SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
//			// 从 Context 读取 setter 并更新
//			setCountFn, _ := ctx.GetState("counter_setCount")
//			if fn, ok := setCountFn.(func(func(int) int)); ok {
//				fn(func(c int) int { return c + 1 })
//			}
//		})
//
//		return ui.Text(fmt.Sprintf("Count: %d", count))
//	}
//
// # 重要提示
//
//   - 每个 Intent 类型只会注册一次（重复调用会被忽略）
//   - handler 应该是纯函数，不应该有副作用
//   - 状态变更应该通过 ActionContext 进行
//   - 如需控制按钮按下后的视觉反馈，Intent 可实现 StayPressedIntent 接口
func On[T intent.Intent](
	intentType T,
	handler func(*intent.ActionContext),
) {
	key := intentType.IntentType()

	// 检查是否已注册（防止重复注册）
	if _, loaded := handlerRegistry.LoadOrStore(key, true); loaded {
		return
	}

	// 注册 Intent 处理器
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
		handler(ctx)
		return intent.HandledResult()
	})
}



// =============================================================================
// StateKey for Type-Safe State Access (Phase 1.2)
// =============================================================================

// StateKey is a type-safe key for accessing state.
// Use this instead of string keys to get compile-time type checking.
type StateKey[T any] struct {
	name string
}

// NewStateKey creates a new type-safe state key.
func NewStateKey[T any](name string) StateKey[T] {
	return StateKey[T]{name: name}
}

// String returns the string representation of the key.
func (k StateKey[T]) String() string {
	return k.name
}

// Get retrieves the typed value from ActionContext.
func (k StateKey[T]) Get(ctx *intent.ActionContext, defaultValue T) T {
	return intent.GetStateAs[T](ctx, k.name, defaultValue)
}

// Set stores the value in ActionContext.
func (k StateKey[T]) Set(ctx *intent.ActionContext, value T) {
	ctx.SetState(k.name, value)
}
