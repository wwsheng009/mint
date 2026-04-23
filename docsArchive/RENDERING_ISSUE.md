package inspector

import (
	"fmt"
	"os"
)

// DebugKeyResponse 添加调试输出来追踪数字键响应
func (si *StandaloneInspector) DebugKeyResponse(key string, alt, ctrl, shift bool) bool {
	if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" || os.Getenv("TUI_DEBUG") == "true" {
		modifiers := ""
		if alt {
			modifiers += "Alt+"
		}
		if ctrl {
			modifiers += "Ctrl+"
		}
		if shift {
			modifiers += "Shift+"
		}

		fmt.Fprintf(os.Stderr, "[Inspector] Key pressed: %s%s\n", modifiers, key)
		fmt.Fprintf(os.Stderr, "[Inspector] Before: activeTab=%v\n", si.activeTab)
	}

	// 调用原始的 HandleKeyEvent
	handled := si.HandleKeyEvent(key, alt, ctrl, shift)

	if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" || os.Getenv("TUI_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] After: activeTab=%v, handled=%v\n", si.activeTab, handled)

		// 检查内容是否可以渲染
		content := si.RenderContent()
		if content != nil {
			fmt.Fprintf(os.Stderr, "[Inspector] RenderContent() returned non-nil VNode\n")
		} else {
			fmt.Fprintf(os.Stderr, "[Inspector] RenderContent() returned nil\n")
		}
	}

	return handled
}

// ExplainRenderingIssue 说明渲染问题的文档
//
// 问题：按数字键切换 tab 后，UI 没有立即更新
//
// 根本原因：
// Inspector 通过 render hook 系统注入到应用中。每次应用重新渲染时，
// hook 会调用 RenderContent() 来获取 Inspector 的最新 UI。
//
// 但是，当用户按数字键时：
// 1. HandleKeyEvent() 被调用，activeTab 改变 ✅
// 2. 事件被阻止（返回 true），不会传播到应用
// 3. 应用不知道需要重新渲染 ❌
// 4. Hook 不会被调用，UI 不会更新
//
// 解决方案：
// Inspector 应该在状态改变时通知应用重新渲染。
// 但是当前的 StandaloneInspector 没有这种能力。
//
// 当前的"触发重新渲染"方式：
// - 等待下一个事件（如键盘输入、鼠标移动等）
// - 或者按任意其他键触发事件循环
//
// 为什么移动窗口（Alt+HJKL）看起来立即有反馈：
// - 因为移动窗口改变了 Inspector 的位置（floatX, floatY）
// - 这些变化可能通过某种方式被检测到并触发渲染
// - 或者实际上也没有立即反馈，用户感知错误
//
// 为什么 Tab 键看起来有反馈：
// - Tab 键可能返回 false，让事件传播
// - 传播的事件触发了应用重新渲染
// - 然后hook 被调用，UI 更新
//
// 解决方案选项：
// 1. 数字键事件返回 false（让事件传播，触发重新渲染）
// 2. 添加一个"dirty"标志，让应用知道需要重新渲染
// 3. 让 Inspector 持续渲染或按需渲染
// 4. 添加一个显式的"请求渲染"机制
//
// 当前限制：
// - StandaloneInspector 没有访问应用渲染循环的能力
// - 它是一个独立组件，通过 hook 系统注入
// - 需要框架层的支持来实现"请求渲染"机制
