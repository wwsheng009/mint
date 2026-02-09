package inspector

import (
	"fmt"
	"os"
)

// DebugKeyEvent returns a version of HandleKeyEvent with extensive debugging
func (si *StandaloneInspector) DebugKeyEvent() func(key string, alt bool, ctrl bool, shift bool) bool {
	return func(key string, alt bool, ctrl bool, shift bool) bool {
		// 调试输出
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" || os.Getenv("TUI_DEBUG") == "true" {
			fmt.Fprintf(os.Stderr, "\n=== Inspector KeyEvent Debug ===\n")
			fmt.Fprintf(os.Stderr, "Key: '%s'\n", key)
			fmt.Fprintf(os.Stderr, "Alt: %v, Ctrl: %v, Shift: %v\n", alt, ctrl, shift)
			fmt.Fprintf(os.Stderr, "Inspector visible: %v\n", si.visible)
			fmt.Fprintf(os.Stderr, "Inspector enabled: %v\n", si.enabled)
			fmt.Fprintf(os.Stderr, "Current tab before: %v\n", si.activeTab)
		}

		// 调用原始 HandleKeyEvent
		result := si.HandleKeyEvent(key, alt, ctrl, shift)

		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" || os.Getenv("TUI_DEBUG") == "true" {
			fmt.Fprintf(os.Stderr, "HandleKeyEvent returned: %v\n", result)
			fmt.Fprintf(os.Stderr, "Current tab after: %v\n", si.activeTab)
			fmt.Fprintf(os.Stderr, "===========================\n")
		}

		return result
	}
}

// ExplainDigitKeyIssue 解释数字键问题
//
// 问题：数字键没有被检测到 / 按数字键没有反馈
//
// 可能的原因：
//
// 1. **事件根本没有到达 Inspector**
//    - framework 可能拦截了数字键
//    - 或者事件过滤器没有被正确设置
//    - 或者 Inspector 没有正确注册到应用
//
// 2. **HandleKeyEvent 被调用，但 activeTab 改变了，UI 没有更新**
//    - 这是渲染问题，不是事件处理问题
//    - 解决方案：让数字键事件返回 false（已实现）
//
// 3. **数字键被其他组件拦截**
//    - 例如 text input 组件可能会拦截数字键
//    - Inspector overlay 的层级可能有问题
//
// 诊断步骤：
//
// 1. 启用调试模式：
//    export TUI_INSPECTOR_VERBOSE=true
//    export TUI_DEBUG=true
//
// 2. 运行程序，按数字键
//    应该会看到调试输出：
//    === Inspector KeyEvent Debug ===
//    Key: '5'
//    Alt: false, Ctrl: false, Shift: false
//    Inspector visible: true
//    Inspector enabled: true
//    Current tab before: 0
//    HandleKeyEvent returned: false
//    Current tab after: 4
//    ============================
//
// 3. 如果没有看到输出，说明事件没有到达 Inspector
//
// 解决方案：
//
// 如果问题 1（事件没到达）：
//    - 检查 framework 是否正确设置了事件过滤器
//    - 检查 Inspector 是否正确注册
//
// 如果问题 2（事件到达但 UI 不更新）：
//    - 已经通过将返回值改为 false 解决
//    - 事件会传播，触发应用重新渲染
//
// 如果问题 3（被其他组件拦截）：
//    - 需要修改拦截数字键的组件
//    - 或者使用其他快捷键（如 Tab 键）
//
// 当前状态：
// - ✅ 已修改数字键返回值为 false
// - ✅ 添加了调试输出
// - ⏳ 等待用户测试确认
