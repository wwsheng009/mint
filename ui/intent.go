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
// 这些 Intent 类型是通用的，无参数，具体行为由 handler 闭包决定
// 适用于组件内部状态管理（配合 UseState + On）
//
// 注意：IntentType 使用 "Simple*" 前缀避免与 runtime/intent/builtin.go 冲突

// SimpleToggleIntent 切换意图 - 用于切换布尔值
// 与 runtime/intent.ToggleIntent 不同：不需要 Key 参数，行为由 handler 闭包决定
type SimpleToggleIntent struct{}

func (SimpleToggleIntent) IntentType() string { return "SimpleToggle" }
func (SimpleToggleIntent) StayPressed() bool  { return true }

// SimpleIncrementIntent 简单递增意图 - 通用递增操作
// 与 runtime/intent.IncrementIntent 不同：不需要 Key/Delta 参数，行为由 handler 闭包决定
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

// registeredHandlers 用于跟踪已注册的 handlers（避免重复注册）
// 使用 sync.Map 保证并发安全
var registeredHandlers sync.Map

// On 注册 Intent 处理器的通用实现（简洁 API）
//
// ⚠️ 警告：此 API 存在闭包过期风险！
//
// handler 闭包会捕获定义时的状态值，如果组件重新渲染，
// handler 内访问的变量可能是旧值。推荐使用 OnWithContext 替代。
//
// 使用示例（使用内置 Simple* Intent 类型）：
//
//	func MyComponent() ui.VNode {
//		visible, setVisible, _ := ui.UseStateBool(false)
//
//		ui.On(ui.SimpleToggleIntent{}, func() {
//			setVisible(!visible)  // ⚠️ visible 可能是旧值
//		})
//
//		return ui.NewButtonBuilder("Toggle").OnPress(ui.SimpleToggleIntent{}).Build()
//	}
//
// 重要提示：
// - 对于 UseBool/UseInt 等简单状态，推荐使用函数形式的 setter（避免闭包捕获旧值）
// - handler 闭包可以直接访问组件内的局部变量（通过闭包）
// - 每个组件渲染时只注册一次 handler（通过包级 map 防止重复注册）
//
// 废弃警告：建议迁移到 OnWithContext 以避免闭包问题
func On[T interface{ IntentType() string; StayPressed() bool }](
	intentType T, // 例如 ui.SimpleIncrementIntent, ui.SimpleDecrementIntent, 或自定义类型
	handler func(), // 处理函数，不需要接收参数（闭包直接访问局部变量）
) {
	// 生成唯一键
	key := intentType.IntentType()

	// 检查是否已注册（包级 map，避免全局状态污染和不同组件间的冲突）
	if _, loaded := registeredHandlers.LoadOrStore(key, true); loaded {
		// 已注册，跳过
		return
	}

	// 注册全局 Intent 处理器
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
		// 调用 handler（闭包直接访问组件内的局部变量）
		handler()

		// 返回处理结果
		return intent.HandledResult()
	})
}

// OnWithContext 注册 Intent 处理器（推荐 API，无闭包风险）
//
// 这是推荐的 Intent 注册方式，handler 接收 *ActionContext 参数，
// 可以从中读取最新状态，避免闭包捕获旧值的问题。
//
// 使用示例：
//
//	func Counter() ui.VNode {
//		count, setCount, _ := ui.UseStateInt(0)
//
//		// ✅ 将状态保存到 GlobalStore 供 handler 读取
//		ctx := ui.GetCurrentContext()
//		if ctx != nil {
//			ctx.GlobalState["counter_count"] = count
//			ctx.GlobalState["counter_setCount"] = setCount
//		}
//
//		ui.OnWithContext(ui.SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
//			// ✅ 从 Context 读取最新值
//			setCountFn, _ := ctx.GlobalStore["counter_setCount"].(func(interface{}))
//			if setCountFn != nil {
//				setCountFn(func(c int) int { return c + 1 })
//			}
//		})
//
//		return ui.Text(fmt.Sprintf("Count: %d", count))
//	}
//
// 更简洁写法（使用类型安全的访问器）：
//
//	func Form() ui.VNode {
//		ctx := ui.GetCurrentContext()
//		ctx.SetState("username", username)  // 保存状态
//
//		ui.OnWithContext(SubmitIntent{}, func(ctx *intent.ActionContext) {
//			// ✅ 类型安全地读取状态
//			username := ctx.GetStringState("username", "")
//			// 处理提交逻辑...
//		})
//	}
func OnWithContext[T interface{ IntentType() string; StayPressed() bool }](
	intentType T,
	handler func(*intent.ActionContext),
) {
	// 生成唯一键
	key := intentType.IntentType()

	// 检查是否已注册
	if _, loaded := registeredHandlers.LoadOrStore(key, true); loaded {
		return
	}

	// 注册全局 Intent 处理器，传递 ActionContext 给 handler
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
